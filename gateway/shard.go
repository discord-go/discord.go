package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
)

// ShardDelay is the minimum time between starting consecutive shard connections,
// as required by Discord's gateway rate limits.
const ShardDelay = 5 * time.Second

// ShardID identifies a specific shard in a sharded gateway setup.
type ShardID struct {
	ShardID   int `json:"shard_id"`
	NumShards int `json:"num_shards"`
}

// String returns a human-readable representation of the ShardID.
func (s ShardID) String() string {
	return fmt.Sprintf("[%d/%d]", s.ShardID, s.NumShards)
}

// ToIdentifyShard returns the shard field for an Identify payload.
func (s ShardID) ToIdentifyShard() []int {
	return []int{s.ShardID, s.NumShards}
}

// ShardManager manages multiple gateway clients, one per shard.
type ShardManager struct {
	mu      sync.RWMutex
	token   string
	intents intents.Intent

	numShards int
	// shardIDs is nil when this manager owns every shard. A non-nil slice lets
	// a process own a selected subset while retaining the Discord-wide count.
	shardIDs []int
	shards   []*shardEntry

	// connFactory creates a Connection for a given shard.
	// Allows injection for testing.
	connFactory    func(shardID ShardID) (Connection, error)
	connURLFactory func(url string, shardID ShardID) (Connection, error)
	gatewayURL     string
	compressed     bool
	capabilities   Capability

	// shardDelay controls the delay between shard startups.
	// Defaults to ShardDelay; can be overridden for testing.
	shardDelay time.Duration
	// restartAttempts is the number of additional manager-level attempts after
	// Client.Start exits. Zero preserves fail-fast supervision.
	restartAttempts int
	restartBackoff  time.Duration

	// Dispatcher handles events for all shards.
	Dispatcher *Dispatcher
	cache      cache.Cache

	identifyTracker *IdentifyTracker
}

// shardEntry holds per-shard state.
type shardEntry struct {
	id     ShardID
	client *Client
	cancel context.CancelFunc
	done   chan struct{}
	err    error
}

// NewShardManager creates a new ShardManager for the given number of shards.
// If numShards is 0, Start will automatically determine the recommended number of shards.
func NewShardManager(token string, numShards int, gatewayIntents intents.Intent) *ShardManager {
	if numShards < 0 {
		numShards = 1
	}

	return &ShardManager{
		token:           token,
		intents:         gatewayIntents,
		numShards:       numShards,
		shards:          make([]*shardEntry, 0),
		shardDelay:      ShardDelay,
		Dispatcher:      NewDispatcher(),
		identifyTracker: NewIdentifyTracker(),
		cache:           cache.NewMemoryCache(),
	}
}

// SetConnectionFactory sets a custom connection factory for the shards.
func (sm *ShardManager) SetConnectionFactory(factory func(shardID ShardID) (Connection, error)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.connFactory = factory
}

// SetConnectionURLFactory configures a factory that receives Discord's resume
// URL when a shard reconnects.
func (sm *ShardManager) SetConnectionURLFactory(factory func(url string, shardID ShardID) (Connection, error)) {
	sm.mu.Lock()
	sm.connURLFactory = factory
	sm.mu.Unlock()
}

// SetGatewayURL sets the initial gateway URL used before a shard has a resume URL.
func (sm *ShardManager) SetGatewayURL(url string) {
	sm.mu.Lock()
	sm.gatewayURL = url
	sm.mu.Unlock()
}

// SetCompression enables zlib-stream compression on every shard connection.
func (sm *ShardManager) SetCompression(enabled bool) {
	sm.mu.Lock()
	sm.compressed = enabled
	sm.mu.Unlock()
}

// SetCapabilities sets the Identify capabilities bitfield sent by every
// shard, for example CapabilityChannelObfuscation.
func (sm *ShardManager) SetCapabilities(caps Capability) {
	sm.mu.Lock()
	sm.capabilities = caps
	sm.mu.Unlock()
}

// SetCache enables shared cache hydration for all shards. Passing nil keeps
// the default per-manager memory cache.
func (sm *ShardManager) SetCache(store cache.Cache) {
	sm.mu.Lock()
	if store != nil {
		sm.cache = store
	}
	sm.mu.Unlock()
}

// Broadcast sends a gateway payload to every active shard.
func (sm *ShardManager) Broadcast(ctx context.Context, payload GatewayPayload) error {
	sm.mu.RLock()
	clients := make([]*Client, 0, len(sm.shards))
	for _, shard := range sm.shards {
		clients = append(clients, shard.client)
	}
	sm.mu.RUnlock()
	for _, client := range clients {
		if err := client.Send(ctx, payload); err != nil {
			return err
		}
	}
	return nil
}

// JoinVoiceChannel forwards the request to the correct shard.
func (sm *ShardManager) JoinVoiceChannel(guildID, channelID snowflake.ID, mute, deaf bool) error {
	sm.mu.RLock()
	// Usually, guildID is used to calculate shard ID: (guildID >> 22) % num_shards
	// But ShardManager here might just be simple. If num_shards=1, it's Shard 0.
	numShards := sm.numShards
	sm.mu.RUnlock()

	shardIdx := CalculateShardID(guildID, numShards)
	client := sm.Shard(shardIdx)
	if client == nil {
		return fmt.Errorf("shard %d not found", shardIdx)
	}
	return client.JoinVoiceChannel(guildID, channelID, mute, deaf)
}

// NumShards returns the total number of shards.
func (sm *ShardManager) NumShards() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.numShards
}

// SetShardIDs configures the subset of Discord shard IDs owned by this
// manager. The configured shard count remains the total count passed to
// NewShardManager (or discovered when it is zero). This is intended for
// deployments that assign non-overlapping shard IDs to separate processes.
// It must be called before Start. A nil or empty slice restores ownership of
// every shard.
func (sm *ShardManager) SetShardIDs(ids []int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if len(sm.shards) > 0 {
		return fmt.Errorf("gateway: shard ownership cannot change after Start")
	}

	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id < 0 {
			return fmt.Errorf("gateway: shard ID must be non-negative: %d", id)
		}
		if sm.numShards > 0 && id >= sm.numShards {
			return fmt.Errorf("gateway: shard ID %d is outside total shard count %d", id, sm.numShards)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("gateway: duplicate shard ID %d", id)
		}
		seen[id] = struct{}{}
	}
	sm.shardIDs = append(sm.shardIDs[:0], ids...)
	if len(sm.shardIDs) == 0 {
		sm.shardIDs = nil
	}
	return nil
}

// SetRestartPolicy enables bounded manager-level restarts after a shard's
// Client.Start exits. attempts is the number of additional connection
// attempts; backoff is the initial delay and doubles after each failure, up
// to 30 seconds. Passing zero attempts disables manager-level restarts.
func (sm *ShardManager) SetRestartPolicy(attempts int, backoff time.Duration) error {
	if attempts < 0 || backoff < 0 || (attempts > 0 && backoff == 0) {
		return fmt.Errorf("gateway: restart attempts and backoff must be non-negative, with positive backoff when attempts are enabled")
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if len(sm.shards) > 0 {
		return fmt.Errorf("gateway: restart policy cannot change after Start")
	}
	sm.restartAttempts = attempts
	sm.restartBackoff = backoff
	return nil
}

type SessionStartLimit struct {
	Total          int `json:"total"`
	Remaining      int `json:"remaining"`
	ResetAfter     int `json:"reset_after"`
	MaxConcurrency int `json:"max_concurrency"`
}

type gatewayBotResponse struct {
	URL               string             `json:"url"`
	Shards            int                `json:"shards"`
	SessionStartLimit *SessionStartLimit `json:"session_start_limit"`
}

// Start connects and starts all shards sequentially with the required delay
// between each connection. It returns an error if any shard fails to start.
func (sm *ShardManager) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.connFactory == nil && sm.connURLFactory == nil {
		return fmt.Errorf("gateway: connection factory not set, call SetConnectionFactory before Start")
	}

	var maxConcurrency = 1

	if sm.numShards == 0 {
		req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/v10/gateway/bot", nil)
		if err != nil {
			return fmt.Errorf("gateway: failed to create request for gateway/bot: %w", err)
		}
		req.Header.Set("Authorization", "Bot "+sm.token)

		// Use a dedicated client with a timeout instead of http.DefaultClient,
		// which has no timeout and can block indefinitely if the caller's
		// context has no deadline.
		gatewayBotClient := &http.Client{Timeout: 30 * time.Second}
		resp, err := gatewayBotClient.Do(req)
		if err != nil {
			return fmt.Errorf("gateway: failed to fetch gateway/bot: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("gateway: unexpected status from gateway/bot: %d", resp.StatusCode)
		}

		var botResp gatewayBotResponse
		if err := json.NewDecoder(resp.Body).Decode(&botResp); err != nil {
			return fmt.Errorf("gateway: failed to decode gateway/bot response: %w", err)
		}

		sm.numShards = botResp.Shards
		if sm.numShards < 1 {
			sm.numShards = 1
		}
		if botResp.SessionStartLimit != nil && botResp.SessionStartLimit.MaxConcurrency > 0 {
			maxConcurrency = botResp.SessionStartLimit.MaxConcurrency
		}
	}

	ownedShardIDs := sm.shardIDs
	if ownedShardIDs == nil {
		ownedShardIDs = make([]int, sm.numShards)
		for i := range ownedShardIDs {
			ownedShardIDs[i] = i
		}
	}
	for _, shardID := range ownedShardIDs {
		if shardID >= sm.numShards {
			return fmt.Errorf("gateway: shard ID %d is outside total shard count %d", shardID, sm.numShards)
		}
	}

	// Re-allocate shards slice for the locally owned shards.
	if cap(sm.shards) < len(ownedShardIDs) {
		sm.shards = make([]*shardEntry, 0, len(ownedShardIDs))
	}

	// Group shards into buckets. Shards in the same bucket start concurrently.
	// Buckets start sequentially with ShardDelay between them.
	buckets := make([][]int, 0)
	for i, shardID := range ownedShardIDs {
		bucketIdx := i / maxConcurrency
		if bucketIdx >= len(buckets) {
			buckets = append(buckets, []int{})
		}
		buckets[bucketIdx] = append(buckets[bucketIdx], shardID)
	}

	for b, bucket := range buckets {
		// Respect context cancellation.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Apply delay between buckets (skip for the first bucket).
		if b > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sm.shardDelay):
			}
		}

		for _, shardIdx := range bucket {
			shardID := ShardID{ShardID: shardIdx, NumShards: sm.numShards}

			var conn Connection
			var err error
			if sm.connURLFactory != nil {
				conn, err = sm.connURLFactory(sm.gatewayURL, shardID)
			} else {
				conn, err = sm.connFactory(shardID)
			}
			if err != nil {
				return fmt.Errorf("gateway: failed to create connection for shard %s: %w", shardID, err)
			}

			connect := func(url string) (Connection, error) {
				if sm.connURLFactory != nil {
					if url == "" {
						url = sm.gatewayURL
					}
					return sm.connURLFactory(url, shardID)
				}
				return sm.connFactory(shardID)
			}
			client := sm.newShardClient(shardID, conn, connect)

			shardCtx, shardCancel := context.WithCancel(ctx)

			entry := &shardEntry{
				id:     shardID,
				client: client,
				cancel: shardCancel,
				done:   make(chan struct{}),
			}

			go sm.runShard(shardCtx, entry, client, connect)

			sm.shards = append(sm.shards, entry)
		}
	}

	return nil
}

func (sm *ShardManager) newShardClient(id ShardID, conn Connection, connect func(string) (Connection, error)) *Client {
	client := NewClient(conn, sm.Dispatcher)
	client.Session = NewSession()
	client.Cache = sm.cache
	client.Compressed = sm.compressed
	client.SetToken(sm.token)
	client.Intents = sm.intents
	client.Capabilities = sm.capabilities
	client.Shard = id.ToIdentifyShard()
	client.IdentifyTracker = sm.identifyTracker
	client.ConnFactory = connect
	return client
}

func (sm *ShardManager) runShard(ctx context.Context, entry *shardEntry, client *Client, connect func(string) (Connection, error)) {
	defer close(entry.done)
	var lastErr error
	var conn Connection
	for attempt := 0; ; attempt++ {
		lastErr = client.Start(ctx)
		if ctx.Err() != nil {
			entry.err = lastErr
			return
		}
		if attempt >= sm.restartAttempts {
			entry.err = lastErr
			return
		}

		delay := sm.restartBackoff
		for i := 0; i < attempt && delay < 30*time.Second; i++ {
			delay *= 2
		}
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			entry.err = ctx.Err()
			return
		case <-timer.C:
		}

		for retry := attempt + 1; ; retry++ {
			conn, lastErr = connect("")
			if lastErr == nil {
				if ctx.Err() != nil {
					entry.err = ctx.Err()
					return
				}
				client = sm.newShardClient(entry.id, conn, connect)
				sm.mu.Lock()
				if ctx.Err() != nil {
					sm.mu.Unlock()
					_ = conn.Close()
					entry.err = ctx.Err()
					return
				}
				entry.client = client
				sm.mu.Unlock()
				break
			}
			if retry >= sm.restartAttempts {
				entry.err = lastErr
				return
			}
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				entry.err = ctx.Err()
				return
			case <-timer.C:
			}
		}
	}
}

// Shutdown gracefully stops all shards.
func (sm *ShardManager) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	sm.mu.Lock()
	entries := append([]*shardEntry(nil), sm.shards...)
	clients := make([]*Client, 0, len(entries))
	for _, entry := range entries {
		entry.cancel()
		if entry.client != nil {
			clients = append(clients, entry.client)
		}
	}
	sm.mu.Unlock()

	var firstErr error

	for _, client := range clients {
		// Close the underlying connection to unblock Read().
		if err := client.Conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	for _, entry := range entries {
		// Wait for the shard to finish or context deadline.
		select {
		case <-entry.done:
		case <-ctx.Done():
			if firstErr == nil {
				firstErr = ctx.Err()
			}
		}
	}

	sm.mu.Lock()
	sm.shards = nil
	sm.mu.Unlock()
	return firstErr
}

// Wait blocks until a shard exits or ctx is cancelled. A shard exit while the
// manager context is still active is returned as an error, even when the
// shard's Client.Start returned nil, because the manager has lost coverage for
// that shard. Call Shutdown after Wait returns an error or the context ends.
func (sm *ShardManager) Wait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	sm.mu.RLock()
	entries := append([]*shardEntry(nil), sm.shards...)
	sm.mu.RUnlock()
	if len(entries) == 0 {
		return fmt.Errorf("gateway: no shards to monitor")
	}

	for {
		for _, entry := range entries {
			select {
			case <-entry.done:
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if entry.err != nil {
					return fmt.Errorf("gateway: shard %s stopped: %w", entry.id, entry.err)
				}
				return fmt.Errorf("gateway: shard %s stopped unexpectedly", entry.id)
			default:
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}

// Shard returns the Client for a specific shard index. Returns nil if not found.
func (sm *ShardManager) Shard(index int) *Client {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, shard := range sm.shards {
		if shard.client == nil {
			continue
		}
		if shard.id.ShardID == index {
			return shard.client
		}
	}
	return nil
}

// CalculateShardID calculates the shard ID for a given guild ID and total number of shards.
func CalculateShardID(guildID snowflake.ID, numShards int) int {
	if numShards <= 0 {
		numShards = 1
	}
	return int((uint64(guildID) >> 22) % uint64(numShards))
}

// CalculateShards calculates the number of shards to use based on the recommended shards.
// It allows for large bot sharding by returning a multiple of the Discord-assigned number.
func CalculateShards(recommended, multiple int) int {
	if multiple <= 0 {
		multiple = 1
	}
	if recommended <= 0 {
		recommended = 1
	}
	return recommended * multiple
}

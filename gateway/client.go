package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/internal/backoff"
	"github.com/discord-go/discord.go/internal/compression"
	"github.com/discord-go/discord.go/snowflake"
)

// Connection defines the interface for a websocket-like connection.
type Connection interface {
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}

// Client represents a Discord gateway client.
type Client struct {
	Conn        Connection
	Dispatcher  *Dispatcher
	Session     *Session
	Heartbeater *Heartbeater

	token   string
	Intents intents.Intent
	Shard   []int
	Cache   cache.Cache

	ConnFactory func(url string) (Connection, error)
	GatewayURL  string
	Compressed  bool
	compression *compression.Stream

	IdentifyTracker *IdentifyTracker

	sendMu      sync.Mutex
	sendTimes   [120]time.Time
	sendTimeIdx int
}

// SetToken sets the bot token used for gateway identification. The token
// is stored unexported and is only accessible via the internal gateway
// path. Treat the token as a secret: do not log it or commit it to
// version control.
func (c *Client) SetToken(token string) {
	c.token = token
}

// NewClient creates a new Client.
func NewClient(conn Connection, dispatcher *Dispatcher) *Client {
	return &Client{
		Conn:       newSynchronizedConnection(conn),
		Dispatcher: dispatcher,
		Cache:      cache.NewMemoryCache(),
	}
}

type synchronizedConnection struct {
	conn    Connection
	writeMu sync.Mutex
}

func newSynchronizedConnection(conn Connection) Connection {
	if conn == nil {
		return nil
	}
	return &synchronizedConnection{conn: conn}
}

func (c *synchronizedConnection) Read() ([]byte, error) { return c.conn.Read() }
func (c *synchronizedConnection) Write(data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.Write(data)
}
func (c *synchronizedConnection) Close() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.Close()
}

// Start starts reading from the connection and dispatches events.
func (c *Client) Start(ctx context.Context) error {
	var attempt int
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := c.readLoop(ctx)

		if err == context.Canceled || err == context.DeadlineExceeded {
			return err
		}

		if errors.Is(err, ErrFatalClose) {
			return err
		}

		// Close existing connection if any
		if c.Conn != nil {
			_ = c.Conn.Close()
		}

		if c.ConnFactory == nil {
			return err
		}

		// Calculate backoff
		delay := backoff.Calculate(attempt, time.Second, 30*time.Second)
		attempt++

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		// Reconnect
		url := c.GatewayURL
		if c.Session != nil && c.Session.CanResume() && !errors.Is(err, ErrInvalidSession) {
			url = c.Session.ResumeURL()
		} else if errors.Is(err, ErrInvalidSession) && c.Session != nil {
			c.Session.Reset()
		}

		if c.ConnFactory != nil {
			conn, connErr := c.ConnFactory(url)
			if connErr != nil {
				continue
			}
			c.Conn = newSynchronizedConnection(conn)
			c.compression = nil
		}

		// Reset attempt on successful reconnect start
		attempt = 0
	}
}

// RequestGuildMembersData represents the data for a Request Guild Members payload.
type RequestGuildMembersData struct {
	GuildID   snowflake.ID   `json:"guild_id,string"`
	Query     *string        `json:"query,omitempty"`
	Limit     int            `json:"limit"`
	Presences bool           `json:"presences,omitempty"`
	UserIDs   []snowflake.ID `json:"user_ids,omitempty"`
	Nonce     string         `json:"nonce,omitempty"`
}

// RequestGuildMembers sends a request to fetch guild members.
func (c *Client) RequestGuildMembers(data RequestGuildMembersData) error {
	return c.RequestGuildMembersContext(context.Background(), data)
}

// RequestGuildMembersContext sends a request to fetch guild members with the
// given context, allowing cancellation of the underlying gateway send.
func (c *Client) RequestGuildMembersContext(ctx context.Context, data RequestGuildMembersData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Send(ctx, GatewayPayload{Op: OpcodeRequestGuildMembers, Data: b})
}

// VoiceStateUpdateData represents the payload to join or leave a voice channel.
type VoiceStateUpdateData struct {
	GuildID   snowflake.ID  `json:"guild_id,string"`
	ChannelID *snowflake.ID `json:"channel_id,string"`
	SelfMute  bool          `json:"self_mute"`
	SelfDeaf  bool          `json:"self_deaf"`
}

// JoinVoiceChannel sends an Opcode 4 VoiceStateUpdate.
func (c *Client) JoinVoiceChannel(guildID, channelID snowflake.ID, mute, deaf bool) error {
	return c.JoinVoiceChannelContext(context.Background(), guildID, channelID, mute, deaf)
}

// JoinVoiceChannelContext sends an Opcode 4 VoiceStateUpdate with the given
// context, allowing cancellation of the underlying gateway send.
func (c *Client) JoinVoiceChannelContext(ctx context.Context, guildID, channelID snowflake.ID, mute, deaf bool) error {
	payloadData := map[string]interface{}{
		"guild_id":  guildID.String(),
		"self_mute": mute,
		"self_deaf": deaf,
	}
	if channelID != 0 {
		payloadData["channel_id"] = channelID.String()
	} else {
		payloadData["channel_id"] = nil
	}
	b, err := json.Marshal(payloadData)
	if err != nil {
		return err
	}
	return c.Send(ctx, GatewayPayload{Op: OpcodeVoiceStateUpdate, Data: b})
}

// Send sends a GatewayPayload to the gateway, enforcing payload size (4096 bytes) and rate limit (120/60s).
func (c *Client) Send(ctx context.Context, payload GatewayPayload) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.Conn == nil {
		return errors.New("gateway: connection is nil")
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if len(b) > 4096 {
		return fmt.Errorf("gateway: payload exceeds 4096 bytes limit")
	}

	for {
		c.sendMu.Lock()
		now := time.Now()
		oldest := c.sendTimes[c.sendTimeIdx]
		var waitDuration time.Duration
		if !oldest.IsZero() && now.Sub(oldest) < time.Minute {
			waitDuration = time.Minute - now.Sub(oldest)
		}
		if waitDuration <= 0 {
			c.sendTimes[c.sendTimeIdx] = now
			c.sendTimeIdx = (c.sendTimeIdx + 1) % 120
			c.sendMu.Unlock()
			break
		}
		c.sendMu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
		}
	}

	return c.Conn.Write(b)
}

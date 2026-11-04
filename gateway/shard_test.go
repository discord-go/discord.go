package gateway

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/discord-go/discord.go/intents"
	"github.com/gorilla/websocket"
)

type shardMockConnection struct {
	closed chan struct{}
}

type shardFatalConnection struct {
	fatal  bool
	closed chan struct{}
	once   sync.Once
}

func (c *shardFatalConnection) Read() ([]byte, error) {
	if c.fatal {
		return nil, &websocket.CloseError{Code: 4004, Text: "authentication failed"}
	}
	<-c.closed
	return nil, fmt.Errorf("connection closed")
}
func (*shardFatalConnection) Write([]byte) error { return nil }
func (c *shardFatalConnection) Close() error {
	c.once.Do(func() { close(c.closed) })
	return nil
}

func (c *shardMockConnection) Read() ([]byte, error) {
	// Block until closed
	<-c.closed
	return nil, fmt.Errorf("connection closed")
}

func (c *shardMockConnection) Write(data []byte) error {
	return nil
}

func (c *shardMockConnection) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
	return nil
}

func TestShardManager_StartAndShutdown(t *testing.T) {
	sm := NewShardManager("fake-token", 3, intents.Guilds)
	// Override delay for testing to speed things up
	sm.shardDelay = 10 * time.Millisecond

	connFactoryCount := 0
	sm.SetConnectionFactory(func(shardID ShardID) (Connection, error) {
		connFactoryCount++
		return &shardMockConnection{closed: make(chan struct{})}, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sm.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if connFactoryCount != 3 {
		t.Fatalf("Expected connection factory to be called 3 times, got %d", connFactoryCount)
	}

	if sm.NumShards() != 3 {
		t.Fatalf("Expected 3 shards, got %d", sm.NumShards())
	}

	if sm.Shard(0) == nil || sm.Shard(2) == nil || sm.Shard(3) != nil {
		t.Fatalf("Shard() returned unexpected results")
	}

	err = sm.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if len(sm.shards) != 0 {
		t.Fatalf("Expected shards to be cleared after shutdown")
	}
}

func TestShardManager_StartCancellation(t *testing.T) {
	sm := NewShardManager("fake-token", 5, intents.Guilds)
	sm.shardDelay = 50 * time.Millisecond

	sm.SetConnectionFactory(func(shardID ShardID) (Connection, error) {
		return &shardMockConnection{closed: make(chan struct{})}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()

	err := sm.Start(ctx)
	if err == nil {
		t.Fatal("Expected error on Start with cancellation, got nil")
	}

	if err != context.DeadlineExceeded {
		t.Fatalf("Expected DeadlineExceeded, got %v", err)
	}
}

func TestShardManager_NoConnectionFactory(t *testing.T) {
	sm := NewShardManager("fake-token", 1, intents.Guilds)
	err := sm.Start(context.Background())
	if err == nil {
		t.Fatal("Expected error when connection factory is not set")
	}
}

func TestShardManager_ConnectionFactoryError(t *testing.T) {
	sm := NewShardManager("fake-token", 1, intents.Guilds)
	sm.SetConnectionFactory(func(shardID ShardID) (Connection, error) {
		return nil, fmt.Errorf("factory error")
	})

	err := sm.Start(context.Background())
	if err == nil {
		t.Fatal("Expected error when connection factory returns error")
	}
}

func TestShardManager_WaitReturnsShardFailure(t *testing.T) {
	wantErr := fmt.Errorf("fatal gateway close")
	done := make(chan struct{})
	entry := &shardEntry{id: ShardID{ShardID: 2, NumShards: 4}, done: done, err: wantErr}
	sm := NewShardManager("fake-token", 4, intents.Guilds)
	sm.shards = []*shardEntry{entry}
	close(done)

	err := sm.Wait(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Wait() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestShardManager_WaitDetectsFatalShardAndShutdownClosesPeers(t *testing.T) {
	sm := NewShardManager("fake-token", 2, intents.Guilds)
	sm.shardDelay = 0
	var mu sync.Mutex
	var conns []*shardFatalConnection
	sm.SetConnectionFactory(func(ShardID) (Connection, error) {
		mu.Lock()
		defer mu.Unlock()
		conn := &shardFatalConnection{fatal: len(conns) == 0, closed: make(chan struct{})}
		conns = append(conns, conn)
		return conn, nil
	})

	if err := sm.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := sm.Wait(ctx)
	if !errors.Is(err, ErrFatalClose) {
		t.Fatalf("Wait error = %v, want ErrFatalClose", err)
	}
	if err := sm.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(conns) != 2 {
		t.Fatalf("created %d connections, want 2", len(conns))
	}
	for i, conn := range conns {
		select {
		case <-conn.closed:
		default:
			t.Errorf("connection %d was not closed during shutdown", i)
		}
	}
}

func TestShardManager_SetCapabilities(t *testing.T) {
	sm := NewShardManager("fake-token", 2, intents.Guilds)
	sm.shardDelay = time.Millisecond
	sm.SetConnectionFactory(func(shardID ShardID) (Connection, error) {
		return &shardMockConnection{closed: make(chan struct{})}, nil
	})
	sm.SetCapabilities(CapabilityChannelObfuscation)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sm.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = sm.Shutdown(context.Background()) }()

	client := sm.Shard(0)
	if client == nil {
		t.Fatal("expected shard 0")
	}
	if client.Capabilities != CapabilityChannelObfuscation {
		t.Errorf("shard capabilities = %d, want %d", client.Capabilities, CapabilityChannelObfuscation)
	}
}

func TestShardManager_StartOnlyOwnedShardIDs(t *testing.T) {
	sm := NewShardManager("fake-token", 6, intents.Guilds)
	sm.shardDelay = 0
	if err := sm.SetShardIDs([]int{1, 4}); err != nil {
		t.Fatalf("SetShardIDs: %v", err)
	}
	var started []int
	sm.SetConnectionFactory(func(id ShardID) (Connection, error) {
		started = append(started, id.ShardID)
		return &shardMockConnection{closed: make(chan struct{})}, nil
	})
	if err := sm.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() { _ = sm.Shutdown(context.Background()) }()
	if sm.NumShards() != 6 {
		t.Fatalf("NumShards() = %d, want total count 6", sm.NumShards())
	}
	if len(started) != 2 || started[0] != 1 || started[1] != 4 {
		t.Fatalf("started shard IDs = %v, want [1 4]", started)
	}
	if sm.Shard(1) == nil || sm.Shard(4) == nil || sm.Shard(0) != nil {
		t.Fatal("Shard() did not reflect locally owned shard IDs")
	}
}

func TestShardManager_SetShardIDsValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		ids  []int
	}{
		{name: "negative", ids: []int{-1}},
		{name: "out of range", ids: []int{3}},
		{name: "duplicate", ids: []int{1, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewShardManager("fake-token", 3, intents.Guilds)
			if err := sm.SetShardIDs(tc.ids); err == nil {
				t.Fatalf("SetShardIDs(%v) succeeded", tc.ids)
			}
		})
	}
}

func TestShardManager_RestartPolicyRestartsExitedShard(t *testing.T) {
	sm := NewShardManager("fake-token", 1, intents.Guilds)
	sm.shardDelay = 0
	if err := sm.SetRestartPolicy(1, time.Millisecond); err != nil {
		t.Fatalf("SetRestartPolicy: %v", err)
	}
	secondConnection := make(chan struct{})
	var mu sync.Mutex
	calls := 0
	sm.SetConnectionFactory(func(ShardID) (Connection, error) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		if calls == 1 {
			return &shardFatalConnection{fatal: true, closed: make(chan struct{})}, nil
		}
		if calls == 2 {
			close(secondConnection)
		}
		return &shardMockConnection{closed: make(chan struct{})}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := sm.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	select {
	case <-secondConnection:
	case <-time.After(time.Second):
		t.Fatal("manager did not create a replacement connection")
	}
	if sm.Shard(0) == nil {
		t.Fatal("replacement shard client is not available")
	}
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer waitCancel()
	if err := sm.Wait(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Wait() = %v while replacement is running", err)
	}
	if err := sm.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestShardManager_RestartPolicyValidation(t *testing.T) {
	sm := NewShardManager("fake-token", 1, intents.Guilds)
	if err := sm.SetRestartPolicy(1, 0); err == nil {
		t.Fatal("restart policy accepted zero backoff")
	}
	if err := sm.SetRestartPolicy(-1, time.Second); err == nil {
		t.Fatal("restart policy accepted a negative retry count")
	}
}

package gateway

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/discord-go/discord.go/intents"
)

type shardMockConnection struct {
	closed chan struct{}
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

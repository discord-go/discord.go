package bot

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/intents"
	"github.com/gorilla/websocket"
)

type shardLifecycleConn struct {
	fatal  bool
	closed chan struct{}
	once   sync.Once
}

func (c *shardLifecycleConn) Read() ([]byte, error) {
	if c.fatal {
		return nil, &websocket.CloseError{Code: 4004, Text: "authentication failed"}
	}
	<-c.closed
	return nil, errors.New("connection closed")
}
func (*shardLifecycleConn) Write([]byte) error { return nil }
func (c *shardLifecycleConn) Close() error {
	c.once.Do(func() { close(c.closed) })
	return nil
}

func TestRunShardsPropagatesFatalShardAndCleansConnection(t *testing.T) {
	manager := gateway.NewShardManager("token", 1, intents.Guilds)
	var mu sync.Mutex
	var connections []*shardLifecycleConn
	manager.SetConnectionFactory(func(gateway.ShardID) (gateway.Connection, error) {
		mu.Lock()
		defer mu.Unlock()
		conn := &shardLifecycleConn{fatal: len(connections) == 0, closed: make(chan struct{})}
		connections = append(connections, conn)
		return conn, nil
	})

	b := New("MTIz.NjQ1.abc123")
	done := make(chan struct{})
	go b.runShards(manager, context.Background(), done)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runShards did not terminate after a fatal shard exit")
	}

	err := b.lastRunError()
	if err == nil || !strings.Contains(err.Error(), "shard [0/1] stopped") {
		t.Fatalf("run error = %v, want shard [0/1] failure", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(connections) != 1 {
		t.Fatalf("created %d shard connections, want 1", len(connections))
	}
	select {
	case <-connections[0].closed:
	default:
		t.Fatal("failed shard connection was not closed during cleanup")
	}
}

package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/discord-go/discord.go/intents"
)

type mockConnection struct {
	readCount int
	data      []byte
	err       error
	closeErr  error
}

func (m *mockConnection) Read() ([]byte, error) {
	if m.readCount > 0 {
		m.readCount--
		return m.data, m.err
	}
	return nil, errors.New("read error")
}

func (m *mockConnection) Write([]byte) error {
	return nil
}

func (m *mockConnection) Close() error {
	return m.closeErr
}

func TestClient_Start_ContextCancelled(t *testing.T) {
	d := NewDispatcher()
	c := NewClient(&mockConnection{readCount: 1, data: []byte("test")}, d)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.Start(ctx)
	if err != context.Canceled {
		t.Errorf("expected context canceled, got %v", err)
	}
}

func TestClient_Start_ReadError(t *testing.T) {
	d := NewDispatcher()
	conn := &mockConnection{readCount: 0}
	c := NewClient(conn, d)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := c.Start(ctx)
	if err.Error() != "read error" {
		t.Errorf("expected read error, got %v", err)
	}
}

func TestClient_Start_Dispatch(t *testing.T) {
	d := NewDispatcher()
	var called atomic.Bool
	d.AddHandler(func(e []byte) {
		called.Store(true)
	})

	conn := &mockConnection{readCount: 1, data: []byte("test"), err: nil}
	c := NewClient(conn, d)

	go func() {
		_ = c.Start(context.Background())
	}()

	time.Sleep(50 * time.Millisecond)
	if !called.Load() {
		t.Errorf("expected dispatcher to be called")
	}
}

type captureConnection struct {
	writes [][]byte
}

func (c *captureConnection) Read() ([]byte, error) { return nil, errors.New("no reads") }
func (c *captureConnection) Write(data []byte) error {
	c.writes = append(c.writes, append([]byte(nil), data...))
	return nil
}
func (c *captureConnection) Close() error { return nil }

// TestSendIdentify_IncludesCapabilities verifies the Client.Capabilities
// bitfield reaches the Identify payload on the wire, which is how the
// temporary CHANNEL_OBFUSCATION (1 << 15) testing opt-in is requested.
func TestSendIdentify_IncludesCapabilities(t *testing.T) {
	conn := &captureConnection{}
	c := NewClient(conn, NewDispatcher())
	c.SetToken("token")
	c.Intents = intents.Guilds
	c.Capabilities = CapabilityChannelObfuscation

	if err := c.sendIdentify(); err != nil {
		t.Fatalf("sendIdentify: %v", err)
	}
	if len(conn.writes) != 1 {
		t.Fatalf("expected 1 write, got %d", len(conn.writes))
	}

	var payload GatewayPayload
	if err := json.Unmarshal(conn.writes[0], &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Op != OpcodeIdentify {
		t.Errorf("op = %d, want %d", payload.Op, OpcodeIdentify)
	}
	var identify Identify
	if err := json.Unmarshal(payload.Data, &identify); err != nil {
		t.Fatalf("unmarshal identify: %v", err)
	}
	if identify.Capabilities != CapabilityChannelObfuscation {
		t.Errorf("capabilities = %d, want %d", identify.Capabilities, CapabilityChannelObfuscation)
	}
}

package gateway

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
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

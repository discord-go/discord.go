package gateway

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

type heartbeatMockConnection struct {
	mu          sync.Mutex
	writeCount  int
	lastPayload []byte
	closed      bool
	readChan    chan []byte
	errChan     chan error
}

func newMockConnection() *heartbeatMockConnection {
	return &heartbeatMockConnection{
		readChan: make(chan []byte, 10),
		errChan:  make(chan error, 10),
	}
}

func (m *heartbeatMockConnection) Read() ([]byte, error) {
	select {
	case data := <-m.readChan:
		return data, nil
	case err := <-m.errChan:
		return nil, err
	}
}

func (m *heartbeatMockConnection) Write(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeCount++
	m.lastPayload = append([]byte(nil), data...)
	return nil
}

func (m *heartbeatMockConnection) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *heartbeatMockConnection) getWriteCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.writeCount
}

func (m *heartbeatMockConnection) getLastPayload() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastPayload
}

func TestHeartbeater_TimingAndACK(t *testing.T) {
	conn := newMockConnection()
	interval := 10 * time.Millisecond
	h := NewHeartbeater(conn, interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- h.Run(ctx)
	}()

	time.Sleep(15 * time.Millisecond)

	if conn.getWriteCount() != 1 {
		t.Fatalf("Expected 1 heartbeat sent, got %d", conn.getWriteCount())
	}

	h.AckReceived()

	time.Sleep(10 * time.Millisecond)

	if conn.getWriteCount() != 2 {
		t.Fatalf("Expected 2 heartbeats sent, got %d", conn.getWriteCount())
	}

	h.Stop()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Expected nil error on Stop(), got %v", err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Heartbeater did not stop in time")
	}
}

func TestHeartbeater_ZombieDetection(t *testing.T) {
	conn := newMockConnection()
	interval := 10 * time.Millisecond
	h := NewHeartbeater(conn, interval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- h.Run(ctx)
	}()

	time.Sleep(30 * time.Millisecond)

	select {
	case err := <-done:
		if err != ErrZombieConnection {
			t.Fatalf("Expected ErrZombieConnection, got %v", err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Heartbeater did not return ErrZombieConnection in time")
	}
}

func TestHeartbeater_SequenceTracking(t *testing.T) {
	conn := newMockConnection()
	interval := 10 * time.Millisecond
	h := NewHeartbeater(conn, interval)

	h.UpdateSequence(42)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- h.Run(ctx)
	}()

	time.Sleep(15 * time.Millisecond)

	payload := conn.getLastPayload()
	if payload == nil {
		t.Fatal("No heartbeat payload sent")
	}

	var p heartbeatPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if p.D == nil || *p.D != 42 {
		t.Fatalf("Expected sequence 42 in heartbeat, got %v", p.D)
	}

	h.Stop()
	<-done
}

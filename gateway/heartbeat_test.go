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
	onWrite     func()
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
	m.writeCount++
	m.lastPayload = append([]byte(nil), data...)
	m.mu.Unlock()
	if m.onWrite != nil {
		m.onWrite()
	}
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
	return append([]byte(nil), m.lastPayload...)
}

func TestHeartbeater_TimingAndACK(t *testing.T) {
	conn := newMockConnection()
	interval := 10 * time.Millisecond
	h := NewHeartbeater(conn, interval)
	heartbeats := make(chan struct{}, 2)
	conn.onWrite = func() {
		h.AckReceived()
		select {
		case heartbeats <- struct{}{}:
		default:
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- h.Run(ctx)
	}()

	for expected := 1; expected <= 2; expected++ {
		select {
		case <-heartbeats:
		case err := <-done:
			t.Fatalf("heartbeater stopped before heartbeat %d: %v", expected, err)
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for heartbeat %d", expected)
		}
	}

	h.Stop()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Expected nil error on Stop(), got %v", err)
		}
	case <-time.After(time.Second):
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

	select {
	case err := <-done:
		if err != ErrZombieConnection {
			t.Fatalf("Expected ErrZombieConnection, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Heartbeater did not return ErrZombieConnection in time")
	}
}

func TestHeartbeater_SequenceTracking(t *testing.T) {
	conn := newMockConnection()
	interval := 10 * time.Millisecond
	h := NewHeartbeater(conn, interval)
	firstHeartbeat := make(chan struct{}, 1)
	conn.onWrite = func() {
		h.AckReceived()
		select {
		case firstHeartbeat <- struct{}{}:
		default:
		}
	}

	h.UpdateSequence(42)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- h.Run(ctx)
	}()

	select {
	case <-firstHeartbeat:
	case err := <-done:
		t.Fatalf("heartbeater stopped before sending: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for heartbeat")
	}

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

func TestHeartbeaterStopBeforeRunIsPersistentAndIdempotent(t *testing.T) {
	h := NewHeartbeater(newMockConnection(), time.Hour)
	h.Stop()
	h.Stop()
	done := make(chan error, 1)
	go func() { done <- h.Run(context.Background()) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run ignored an earlier Stop")
	}
}

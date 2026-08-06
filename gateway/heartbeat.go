package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"
)

// ErrZombieConnection is returned when a heartbeat ACK is not received
// before the next heartbeat is due, indicating a dead connection.
var ErrZombieConnection = errors.New("gateway: zombie connection detected, no heartbeat ACK received")

// heartbeatPayload is the JSON payload sent for a heartbeat (opcode 1).
type heartbeatPayload struct {
	Op Opcode `json:"op"`
	D  *int64 `json:"d"`
}

// Heartbeater manages the heartbeat loop for a gateway connection.
// It periodically sends heartbeat payloads and monitors for ACK responses
// to detect zombie (dead) connections.
type Heartbeater struct {
	interval    time.Duration
	sequence    atomic.Int64
	conn        Connection
	ackReceived atomic.Bool
	lastSent    atomic.Int64
	lastPing    atomic.Int64
	stop        chan struct{}
	// nowFunc allows overriding time.Now for testing.
	nowFunc func() time.Time
	// newTickerFunc allows overriding time.NewTicker for testing.
	newTickerFunc func(d time.Duration) *time.Ticker
}

// NewHeartbeater creates a new Heartbeater for the given connection and interval.
func NewHeartbeater(conn Connection, interval time.Duration) *Heartbeater {
	h := &Heartbeater{
		interval:      interval,
		conn:          conn,
		stop:          make(chan struct{}),
		nowFunc:       time.Now,
		newTickerFunc: time.NewTicker,
	}
	// Mark initial ACK as received so first heartbeat doesn't trigger zombie detection.
	h.ackReceived.Store(true)
	// Initialize sequence to -1 to indicate no sequence received yet.
	h.sequence.Store(-1)
	return h
}

// Run starts the heartbeat loop. It blocks until the context is cancelled,
// Stop is called, or a zombie connection is detected.
func (h *Heartbeater) Run(ctx context.Context) error {
	ticker := h.newTickerFunc(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-h.stop:
			return nil
		case <-ticker.C:
			// Check if we received an ACK for the last heartbeat.
			if !h.ackReceived.Load() {
				return ErrZombieConnection
			}

			// Send heartbeat.
			h.ackReceived.Store(false)
			if err := h.sendHeartbeat(); err != nil {
				return err
			}
		}
	}
}

// sendHeartbeat sends a heartbeat payload with the current sequence number.
func (h *Heartbeater) sendHeartbeat() error {
	seq := h.sequence.Load()

	payload := heartbeatPayload{
		Op: OpcodeHeartbeat,
	}

	// If sequence is -1 (no events received yet), send null.
	if seq >= 0 {
		payload.D = &seq
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	h.lastSent.Store(time.Now().UnixNano())
	if err := h.conn.Write(data); err != nil {
		return err
	}
	return nil
}

// UpdateSequence updates the last known sequence number from dispatch events.
func (h *Heartbeater) UpdateSequence(seq int64) {
	h.sequence.Store(seq)
}

// AckReceived marks that a heartbeat ACK was received from the gateway.
func (h *Heartbeater) AckReceived() {
	h.ackReceived.Store(true)
	if sent := h.lastSent.Load(); sent > 0 {
		h.lastPing.Store(time.Now().UnixNano() - sent)
	}
}

// Ping returns the most recent heartbeat round-trip time.
func (h *Heartbeater) Ping() time.Duration {
	return time.Duration(h.lastPing.Load())
}

// Stop signals the heartbeat loop to stop gracefully.
func (h *Heartbeater) Stop() {
	select {
	case h.stop <- struct{}{}:
	default:
	}
}

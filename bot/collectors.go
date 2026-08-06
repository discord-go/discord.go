package bot

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	// ErrCollectorClosed indicates that a collector was cancelled before a
	// matching event arrived.
	ErrCollectorClosed = errors.New("bot: interaction collector closed")
)

// InteractionFilter decides whether an interaction should be collected.
type InteractionFilter func(*InteractionContext) bool

// MessageFilter decides whether a message should be collected.
type MessageFilter func(*MessageContext) bool

type interactionCollector struct {
	filter InteractionFilter
	result chan *InteractionContext
}

type messageCollector struct {
	filter MessageFilter
	result chan *MessageContext
}

// AwaitInteraction waits for one interaction matching filter or context
// cancellation. It is useful for button, select, modal, and wizard flows.
func (b *Bot) AwaitInteraction(ctx context.Context, filter InteractionFilter) (*InteractionContext, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	result := make(chan *InteractionContext, 1)
	id := b.subscriptions.Add(1)
	b.collectorMu.Lock()
	if b.interactionCollectors == nil {
		b.interactionCollectors = make(map[uint64]interactionCollector)
	}
	b.interactionCollectors[id] = interactionCollector{filter: filter, result: result}
	b.collectorMu.Unlock()
	defer func() {
		b.collectorMu.Lock()
		delete(b.interactionCollectors, id)
		b.collectorMu.Unlock()
	}()
	select {
	case interaction := <-result:
		return interaction, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// AwaitMessage waits for one message matching filter or context cancellation.
func (b *Bot) AwaitMessage(ctx context.Context, filter MessageFilter) (*MessageContext, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	result := make(chan *MessageContext, 1)
	id := b.subscriptions.Add(1)
	b.collectorMu.Lock()
	if b.messageCollectors == nil {
		b.messageCollectors = make(map[uint64]messageCollector)
	}
	b.messageCollectors[id] = messageCollector{filter: filter, result: result}
	b.collectorMu.Unlock()
	defer func() {
		b.collectorMu.Lock()
		delete(b.messageCollectors, id)
		b.collectorMu.Unlock()
	}()
	select {
	case message := <-result:
		return message, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (b *Bot) publishInteraction(interaction *InteractionContext) {
	b.collectorMu.Lock()
	for id, collector := range b.interactionCollectors {
		if collector.filter != nil && !collector.filter(interaction) {
			continue
		}
		select {
		case collector.result <- interaction:
			delete(b.interactionCollectors, id)
		default:
		}
	}
	b.collectorMu.Unlock()
}

func (b *Bot) publishMessage(message *MessageContext) {
	b.collectorMu.Lock()
	for id, collector := range b.messageCollectors {
		if collector.filter != nil && !collector.filter(message) {
			continue
		}
		select {
		case collector.result <- message:
			delete(b.messageCollectors, id)
		default:
		}
	}
	b.collectorMu.Unlock()
}

// Every starts a lifecycle-managed background job. The returned function
// cancels the job and is safe to call repeatedly.
func (b *Bot) Every(ctx context.Context, interval time.Duration, job func(context.Context)) func() {
	if interval <= 0 || job == nil {
		return func() {}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	b.stateMu.RLock()
	runCtx := b.runCtx
	b.stateMu.RUnlock()
	if runCtx == nil {
		runCtx = context.Background()
	}
	jobCtx, cancel := context.WithCancel(runCtx)
	id := b.subscriptions.Add(1)
	b.jobsMu.Lock()
	if b.jobs == nil {
		b.jobs = make(map[uint64]context.CancelFunc)
	}
	b.jobs[id] = func() {
		cancel()
	}
	b.jobsMu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer func() {
			b.jobsMu.Lock()
			delete(b.jobs, id)
			b.jobsMu.Unlock()
		}()
		for {
			select {
			case <-jobCtx.Done():
				return
			case <-ticker.C:
				func() {
					defer func() {
						if value := recover(); value != nil {
							b.reportError(&HandlerPanicError{Event: "scheduled job", Value: value})
						}
					}()
					job(jobCtx)
				}()
			}
		}
	}()
	var once atomic.Bool
	return func() {
		if once.CompareAndSwap(false, true) {
			b.jobsMu.Lock()
			cancelJob := b.jobs[id]
			delete(b.jobs, id)
			b.jobsMu.Unlock()
			if cancelJob != nil {
				cancelJob()
			}
		}
	}
}

func (b *Bot) cancelJobs() {
	b.jobsMu.Lock()
	jobs := make([]context.CancelFunc, 0, len(b.jobs))
	for id, cancel := range b.jobs {
		jobs = append(jobs, cancel)
		delete(b.jobs, id)
	}
	b.jobsMu.Unlock()
	for _, cancel := range jobs {
		cancel()
	}
}

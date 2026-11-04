package bot

import (
	"context"
	"errors"
	"sync"
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

// ReactionFilter decides whether a reaction should be collected.
type ReactionFilter func(*ReactionContext) bool

type interactionCollector struct {
	filter   InteractionFilter
	filterMu *sync.Mutex
	result   chan *InteractionContext
	multi    bool
}

type messageCollector struct {
	filter   MessageFilter
	filterMu *sync.Mutex
	result   chan *MessageContext
}

type reactionCollector struct {
	filter   ReactionFilter
	filterMu *sync.Mutex
	result   chan *ReactionContext
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
	b.interactionCollectors[id] = interactionCollector{filter: filter, filterMu: &sync.Mutex{}, result: result}
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
	b.messageCollectors[id] = messageCollector{filter: filter, filterMu: &sync.Mutex{}, result: result}
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
	type pending struct {
		id        uint64
		collector interactionCollector
	}
	pendingCollectors := make([]pending, 0, len(b.interactionCollectors))
	for id, collector := range b.interactionCollectors {
		pendingCollectors = append(pendingCollectors, pending{id: id, collector: collector})
	}
	b.collectorMu.Unlock()

	for _, item := range pendingCollectors {
		collector := item.collector
		if collector.filterMu != nil {
			collector.filterMu.Lock()
		}
		b.collectorMu.Lock()
		_, active := b.interactionCollectors[item.id]
		b.collectorMu.Unlock()
		if !active {
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		if collector.filter != nil && !collector.filter(interaction) {
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		b.collectorMu.Lock()
		current, ok := b.interactionCollectors[item.id]
		if !ok {
			b.collectorMu.Unlock()
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		select {
		case current.result <- interaction:
			if !current.multi {
				delete(b.interactionCollectors, item.id)
			}
		default:
		}
		b.collectorMu.Unlock()
		if collector.filterMu != nil {
			collector.filterMu.Unlock()
		}
	}
}

func (b *Bot) publishMessage(message *MessageContext) {
	b.collectorMu.Lock()
	type pending struct {
		id        uint64
		collector messageCollector
	}
	pendingCollectors := make([]pending, 0, len(b.messageCollectors))
	for id, collector := range b.messageCollectors {
		pendingCollectors = append(pendingCollectors, pending{id: id, collector: collector})
	}
	b.collectorMu.Unlock()

	for _, item := range pendingCollectors {
		collector := item.collector
		if collector.filterMu != nil {
			collector.filterMu.Lock()
		}
		b.collectorMu.Lock()
		_, active := b.messageCollectors[item.id]
		b.collectorMu.Unlock()
		if !active {
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		if collector.filter != nil && !collector.filter(message) {
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		b.collectorMu.Lock()
		current, ok := b.messageCollectors[item.id]
		if !ok {
			b.collectorMu.Unlock()
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		select {
		case current.result <- message:
			delete(b.messageCollectors, item.id)
		default:
		}
		b.collectorMu.Unlock()
		if collector.filterMu != nil {
			collector.filterMu.Unlock()
		}
	}
}

// AwaitReaction waits for one reaction matching filter or context
// cancellation. It is useful for confirmation prompts, pagination, and
// reaction-based menus.
func (b *Bot) AwaitReaction(ctx context.Context, filter ReactionFilter) (*ReactionContext, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	result := make(chan *ReactionContext, 1)
	id := b.subscriptions.Add(1)
	b.collectorMu.Lock()
	if b.reactionCollectors == nil {
		b.reactionCollectors = make(map[uint64]reactionCollector)
	}
	b.reactionCollectors[id] = reactionCollector{filter: filter, filterMu: &sync.Mutex{}, result: result}
	b.collectorMu.Unlock()
	defer func() {
		b.collectorMu.Lock()
		delete(b.reactionCollectors, id)
		b.collectorMu.Unlock()
	}()
	select {
	case reaction := <-result:
		return reaction, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (b *Bot) publishReaction(reaction *ReactionContext) {
	b.collectorMu.Lock()
	type pending struct {
		id        uint64
		collector reactionCollector
	}
	pendingCollectors := make([]pending, 0, len(b.reactionCollectors))
	for id, collector := range b.reactionCollectors {
		pendingCollectors = append(pendingCollectors, pending{id: id, collector: collector})
	}
	b.collectorMu.Unlock()

	for _, item := range pendingCollectors {
		collector := item.collector
		if collector.filterMu != nil {
			collector.filterMu.Lock()
		}
		b.collectorMu.Lock()
		_, active := b.reactionCollectors[item.id]
		b.collectorMu.Unlock()
		if !active {
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		if collector.filter != nil && !collector.filter(reaction) {
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		b.collectorMu.Lock()
		current, ok := b.reactionCollectors[item.id]
		if !ok {
			b.collectorMu.Unlock()
			if collector.filterMu != nil {
				collector.filterMu.Unlock()
			}
			continue
		}
		select {
		case current.result <- reaction:
			delete(b.reactionCollectors, item.id)
		default:
		}
		b.collectorMu.Unlock()
		if collector.filterMu != nil {
			collector.filterMu.Unlock()
		}
	}
}

// CollectInteractions collects multiple interactions matching filter into a
// channel until ctx is cancelled or the collector is stopped. It is a
// general-purpose component collector for message-level interactions —
// buttons, select menus, and modals — that runs until the caller stops it.
//
// The returned channel is buffered (1) and is closed when the collector
// stops. The returned stop function cancels the collector and is safe to
// call multiple times.
//
// Unlike AwaitInteraction, which collects exactly one interaction,
// CollectInteractions is designed for flows that need to receive multiple
// interactions over time, such as live polls, reaction roles, or
// multi-step wizards.
func (b *Bot) CollectInteractions(ctx context.Context, filter InteractionFilter) (<-chan *InteractionContext, func()) {
	if ctx == nil {
		ctx = context.Background()
	}
	result := make(chan *InteractionContext, 16)
	collectorCtx, cancel := context.WithCancel(ctx)
	id := b.subscriptions.Add(1)
	b.collectorMu.Lock()
	if b.interactionCollectors == nil {
		b.interactionCollectors = make(map[uint64]interactionCollector)
	}
	b.interactionCollectors[id] = interactionCollector{filter: filter, filterMu: &sync.Mutex{}, result: result, multi: true}
	b.collectorMu.Unlock()

	var stopOnce sync.Once
	stop := func() {
		stopOnce.Do(func() {
			cancel()
			b.collectorMu.Lock()
			delete(b.interactionCollectors, id)
			close(result)
			b.collectorMu.Unlock()
		})
	}

	go func() {
		<-collectorCtx.Done()
		stop()
	}()

	return result, stop
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

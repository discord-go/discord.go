package bot

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type eventSubscription struct {
	id      uint64
	once    bool
	handler EventHandler
}

// EventContext is the generic context for a gateway dispatch event.
// Decode can be used when the repository does not yet have a typed event
// struct for the Discord event being consumed.
type EventContext struct {
	BaseContext
	Name string
	Data json.RawMessage
}

// Decode unmarshals the event data into a caller-provided value.
func (e *EventContext) Decode(value any) error {
	return e.BaseContext.Decode(value)
}

// Raw returns a copy of the raw event data.
func (e *EventContext) Raw() json.RawMessage {
	return append(json.RawMessage(nil), e.Data...)
}

// OnEvent registers a handler for a named Discord dispatch event. The returned
// function unsubscribes that handler and is safe to call more than once.
func (b *Bot) OnEvent(name string, handler EventHandler) func() {
	return b.subscribeEvent(name, false, handler)
}

// On is a short alias for OnEvent, useful for events not yet modeled by the
// repository's typed event structs.
func (b *Bot) On(name string, handler EventHandler) func() {
	return b.OnEvent(name, handler)
}

// OnceEvent registers a handler that is removed after its first invocation.
func (b *Bot) OnceEvent(name string, handler EventHandler) func() {
	return b.subscribeEvent(name, true, handler)
}

// Once is a short alias for OnceEvent.
func (b *Bot) Once(name string, handler EventHandler) func() {
	return b.OnceEvent(name, handler)
}

func (b *Bot) subscribeEvent(name string, once bool, handler EventHandler) func() {
	name = strings.ToUpper(strings.TrimSpace(name))
	if b == nil || name == "" || handler == nil {
		return func() {}
	}
	id := b.subscriptions.Add(1)
	b.mu.Lock()
	if b.eventHandlers == nil {
		b.eventHandlers = make(map[string][]eventSubscription)
	}
	b.eventHandlers[name] = append(b.eventHandlers[name], eventSubscription{id: id, once: once, handler: handler})
	b.mu.Unlock()

	var onceUnsubscribe sync.Once
	return func() {
		onceUnsubscribe.Do(func() { b.removeEventSubscription(name, id) })
	}
}

func (b *Bot) removeEventSubscription(name string, id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	handlers := b.eventHandlers[name]
	for index, subscription := range handlers {
		if subscription.id == id {
			b.eventHandlers[name] = append(handlers[:index], handlers[index+1:]...)
			return
		}
	}
}

func (b *Bot) dispatchEvent(name string, data json.RawMessage) {
	b.mu.Lock()
	handlers := append([]eventSubscription(nil), b.eventHandlers[name]...)
	for _, subscription := range handlers {
		if !subscription.once {
			continue
		}
		current := b.eventHandlers[name]
		for index, candidate := range current {
			if candidate.id == subscription.id {
				b.eventHandlers[name] = append(current[:index], current[index+1:]...)
				break
			}
		}
	}
	b.mu.Unlock()
	if len(handlers) == 0 {
		return
	}
	ctx := &EventContext{BaseContext: b.baseContextWithRaw(data), Name: name, Data: append(json.RawMessage(nil), data...)}
	for _, subscription := range handlers {
		b.invoke(name, func() { subscription.handler(ctx) })
	}
}

// OnInteractionCreate is the discoverable Discord.js-style alias for
// OnInteraction.
func (b *Bot) OnInteractionCreate(handler InteractionHandler) {
	b.OnInteraction(handler)
}

// OnChannelCreate registers a handler for CHANNEL_CREATE events.
func (b *Bot) OnChannelCreate(handler ChannelCreateHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.channelCreate = append(b.channelCreate, handler)
	b.mu.Unlock()
}

// OnChannelUpdate registers a handler for CHANNEL_UPDATE events.
func (b *Bot) OnChannelUpdate(handler ChannelUpdateHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.channelUpdate = append(b.channelUpdate, handler)
	b.mu.Unlock()
}

// OnGuildAuditLogEntryCreate registers a handler for audit log events.
func (b *Bot) OnGuildAuditLogEntryCreate(handler GuildAuditLogEntryCreateHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.auditLogCreate = append(b.auditLogCreate, handler)
	b.mu.Unlock()
}

// OnPresenceUpdate registers a generic PRESENCE_UPDATE handler until a typed
// presence model is added to the repository.
func (b *Bot) OnPresenceUpdate(handler EventHandler) func() {
	return b.OnEvent("PRESENCE_UPDATE", handler)
}

// OnChannelDelete registers a generic CHANNEL_DELETE handler.
func (b *Bot) OnChannelDelete(handler EventHandler) func() {
	return b.OnEvent("CHANNEL_DELETE", handler)
}

// OnMessageReactionRemove registers a generic MESSAGE_REACTION_REMOVE handler.
func (b *Bot) OnMessageReactionRemove(handler EventHandler) func() {
	return b.OnEvent("MESSAGE_REACTION_REMOVE", handler)
}

// OnTypingStart registers a generic TYPING_START handler.
func (b *Bot) OnTypingStart(handler EventHandler) func() {
	return b.OnEvent("TYPING_START", handler)
}

// Common raw-event aliases cover gateway payloads that do not yet have a
// dedicated model in the events package.
func (b *Bot) OnGuildMemberAdd(handler EventHandler) func() {
	return b.OnEvent("GUILD_MEMBER_ADD", handler)
}
func (b *Bot) OnGuildMemberRemove(handler EventHandler) func() {
	return b.OnEvent("GUILD_MEMBER_REMOVE", handler)
}
func (b *Bot) OnGuildMemberUpdate(handler EventHandler) func() {
	return b.OnEvent("GUILD_MEMBER_UPDATE", handler)
}
func (b *Bot) OnVoiceStateUpdate(handler EventHandler) func() {
	return b.OnEvent("VOICE_STATE_UPDATE", handler)
}
func (b *Bot) OnVoiceServerUpdate(handler EventHandler) func() {
	return b.OnEvent("VOICE_SERVER_UPDATE", handler)
}
func (b *Bot) OnInviteCreate(handler EventHandler) func() { return b.OnEvent("INVITE_CREATE", handler) }
func (b *Bot) OnInviteDelete(handler EventHandler) func() { return b.OnEvent("INVITE_DELETE", handler) }
func (b *Bot) OnAutoModerationAction(handler EventHandler) func() {
	return b.OnEvent("AUTO_MODERATION_ACTION_EXECUTION", handler)
}
func (b *Bot) OnThreadCreate(handler EventHandler) func() { return b.OnEvent("THREAD_CREATE", handler) }
func (b *Bot) OnThreadDelete(handler EventHandler) func() { return b.OnEvent("THREAD_DELETE", handler) }
func (b *Bot) OnThreadUpdate(handler EventHandler) func() { return b.OnEvent("THREAD_UPDATE", handler) }
func (b *Bot) OnReactionRemove(handler EventHandler) func() {
	return b.OnEvent("MESSAGE_REACTION_REMOVE", handler)
}

// OnReconnect runs when Discord requests a reconnect with gateway OP 7.
func (b *Bot) OnReconnect(handler LifecycleHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.reconnectHandlers = append(b.reconnectHandlers, handler)
	b.mu.Unlock()
}

// OnResume runs after a gateway session resumes successfully.
func (b *Bot) OnResume(handler LifecycleHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.resumeHandlers = append(b.resumeHandlers, handler)
	b.mu.Unlock()
}

// OnInvalidated runs when Discord invalidates the gateway session.
func (b *Bot) OnInvalidated(handler LifecycleHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.invalidatedHandlers = append(b.invalidatedHandlers, handler)
	b.mu.Unlock()
}

// OnDisconnect runs whenever the gateway loop terminates.
func (b *Bot) OnDisconnect(handler LifecycleHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.disconnectHandlers = append(b.disconnectHandlers, handler)
	b.mu.Unlock()
}

func (b *Bot) dispatchLifecycle(handlers []LifecycleHandler, event string) {
	for _, handler := range handlers {
		b.invoke(event, handler)
	}
}

func (b *Bot) markReady() {
	b.stateMu.Lock()
	if b.ready {
		b.stateMu.Unlock()
		return
	}
	b.ready = true
	b.readyAt = time.Now()
	readyCh := b.readyCh
	b.stateMu.Unlock()
	if readyCh != nil {
		close(readyCh)
	}
}

// WaitReady blocks until the bot receives READY or the context is cancelled.
func (b *Bot) WaitReady(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	b.stateMu.RLock()
	ready := b.ready
	readyCh := b.readyCh
	b.stateMu.RUnlock()
	if ready {
		return nil
	}
	if readyCh == nil {
		return ErrBotNotRunning
	}
	select {
	case <-readyCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsReady reports whether the current run has received READY.
func (b *Bot) IsReady() bool {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	return b.ready
}

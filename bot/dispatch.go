package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"
)

// handleRawDispatch parses a gateway dispatch, builds typed contexts, and
// fans out to registered handlers without blocking the gateway read loop.
func (b *Bot) handleRawDispatch(data []byte) {
	var payload gateway.GatewayPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		b.reportError(fmt.Errorf("parse gateway payload: %w", err))
		return
	}
	if payload.Type == nil {
		switch payload.Op {
		case gateway.OpcodeReconnect:
			b.mu.RLock()
			handlers := append([]LifecycleHandler(nil), b.reconnectHandlers...)
			b.mu.RUnlock()
			b.dispatchLifecycle(handlers, "RECONNECT")
		case gateway.OpcodeInvalidSession:
			b.mu.RLock()
			handlers := append([]LifecycleHandler(nil), b.invalidatedHandlers...)
			b.mu.RUnlock()
			b.dispatchLifecycle(handlers, "INVALIDATED")
		}
		return
	}
	event := *payload.Type
	b.eventsReceived.Add(1)
	b.dispatchRaw(event, payload.Data)
	b.dispatchEvent(event, payload.Data)

	// Track voice states before typed dispatch so handlers that query the
	// tracker observe the state that triggered them.
	if event == "VOICE_STATE_UPDATE" && b.voice != nil {
		var state voice.VoiceState
		if err := json.Unmarshal(payload.Data, &state); err == nil {
			b.voice.apply(state)
		}
	} else if event == "GUILD_DELETE" && b.voice != nil {
		var deleted struct {
			ID snowflake.ID `json:"id,string"`
		}
		if err := json.Unmarshal(payload.Data, &deleted); err == nil {
			b.voice.dropGuild(deleted.ID)
		}
	}

	switch event {
	case "READY":
		var ready events.Ready
		if err := json.Unmarshal(payload.Data, &ready); err != nil {
			b.reportError(fmt.Errorf("parse READY: %w", err))
			return
		}
		var readyMetadata struct {
			ResumeGatewayURL string `json:"resume_gateway_url"`
		}
		_ = json.Unmarshal(payload.Data, &readyMetadata)
		b.stateMu.Lock()
		b.appID = ready.User.ID
		readyUser := ready.User
		b.user = &readyUser
		if b.gwClient != nil && b.gwClient.Session != nil {
			b.gwClient.Session.SetSessionID(ready.SessionID)
			resumeURL := readyMetadata.ResumeGatewayURL
			if resumeURL == "" {
				resumeURL = b.gatewayURL
			}
			b.gwClient.Session.SetResumeURL(resumeURL)
		}
		b.stateMu.Unlock()
		b.markReady()
		go b.reapplyPresence()
		b.logger.Printf("logged in as %s#%s (ID: %s)", ready.User.Username, ready.User.Discriminator, ready.User.ID)
		if b.router != nil && b.commandSync.Mode != CommandSyncDisabled {
			go b.syncCommands()
		}
		ctx := &ReadyContext{BaseContext: b.baseContext(), Ready: &ready}
		b.mu.RLock()
		handlers := append([]ReadyHandler(nil), b.readyHandlers...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("READY", func() { handler(ctx) })
		}

	case "MESSAGE_CREATE":
		var msg events.MessageCreate
		if err := json.Unmarshal(payload.Data, &msg); err != nil {
			b.reportError(fmt.Errorf("parse MESSAGE_CREATE: %w", err))
			return
		}
		if msg.Author != nil && msg.Author.Bot {
			return
		}
		ctx := &MessageContext{BaseContext: b.baseContext(), MessageCreate: &msg}
		b.publishMessage(ctx)
		if b.router != nil {
			content := b.messageCommandContent(ctx)
			if content != "" {
				b.invoke("MESSAGE_CREATE prefix", func() {
					if b.router.handlePrefixContent(ctx, content, b.prefix) {
						b.prefixCommands.Add(1)
					}
				})
			}
		}
		b.mu.RLock()
		handlers := append([]MessageHandler(nil), b.messageHandlers...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("MESSAGE_CREATE", func() { handler(ctx) })
		}

	case "MESSAGE_UPDATE":
		var msg events.MessageUpdate
		if err := json.Unmarshal(payload.Data, &msg); err != nil {
			b.reportError(fmt.Errorf("parse MESSAGE_UPDATE: %w", err))
			return
		}
		ctx := &MessageUpdateContext{BaseContext: b.baseContext(), MessageUpdate: &msg}
		b.mu.RLock()
		handlers := append([]MessageUpdateHandler(nil), b.messageUpdate...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("MESSAGE_UPDATE", func() { handler(ctx) })
		}

	case "MESSAGE_DELETE":
		var deleted events.MessageDelete
		if err := json.Unmarshal(payload.Data, &deleted); err != nil {
			b.reportError(fmt.Errorf("parse MESSAGE_DELETE: %w", err))
			return
		}
		ctx := &MessageDeleteContext{BaseContext: b.baseContext(), MessageDelete: &deleted}
		b.mu.RLock()
		handlers := append([]MessageDeleteHandler(nil), b.messageDelete...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("MESSAGE_DELETE", func() { handler(ctx) })
		}

	case "INTERACTION_CREATE":
		var interaction interactions.Interaction
		if err := json.Unmarshal(payload.Data, &interaction); err != nil {
			b.reportError(fmt.Errorf("parse INTERACTION_CREATE: %w", err))
			return
		}
		ctx := newInteractionContext(b.baseContext(), &interaction)
		b.publishInteraction(ctx)
		if b.router != nil && (interaction.Type == interactions.InteractionTypeApplicationCommand || interaction.Type == interactions.InteractionTypeApplicationCommandAutocomplete || interaction.Type == interactions.InteractionTypeMessageComponent || interaction.Type == interactions.InteractionTypeModalSubmit) {
			b.invoke("INTERACTION_CREATE command", func() {
				if b.router.handleInteraction(ctx) {
					b.slashCommands.Add(1)
				}
			})
		}
		b.mu.RLock()
		handlers := append([]InteractionHandler(nil), b.interactionHandlers...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("INTERACTION_CREATE", func() { handler(ctx) })
		}

	case "MESSAGE_REACTION_ADD":
		var reaction events.MessageReactionAdd
		if err := json.Unmarshal(payload.Data, &reaction); err != nil {
			b.reportError(fmt.Errorf("parse MESSAGE_REACTION_ADD: %w", err))
			return
		}
		ctx := &ReactionContext{BaseContext: b.baseContext(), MessageReactionAdd: &reaction}
		b.publishReaction(ctx)
		b.mu.RLock()
		handlers := append([]ReactionHandler(nil), b.reactionHandlers...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("MESSAGE_REACTION_ADD", func() { handler(ctx) })
		}

	case "CHANNEL_CREATE":
		var channel events.ChannelCreate
		if err := json.Unmarshal(payload.Data, &channel); err != nil {
			b.reportError(fmt.Errorf("parse CHANNEL_CREATE: %w", err))
			return
		}
		ctx := &ChannelContext{BaseContext: b.baseContext(), ChannelCreate: &channel}
		b.mu.RLock()
		handlers := append([]ChannelCreateHandler(nil), b.channelCreate...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("CHANNEL_CREATE", func() { handler(ctx) })
		}

	case "CHANNEL_UPDATE":
		var channel events.ChannelUpdate
		if err := json.Unmarshal(payload.Data, &channel); err != nil {
			b.reportError(fmt.Errorf("parse CHANNEL_UPDATE: %w", err))
			return
		}
		ctx := &ChannelUpdateContext{BaseContext: b.baseContext(), ChannelUpdate: &channel}
		b.mu.RLock()
		handlers := append([]ChannelUpdateHandler(nil), b.channelUpdate...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("CHANNEL_UPDATE", func() { handler(ctx) })
		}

	case "GUILD_AUDIT_LOG_ENTRY_CREATE":
		var entry events.GuildAuditLogEntryCreate
		if err := json.Unmarshal(payload.Data, &entry); err != nil {
			b.reportError(fmt.Errorf("parse GUILD_AUDIT_LOG_ENTRY_CREATE: %w", err))
			return
		}
		ctx := &GuildAuditLogEntryContext{BaseContext: b.baseContext(), GuildAuditLogEntryCreate: &entry}
		b.mu.RLock()
		handlers := append([]GuildAuditLogEntryCreateHandler(nil), b.auditLogCreate...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("GUILD_AUDIT_LOG_ENTRY_CREATE", func() { handler(ctx) })
		}

	case "RESUMED":
		b.mu.RLock()
		handlers := append([]LifecycleHandler(nil), b.resumeHandlers...)
		b.mu.RUnlock()
		b.dispatchLifecycle(handlers, "RESUMED")

	case "GUILD_CREATE":
		var guild events.GuildCreate
		if err := json.Unmarshal(payload.Data, &guild); err != nil {
			b.reportError(fmt.Errorf("parse GUILD_CREATE: %w", err))
			return
		}
		ctx := &GuildContext{BaseContext: b.baseContext(), GuildCreate: &guild}
		b.mu.RLock()
		handlers := append([]GuildCreateHandler(nil), b.guildCreateHandlers...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("GUILD_CREATE", func() { handler(ctx) })
		}

	case "GUILD_UPDATE":
		var guild events.GuildUpdate
		if err := json.Unmarshal(payload.Data, &guild); err != nil {
			b.reportError(fmt.Errorf("parse GUILD_UPDATE: %w", err))
			return
		}
		ctx := &GuildUpdateContext{BaseContext: b.baseContext(), GuildUpdate: &guild}
		b.mu.RLock()
		handlers := append([]GuildUpdateHandler(nil), b.guildUpdateHandlers...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("GUILD_UPDATE", func() { handler(ctx) })
		}

	case "GUILD_DELETE":
		var guild events.GuildDelete
		if err := json.Unmarshal(payload.Data, &guild); err != nil {
			b.reportError(fmt.Errorf("parse GUILD_DELETE: %w", err))
			return
		}
		ctx := &GuildDeleteContext{BaseContext: b.baseContext(), GuildDelete: &guild}
		b.mu.RLock()
		handlers := append([]GuildDeleteHandler(nil), b.guildDeleteHandlers...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			b.invoke("GUILD_DELETE", func() { handler(ctx) })
		}
	}
}

func (b *Bot) messageCommandContent(ctx *MessageContext) string {
	content := ctx.Content
	if b.prefix != "" && strings.HasPrefix(content, b.prefix) {
		return strings.TrimPrefix(content, b.prefix)
	}
	if b.botName != "" && strings.HasPrefix(strings.ToLower(content), strings.ToLower(b.botName)) {
		rest := content[len(b.botName):]
		if rest == "" || strings.HasPrefix(rest, " ") || strings.HasPrefix(rest, "\t") {
			return strings.TrimSpace(rest)
		}
	}
	if !b.mentionTriggers {
		return ""
	}
	appID := b.AppID()
	if appID == 0 {
		return ""
	}
	for _, mention := range []string{"<@" + appID.String() + ">", "<@!" + appID.String() + ">"} {
		if strings.HasPrefix(content, mention) {
			return strings.TrimSpace(strings.TrimPrefix(content, mention))
		}
	}
	return ""
}

func (b *Bot) dispatchRaw(event string, data json.RawMessage) {
	b.mu.RLock()
	handlers := append([]RawEventHandler(nil), b.rawHandlers...)
	b.mu.RUnlock()
	if len(handlers) == 0 {
		return
	}
	copyData := append(json.RawMessage(nil), data...)
	base := b.baseContext()
	for _, handler := range handlers {
		b.invoke(event+" raw", func() { handler(base.Context(), event, copyData) })
	}
}

func (b *Bot) invoke(event string, handler func()) {
	b.handlerWG.Add(1)
	go func() {
		defer b.handlerWG.Done()
		if b.handlerSlots != nil {
			b.handlerSlots <- struct{}{}
			defer func() { <-b.handlerSlots }()
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				b.handlerPanics.Add(1)
				b.reportError(&HandlerPanicError{Event: event, Value: recovered})
			}
		}()
		handler()
	}()
}

func (b *Bot) reportError(err error) {
	if err == nil {
		return
	}
	b.mu.RLock()
	defaultHandler := b.errorHandler
	handlers := append([]ErrorHandler(nil), b.errorHandlers...)
	b.mu.RUnlock()
	call := func(handler ErrorHandler) {
		defer func() { _ = recover() }()
		handler(err)
	}
	if defaultHandler != nil {
		call(defaultHandler)
	}
	for _, handler := range handlers {
		call(handler)
	}
}

func (b *Bot) syncCommands() {
	b.syncMu.Lock()
	defer b.syncMu.Unlock()
	if b.router == nil || b.commandSync.Mode == CommandSyncDisabled {
		return
	}
	cmds, err := b.router.slashCommands()
	if err != nil {
		b.reportError(fmt.Errorf("validate slash commands: %w", err))
		return
	}
	b.stateMu.RLock()
	appID := b.appID
	parent := b.runCtx
	b.stateMu.RUnlock()
	if appID == 0 {
		b.reportError(errors.New("bot: cannot sync commands before application ID is known"))
		return
	}
	if parent == nil {
		parent = context.Background()
	}
	timeout := b.commandSync.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	var registered []interactions.ApplicationCommand
	switch b.commandSync.Mode {
	case CommandSyncGuild:
		if b.commandSync.GuildID == 0 {
			b.reportError(errors.New("bot: guild command sync requires a guild ID"))
			return
		}
		registered, err = b.Rest.BulkOverwriteGuildCommands(ctx, appID, b.commandSync.GuildID, cmds)
	default:
		registered, err = b.Rest.BulkOverwriteGlobalCommands(ctx, appID, cmds)
	}
	if err != nil {
		b.reportError(fmt.Errorf("sync slash commands: %w", err))
		return
	}
	b.commandSyncs.Add(1)
	names := make([]string, 0, len(registered))
	for _, command := range registered {
		names = append(names, "/"+command.Name)
	}
	sort.Strings(names)
	b.logger.Printf("registered %d slash commands: %s", len(registered), strings.Join(names, ", "))
}

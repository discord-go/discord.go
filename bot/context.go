package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// BaseContext provides shared functionality for all event context types.
type BaseContext struct {
	// Bot is the bot instance that received this event.
	Bot *Bot

	rest    *rest.Client
	ctx     context.Context
	timeout time.Duration
}

// Context returns the run context associated with this event.
func (c *BaseContext) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

func (c *BaseContext) newCtx() (context.Context, context.CancelFunc) {
	d := c.timeout
	if d <= 0 {
		d = 10 * time.Second
	}
	return context.WithTimeout(c.Context(), d)
}

// ReadyContext wraps the READY gateway event.
type ReadyContext struct {
	BaseContext
	*events.Ready
}

// MessageContext wraps a MESSAGE_CREATE event with convenience methods.
type MessageContext struct {
	BaseContext
	*events.MessageCreate
}

// ChannelID returns the channel ID where the message was sent.
// This is a method for consistency with InteractionContext.ChannelID();
// the underlying value comes from the embedded MessageCreate event.
func (m *MessageContext) ChannelID() snowflake.ID {
	return m.MessageCreate.ChannelID
}

// Reply sends a text message to the same channel.
func (m *MessageContext) Reply(content string) (*messages.Message, error) {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.CreateMessage(ctx, m.ChannelID(), content)
}

// Replyf sends a formatted text message to the same channel.
func (m *MessageContext) Replyf(format string, args ...interface{}) (*messages.Message, error) {
	return m.Reply(fmt.Sprintf(format, args...))
}

// ReplyEmbed sends one or more embeds to the same channel.
func (m *MessageContext) ReplyEmbed(embeds ...messages.Embed) (*messages.Message, error) {
	return m.ReplyComplex(messages.MessageSend{Embeds: embeds})
}

// ReplyComplex sends a fully customized message.
func (m *MessageContext) ReplyComplex(send messages.MessageSend) (*messages.Message, error) {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.CreateMessageComplex(ctx, m.ChannelID(), send)
}

// ReplyComplexWithFiles sends a customized message with multipart attachments.
func (m *MessageContext) ReplyComplexWithFiles(send messages.MessageSend, files ...rest.File) (*messages.Message, error) {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.CreateMessageComplexWithFiles(ctx, m.ChannelID(), send, files)
}

// Edit edits this message.
func (m *MessageContext) Edit(content string) (*messages.Message, error) {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.EditMessage(ctx, m.ChannelID(), m.ID, rest.EditMessageParams{Content: &content})
}

// Fetch retrieves the current message from Discord.
func (m *MessageContext) Fetch() (*messages.Message, error) {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.GetChannelMessage(ctx, m.ChannelID(), m.ID)
}

// React adds a reaction emoji to this message.
func (m *MessageContext) React(emoji string) error {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.CreateReaction(ctx, m.ChannelID(), m.ID, emoji)
}

// Delete deletes this message.
func (m *MessageContext) Delete() error {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.DeleteMessage(ctx, m.ChannelID(), m.ID)
}

// Pin pins this message in its channel.
func (m *MessageContext) Pin() error {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.PinMessage(ctx, m.ChannelID(), m.ID)
}

// Unpin removes this message from its channel's pins.
func (m *MessageContext) Unpin() error {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.UnpinMessage(ctx, m.ChannelID(), m.ID)
}

// TriggerTyping starts the typing indicator in this message's channel.
func (m *MessageContext) TriggerTyping() error {
	ctx, cancel := m.newCtx()
	defer cancel()
	return m.rest.TriggerTypingIndicator(ctx, m.ChannelID())
}

// MessageUpdateContext wraps a MESSAGE_UPDATE event.
type MessageUpdateContext struct {
	BaseContext
	*events.MessageUpdate
}

// MessageDeleteContext wraps a MESSAGE_DELETE event.
type MessageDeleteContext struct {
	BaseContext
	*events.MessageDelete
}

// InteractionContext wraps an INTERACTION_CREATE event.
type InteractionContext struct {
	BaseContext
	*interactions.Interaction

	cmdData       *interactions.ApplicationCommandInteractionData
	data          *interactionData
	matchedPrefix string
	responded     bool
	deferred      bool
	ephemeral     bool
	responseMu    sync.RWMutex
}

type interactionData struct {
	CustomID      string                                                 `json:"custom_id,omitempty"`
	ComponentType components.ComponentType                               `json:"component_type,omitempty"`
	Values        []string                                               `json:"values,omitempty"`
	Options       []interactions.ApplicationCommandInteractionDataOption `json:"options,omitempty"`
	Components    []modalComponent                                       `json:"components,omitempty"`
}

type modalComponent struct {
	CustomID   string           `json:"custom_id,omitempty"`
	Value      string           `json:"value,omitempty"`
	Components []modalComponent `json:"components,omitempty"`
}

func newInteractionContext(base BaseContext, interaction *interactions.Interaction) *InteractionContext {
	ctx := &InteractionContext{BaseContext: base, Interaction: interaction}
	if (interaction.Type == interactions.InteractionTypeApplicationCommand || interaction.Type == interactions.InteractionTypeApplicationCommandAutocomplete) && len(interaction.Data) > 0 {
		var commandData interactions.ApplicationCommandInteractionData
		decoder := json.NewDecoder(bytes.NewReader(interaction.Data))
		decoder.UseNumber()
		if err := decoder.Decode(&commandData); err == nil {
			ctx.cmdData = &commandData
		}
	}
	if len(interaction.Data) > 0 {
		var data interactionData
		decoder := json.NewDecoder(bytes.NewReader(interaction.Data))
		decoder.UseNumber()
		if err := decoder.Decode(&data); err == nil {
			ctx.data = &data
		}
	}
	return ctx
}

// CommandName returns the slash command name.
func (i *InteractionContext) CommandName() string {
	if i.cmdData != nil {
		return i.cmdData.Name
	}
	return ""
}

// CommandType returns the application command type.
func (i *InteractionContext) CommandType() interactions.ApplicationCommandType {
	if i.cmdData != nil {
		return interactions.ApplicationCommandType(i.cmdData.Type)
	}
	return 0
}

// CustomID returns the component or modal custom ID.
func (i *InteractionContext) CustomID() string {
	if i.data == nil {
		return ""
	}
	return i.data.CustomID
}

// Suffix returns the portion of the custom ID after the matched prefix.
// When the interaction was dispatched via ButtonPrefix, SelectPrefix, or
// ModalPrefix, Suffix returns the part of the custom ID that follows the
// registered prefix. For exact-match routes (Button, Select, Modal) or when
// no prefix was matched, Suffix returns an empty string.
//
// For example, if ButtonPrefix("ticket:close:", handler) matches a button
// with custom ID "ticket:close:confirm:123456789", Suffix() returns
// "confirm:123456789".
func (i *InteractionContext) Suffix() string {
	if i.matchedPrefix == "" {
		return ""
	}
	return strings.TrimPrefix(i.CustomID(), i.matchedPrefix)
}

// ComponentType returns the component type for a message component interaction.
func (i *InteractionContext) ComponentType() components.ComponentType {
	if i.data == nil {
		return 0
	}
	return i.data.ComponentType
}

// Values returns selected values from a select menu interaction.
func (i *InteractionContext) Values() []string {
	if i.data == nil {
		return nil
	}
	return append([]string(nil), i.data.Values...)
}

// ModalRow represents a single action row within a modal submission.
// Each row contains the text input components that were submitted in that row.
type ModalRow struct {
	CustomID   string       `json:"custom_id,omitempty"`
	Components []ModalField `json:"components,omitempty"`
}

// ModalField represents a single text input field within a modal action row.
type ModalField struct {
	CustomID string `json:"custom_id,omitempty"`
	Value    string `json:"value,omitempty"`
}

// ModalValue returns a submitted text input value by custom ID.
func (i *InteractionContext) ModalValue(customID string) string {
	for _, component := range i.modalComponents() {
		if component.CustomID == customID {
			return component.Value
		}
	}
	return ""
}

// ModalValues returns all submitted modal text input values keyed by custom ID.
// The map is flat — if two inputs share a custom ID, the last one wins.
// Use ModalRows for the structured action-row hierarchy.
func (i *InteractionContext) ModalValues() map[string]string {
	values := make(map[string]string)
	for _, component := range i.modalComponents() {
		if component.CustomID != "" {
			values[component.CustomID] = component.Value
		}
	}
	return values
}

// ModalRows returns the submitted modal values structured by action row.
// Each row corresponds to a top-level action row in the modal, and each
// row's components are the text inputs within that row. This preserves the
// modal's layout hierarchy, unlike the flat ModalValues map.
func (i *InteractionContext) ModalRows() []ModalRow {
	if i.data == nil {
		return nil
	}
	rows := make([]ModalRow, 0, len(i.data.Components))
	for _, row := range i.data.Components {
		modalRow := ModalRow{CustomID: row.CustomID}
		for _, child := range row.Components {
			if child.CustomID != "" {
				modalRow.Components = append(modalRow.Components, ModalField{
					CustomID: child.CustomID,
					Value:    child.Value,
				})
			}
		}
		rows = append(rows, modalRow)
	}
	return rows
}

// FocusedOption returns the option currently focused for autocomplete.
func (i *InteractionContext) FocusedOption() *interactions.ApplicationCommandInteractionDataOption {
	return findFocusedOption(i.Options())
}

// FocusedOptionString returns the focused option's value as a string. It is
// a safe accessor for autocomplete handlers — if the focused value is nil,
// not a string, or no option is focused, it returns an empty string.
func (i *InteractionContext) FocusedOptionString() string {
	focused := i.FocusedOption()
	if focused == nil || focused.Value == nil {
		return ""
	}
	if s, ok := focused.Value.(string); ok {
		return s
	}
	return ""
}

// IsChatInputCommand reports whether this is a slash command interaction.
func (i *InteractionContext) IsChatInputCommand() bool {
	return i.Type == interactions.InteractionTypeApplicationCommand && i.CommandType() == interactions.ApplicationCommandTypeChatInput
}

// IsCommand reports whether this is any application command interaction.
func (i *InteractionContext) IsCommand() bool {
	return i.Type == interactions.InteractionTypeApplicationCommand || i.IsAutocomplete()
}

// IsContextMenuCommand reports whether this is a user or message context command.
func (i *InteractionContext) IsContextMenuCommand() bool {
	return i.Type == interactions.InteractionTypeApplicationCommand && (i.CommandType() == interactions.ApplicationCommandTypeUser || i.CommandType() == interactions.ApplicationCommandTypeMessage)
}

// IsUserContextMenuCommand reports whether this is a user context command.
func (i *InteractionContext) IsUserContextMenuCommand() bool {
	return i.IsContextMenuCommand() && i.CommandType() == interactions.ApplicationCommandTypeUser
}

// IsMessageContextMenuCommand reports whether this is a message context command.
func (i *InteractionContext) IsMessageContextMenuCommand() bool {
	return i.IsContextMenuCommand() && i.CommandType() == interactions.ApplicationCommandTypeMessage
}

// IsRepliable reports whether this interaction accepts a callback response.
func (i *InteractionContext) IsRepliable() bool {
	switch i.Type {
	case interactions.InteractionTypePing, interactions.InteractionTypeApplicationCommand, interactions.InteractionTypeMessageComponent, interactions.InteractionTypeApplicationCommandAutocomplete, interactions.InteractionTypeModalSubmit:
		return true
	default:
		return false
	}
}

// InGuild reports whether the interaction was received in a guild.
func (i *InteractionContext) InGuild() bool {
	return i.Interaction.GuildID != nil
}

// User returns the invoking user. In a guild interaction the user is nested
// inside Member; in a DM interaction it is on the top-level User field. This
// accessor handles both cases and returns nil when neither is present.
func (i *InteractionContext) User() *users.User {
	if i.Interaction.User != nil {
		return i.Interaction.User
	}
	if i.Interaction.Member != nil && i.Interaction.Member.User != nil {
		return i.Interaction.Member.User
	}
	return nil
}

// GuildID returns the guild ID for a guild interaction, or zero for a DM
// interaction. It dereferences the pointer so callers do not need to check
// nil themselves.
func (i *InteractionContext) GuildID() snowflake.ID {
	if i.Interaction.GuildID != nil {
		return *i.Interaction.GuildID
	}
	return 0
}

// ChannelID returns the channel ID where the interaction occurred, or zero
// when absent. It dereferences the pointer so callers do not need to check
// nil themselves.
func (i *InteractionContext) ChannelID() snowflake.ID {
	if i.Interaction.ChannelID != nil {
		return *i.Interaction.ChannelID
	}
	return 0
}

// IsAutocomplete reports whether this is an autocomplete interaction.
func (i *InteractionContext) IsAutocomplete() bool {
	return i.Type == interactions.InteractionTypeApplicationCommandAutocomplete
}

// IsMessageComponent reports whether this is a button or select interaction.
func (i *InteractionContext) IsMessageComponent() bool {
	return i.Type == interactions.InteractionTypeMessageComponent
}

// IsButton reports whether this is a button interaction.
func (i *InteractionContext) IsButton() bool {
	return i.IsMessageComponent() && i.ComponentType() == components.ComponentTypeButton
}

// IsSelectMenu reports whether this is any select menu interaction.
func (i *InteractionContext) IsSelectMenu() bool {
	if !i.IsMessageComponent() {
		return false
	}
	typ := i.ComponentType()
	return typ >= components.ComponentTypeStringSelect && typ <= components.ComponentTypeChannelSelect
}

// IsModalSubmit reports whether this is a modal submission.
func (i *InteractionContext) IsModalSubmit() bool {
	return i.Type == interactions.InteractionTypeModalSubmit
}

// MemberPermissions returns the invoking member's channel permissions.
func (i *InteractionContext) MemberPermissions() permissions.Permission {
	if i.Member == nil {
		return 0
	}
	return i.Member.Permissions
}

// BotPermissions returns the bot's channel permissions from the interaction.
func (i *InteractionContext) BotPermissions() permissions.Permission {
	value, _ := strconv.ParseUint(i.AppPermissions, 10, 64)
	return permissions.Permission(value)
}

// TargetID returns the target ID for a user or message context command.
func (i *InteractionContext) TargetID() snowflake.ID {
	if i.cmdData == nil {
		return 0
	}
	id, _ := snowflake.Parse(i.cmdData.TargetID)
	return id
}

// FetchTargetMember fetches the target user's guild member via REST. It is
// useful for moderation commands that need to inspect the target's roles
// (e.g., to prevent banning another moderator). The interaction must be
// in a guild; in a DM it returns nil.
func (i *InteractionContext) FetchTargetMember(ctx context.Context) (*users.Member, error) {
	if !i.InGuild() {
		return nil, nil
	}
	targetID := i.TargetID()
	if targetID == 0 {
		return nil, nil
	}
	return i.rest.GetGuildMember(ctx, i.GuildID(), targetID)
}

// Options returns all top-level command options provided by the user.
func (i *InteractionContext) Options() []interactions.ApplicationCommandInteractionDataOption {
	if i.cmdData != nil {
		return i.cmdData.Options
	}
	return nil
}

// Subcommand returns the selected subcommand name, if any.
func (i *InteractionContext) Subcommand() string {
	option := firstOptionOfType(i.Options(), interactions.ApplicationCommandOptionTypeSubCommand)
	if option != nil {
		return option.Name
	}
	return ""
}

// SubcommandGroup returns the selected subcommand group name, if any.
func (i *InteractionContext) SubcommandGroup() string {
	option := firstOptionOfType(i.Options(), interactions.ApplicationCommandOptionTypeSubCommandGroup)
	if option != nil {
		return option.Name
	}
	return ""
}

// HasOption reports whether an option exists, including nested subcommand
// options.
func (i *InteractionContext) HasOption(name string) bool {
	return i.GetOption(name) != nil
}

// GetOption returns a specific option by name, searching nested options too.
func (i *InteractionContext) GetOption(name string) *interactions.ApplicationCommandInteractionDataOption {
	return findOption(i.Options(), name)
}

// GetStringOption returns a string option value, or an empty string if absent.
func (i *InteractionContext) GetStringOption(name string) string {
	option := i.GetOption(name)
	if option == nil || option.Value == nil {
		return ""
	}
	if value, ok := option.Value.(string); ok {
		return value
	}
	return fmt.Sprint(option.Value)
}

// GetIntOption returns an integer option value, or zero if absent or invalid.
func (i *InteractionContext) GetIntOption(name string) int64 {
	option := i.GetOption(name)
	if option == nil || option.Value == nil {
		return 0
	}
	switch value := option.Value.(type) {
	case json.Number:
		result, _ := value.Int64()
		return result
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case string:
		result, _ := strconv.ParseInt(value, 10, 64)
		return result
	default:
		return 0
	}
}

// GetFloatOption returns a numeric option value, or zero if absent or invalid.
func (i *InteractionContext) GetFloatOption(name string) float64 {
	option := i.GetOption(name)
	if option == nil || option.Value == nil {
		return 0
	}
	switch value := option.Value.(type) {
	case json.Number:
		result, _ := value.Float64()
		return result
	case float64:
		return value
	case string:
		result, _ := strconv.ParseFloat(value, 64)
		return result
	default:
		return 0
	}
}

// GetBoolOption returns a boolean option value, or false if absent or invalid.
func (i *InteractionContext) GetBoolOption(name string) bool {
	option := i.GetOption(name)
	if option == nil || option.Value == nil {
		return false
	}
	if value, ok := option.Value.(bool); ok {
		return value
	}
	value, _ := strconv.ParseBool(fmt.Sprint(option.Value))
	return value
}

// GetUserID returns a user option as a snowflake.ID.
func (i *InteractionContext) GetUserID(name string) snowflake.ID {
	return i.getSnowflakeOption(name)
}

// GetRoleID returns a role option as a snowflake.ID.
func (i *InteractionContext) GetRoleID(name string) snowflake.ID {
	return i.getSnowflakeOption(name)
}

// GetChannelID returns a channel option as a snowflake.ID.
func (i *InteractionContext) GetChannelID(name string) snowflake.ID {
	return i.getSnowflakeOption(name)
}

func (i *InteractionContext) getSnowflakeOption(name string) snowflake.ID {
	option := i.GetOption(name)
	if option == nil || option.Value == nil {
		return 0
	}
	var value string
	switch raw := option.Value.(type) {
	case string:
		value = raw
	case json.Number:
		value = raw.String()
	case float64:
		value = strconv.FormatUint(uint64(raw), 10)
	default:
		value = fmt.Sprint(raw)
	}
	id, _ := snowflake.Parse(value)
	return id
}

// Reply sends a public response to the interaction.
func (i *InteractionContext) Reply(content string) error {
	return i.respond(&interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
		Data: &interactions.InteractionCallbackData{Content: content},
	})
}

// ReplyEphemeral sends a private response visible only to the invoker.
func (i *InteractionContext) ReplyEphemeral(content string) error {
	return i.respond(&interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
		Data: &interactions.InteractionCallbackData{Content: content, Flags: messages.FlagEphemeral},
	})
}

// ReplyEmbed sends one or more embeds as the interaction response.
func (i *InteractionContext) ReplyEmbed(embeds ...messages.Embed) error {
	return i.ReplyComplex(&interactions.InteractionCallbackData{Embeds: embeds})
}

// ReplyComplex sends a fully customized interaction response.
func (i *InteractionContext) ReplyComplex(data *interactions.InteractionCallbackData) error {
	if data == nil {
		return errors.New("bot: interaction response data is nil")
	}
	return i.respond(&interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
		Data: data,
	})
}

// ReplyEphemeralComplex sends a fully customized interaction response that is
// visible only to the invoker. This is the complex counterpart to
// ReplyEphemeral, allowing embeds and components in an ephemeral response.
func (i *InteractionContext) ReplyEphemeralComplex(data *interactions.InteractionCallbackData) error {
	if data == nil {
		return errors.New("bot: interaction response data is nil")
	}
	data.Flags |= messages.FlagEphemeral
	return i.respond(&interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
		Data: data,
	})
}

// ReplyComplexWithFiles sends an initial interaction response with attachments.
func (i *InteractionContext) ReplyComplexWithFiles(data *interactions.InteractionCallbackData, files ...rest.File) error {
	if data == nil {
		return errors.New("bot: interaction response data is nil")
	}
	i.responseMu.Lock()
	if i.responded {
		i.responseMu.Unlock()
		return ErrInteractionAlreadyResponded
	}
	i.responded = true
	i.deferred = false
	i.ephemeral = data.Flags&messages.FlagEphemeral != 0
	i.responseMu.Unlock()
	ctx, cancel := i.responseCtx()
	defer cancel()
	response := interactions.InteractionResponse{Type: interactions.InteractionCallbackTypeChannelMessageWithSource, Data: data}
	err := i.rest.CreateInteractionResponseWithFiles(ctx, i.ID, i.Token, response, files)
	if err == nil {
		return nil
	}
	i.responseMu.Lock()
	i.responded = false
	i.deferred = false
	i.ephemeral = false
	i.responseMu.Unlock()
	return err
}

// Defer acknowledges the interaction and shows a thinking indicator.
func (i *InteractionContext) Defer() error {
	return i.respond(&interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeDeferredChannelMessageWithSource,
	})
}

// DeferEphemeral defers with an ephemeral loading state.
func (i *InteractionContext) DeferEphemeral() error {
	return i.respond(&interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeDeferredChannelMessageWithSource,
		Data: &interactions.InteractionCallbackData{Flags: messages.FlagEphemeral},
	})
}

func (i *InteractionContext) respond(response *interactions.InteractionResponse) error {
	i.responseMu.Lock()
	if i.responded {
		i.responseMu.Unlock()
		return ErrInteractionAlreadyResponded
	}
	i.responded = true
	i.deferred = response.Type == interactions.InteractionCallbackTypeDeferredChannelMessageWithSource || response.Type == interactions.InteractionCallbackTypeDeferredUpdateMessage
	i.ephemeral = response.Data != nil && response.Data.Flags&messages.FlagEphemeral != 0
	i.responseMu.Unlock()

	ctx, cancel := i.responseCtx()
	defer cancel()
	err := i.rest.CreateInteractionResponse(ctx, i.ID, i.Token, *response)
	if err != nil {
		i.responseMu.Lock()
		i.responded = false
		i.deferred = false
		i.ephemeral = false
		i.responseMu.Unlock()
	}
	return err
}

func (i *InteractionContext) responseCtx() (context.Context, context.CancelFunc) {
	d := 3 * time.Second
	if i.Bot != nil && i.Bot.interactionTimeout > 0 {
		d = i.Bot.interactionTimeout
	}
	return context.WithTimeout(i.Context(), d)
}

// Update edits the message that triggered a component interaction.
func (i *InteractionContext) Update(data *interactions.InteractionCallbackData) error {
	if data == nil {
		return errors.New("bot: interaction update data is nil")
	}
	return i.respond(&interactions.InteractionResponse{Type: interactions.InteractionCallbackTypeUpdateMessage, Data: data})
}

// UpdateContent edits the triggering message's text.
func (i *InteractionContext) UpdateContent(content string) error {
	return i.Update(&interactions.InteractionCallbackData{Content: content})
}

// DeferUpdate acknowledges a component interaction without changing its message.
func (i *InteractionContext) DeferUpdate() error {
	return i.respond(&interactions.InteractionResponse{Type: interactions.InteractionCallbackTypeDeferredUpdateMessage})
}

// Pong acknowledges a ping interaction.
func (i *InteractionContext) Pong() error {
	return i.respond(&interactions.InteractionResponse{Type: interactions.InteractionCallbackTypePong})
}

// LaunchActivity invokes a Discord embedded activity callback.
func (i *InteractionContext) LaunchActivity(data *interactions.InteractionCallbackData) error {
	if data == nil {
		data = &interactions.InteractionCallbackData{}
	}
	return i.respond(&interactions.InteractionResponse{Type: interactions.InteractionCallbackTypeLaunchActivity, Data: data})
}

// Autocomplete responds with choices for the focused option.
func (i *InteractionContext) Autocomplete(choices ...interactions.ApplicationCommandOptionChoice) error {
	return i.respond(&interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeApplicationCommandAutocompleteResult,
		Data: &interactions.InteractionCallbackData{Choices: choices},
	})
}

// ShowModal displays a modal with the provided components.
func (i *InteractionContext) ShowModal(customID, title string, modalComponents ...components.Component) error {
	return i.ShowModalComplex(&interactions.InteractionCallbackData{CustomID: customID, Title: title, Components: modalComponents})
}

// ShowModalBuilder displays a modal built by components.ModalBuilder.
func (i *InteractionContext) ShowModalBuilder(modal components.ModalData) error {
	return i.ShowModal(modal.CustomID, modal.Title, modal.Components...)
}

// ShowModalComplex displays a fully customized modal response.
func (i *InteractionContext) ShowModalComplex(data *interactions.InteractionCallbackData) error {
	if data == nil {
		return errors.New("bot: modal data is nil")
	}
	return i.respond(&interactions.InteractionResponse{Type: interactions.InteractionCallbackTypeModal, Data: data})
}

// GetReply retrieves the original interaction response.
func (i *InteractionContext) GetReply() (*messages.Message, error) {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.GetOriginalInteractionResponse(ctx, i.ApplicationID, i.Token)
}

// EditReply edits the original interaction response.
func (i *InteractionContext) EditReply(content string) (*messages.Message, error) {
	return i.EditReplyComplex(rest.EditMessageParams{Content: &content})
}

// EditReplyText edits the original interaction response's text content,
// discarding the returned message. This is a convenience for handlers that
// only need to know whether the edit succeeded.
func (i *InteractionContext) EditReplyText(content string) error {
	_, err := i.EditReply(content)
	return err
}

// EditReplyComplex edits the original interaction response with full control.
func (i *InteractionContext) EditReplyComplex(params rest.EditMessageParams) (*messages.Message, error) {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.EditOriginalInteractionResponse(ctx, i.ApplicationID, i.Token, params)
}

// EditReplyWithFiles edits the original response and uploads attachments.
func (i *InteractionContext) EditReplyWithFiles(params rest.EditMessageParams, files ...rest.File) (*messages.Message, error) {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.EditOriginalInteractionResponseWithFiles(ctx, i.ApplicationID, i.Token, params, files)
}

// DeleteReply deletes the original interaction response.
func (i *InteractionContext) DeleteReply() error {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.DeleteOriginalInteractionResponse(ctx, i.ApplicationID, i.Token)
}

// Followup sends a follow-up message after responding or deferring.
func (i *InteractionContext) Followup(content string) (*messages.Message, error) {
	return i.FollowupComplex(rest.ExecuteWebhookParams{Content: content})
}

// FollowupEmbed sends one or more embeds as a follow-up.
func (i *InteractionContext) FollowupEmbed(embeds ...messages.Embed) (*messages.Message, error) {
	return i.FollowupComplex(rest.ExecuteWebhookParams{Embeds: embeds})
}

// FollowupEphemeral sends a private follow-up message.
func (i *InteractionContext) FollowupEphemeral(content string) (*messages.Message, error) {
	return i.FollowupComplex(rest.ExecuteWebhookParams{Content: content, Flags: messages.FlagEphemeral})
}

// FollowupComplex sends a fully customized follow-up message.
func (i *InteractionContext) FollowupComplex(params rest.ExecuteWebhookParams) (*messages.Message, error) {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.CreateFollowupMessage(ctx, i.ApplicationID, i.Token, params)
}

// FollowupComplexWithFiles sends a multipart follow-up message.
func (i *InteractionContext) FollowupComplexWithFiles(params rest.ExecuteWebhookParams, files ...rest.File) (*messages.Message, error) {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.CreateFollowupMessageWithFiles(ctx, i.ApplicationID, i.Token, params, files)
}

// GetFollowup retrieves a follow-up message by ID.
func (i *InteractionContext) GetFollowup(messageID snowflake.ID) (*messages.Message, error) {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.GetFollowupMessage(ctx, i.ApplicationID, i.Token, messageID)
}

// EditFollowup edits a follow-up message.
func (i *InteractionContext) EditFollowup(messageID snowflake.ID, params rest.EditMessageParams) (*messages.Message, error) {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.EditFollowupMessage(ctx, i.ApplicationID, i.Token, messageID, params)
}

// DeleteFollowup deletes a follow-up message.
func (i *InteractionContext) DeleteFollowup(messageID snowflake.ID) error {
	ctx, cancel := i.newCtx()
	defer cancel()
	return i.rest.DeleteFollowupMessage(ctx, i.ApplicationID, i.Token, messageID)
}

// HasResponded reports whether an initial response or deferral was accepted.
func (i *InteractionContext) HasResponded() bool {
	i.responseMu.RLock()
	defer i.responseMu.RUnlock()
	return i.responded
}

// Replied reports whether an initial non-deferred response was sent.
func (i *InteractionContext) Replied() bool {
	i.responseMu.RLock()
	defer i.responseMu.RUnlock()
	return i.responded && !i.deferred
}

// Deferred reports whether the interaction was acknowledged with a defer.
func (i *InteractionContext) Deferred() bool {
	i.responseMu.RLock()
	defer i.responseMu.RUnlock()
	return i.deferred
}

// Ephemeral reports whether the initial response was marked ephemeral.
func (i *InteractionContext) Ephemeral() bool {
	i.responseMu.RLock()
	defer i.responseMu.RUnlock()
	return i.ephemeral
}

// ReactionContext wraps a MESSAGE_REACTION_ADD event.
type ReactionContext struct {
	BaseContext
	*events.MessageReactionAdd
}

// ChannelID returns the channel ID where the reaction occurred.
// This is a method for consistency with InteractionContext.ChannelID()
// and MessageContext.ChannelID().
func (r *ReactionContext) ChannelID() snowflake.ID {
	return r.MessageReactionAdd.ChannelID
}

// Reply sends a text message to the channel where the reaction occurred.
func (r *ReactionContext) Reply(content string) (*messages.Message, error) {
	ctx, cancel := r.newCtx()
	defer cancel()
	return r.rest.CreateMessage(ctx, r.ChannelID(), content)
}

// GuildContext wraps a GUILD_CREATE event.
type GuildContext struct {
	BaseContext
	*events.GuildCreate
}

// GuildUpdateContext wraps a GUILD_UPDATE event.
type GuildUpdateContext struct {
	BaseContext
	*events.GuildUpdate
}

// GuildDeleteContext wraps a GUILD_DELETE event.
type GuildDeleteContext struct {
	BaseContext
	*events.GuildDelete
}

// ChannelContext wraps a CHANNEL_CREATE event.
type ChannelContext struct {
	BaseContext
	*events.ChannelCreate
}

// ChannelUpdateContext wraps a CHANNEL_UPDATE event.
type ChannelUpdateContext struct {
	BaseContext
	*events.ChannelUpdate
}

// GuildAuditLogEntryContext wraps a GUILD_AUDIT_LOG_ENTRY_CREATE event.
type GuildAuditLogEntryContext struct {
	BaseContext
	*events.GuildAuditLogEntryCreate
}

func findOption(options []interactions.ApplicationCommandInteractionDataOption, name string) *interactions.ApplicationCommandInteractionDataOption {
	for index := range options {
		if options[index].Name == name {
			return &options[index]
		}
		if nested := findOption(options[index].Options, name); nested != nil {
			return nested
		}
	}
	return nil
}

func firstOptionOfType(options []interactions.ApplicationCommandInteractionDataOption, typ interactions.ApplicationCommandOptionType) *interactions.ApplicationCommandInteractionDataOption {
	for index := range options {
		if options[index].Type == typ {
			return &options[index]
		}
		if nested := firstOptionOfType(options[index].Options, typ); nested != nil {
			return nested
		}
	}
	return nil
}

func findFocusedOption(options []interactions.ApplicationCommandInteractionDataOption) *interactions.ApplicationCommandInteractionDataOption {
	for index := range options {
		if options[index].Focused {
			return &options[index]
		}
		if nested := findFocusedOption(options[index].Options); nested != nil {
			return nested
		}
	}
	return nil
}

func (i *InteractionContext) modalComponents() []modalComponent {
	if i.data == nil {
		return nil
	}
	var result []modalComponent
	var walk func([]modalComponent)
	walk = func(components []modalComponent) {
		for _, component := range components {
			if component.CustomID != "" {
				result = append(result, component)
			}
			walk(component.Components)
		}
	}
	walk(i.data.Components)
	return result
}

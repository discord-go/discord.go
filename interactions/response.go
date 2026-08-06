package interactions

import (
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
)

// InteractionCallbackType is the type of interaction response.
type InteractionCallbackType int

const (
	InteractionCallbackTypePong                                 InteractionCallbackType = 1
	InteractionCallbackTypeChannelMessageWithSource             InteractionCallbackType = 4
	InteractionCallbackTypeDeferredChannelMessageWithSource     InteractionCallbackType = 5
	InteractionCallbackTypeDeferredUpdateMessage                InteractionCallbackType = 6
	InteractionCallbackTypeUpdateMessage                        InteractionCallbackType = 7
	InteractionCallbackTypeApplicationCommandAutocompleteResult InteractionCallbackType = 8
	InteractionCallbackTypeModal                                InteractionCallbackType = 9
	InteractionCallbackTypePremiumRequired                      InteractionCallbackType = 10
	InteractionCallbackTypeLaunchActivity                       InteractionCallbackType = 12
)

// ApplicationCommandOptionChoice represents a choice for an application command option.
type ApplicationCommandOptionChoice struct {
	Name              string            `json:"name"`
	NameLocalizations map[string]string `json:"name_localizations,omitempty"`
	Value             interface{}       `json:"value"`
}

// InteractionCallbackData is the data for the interaction response.
type InteractionCallbackData struct {
	TTS             bool                             `json:"tts,omitempty"`
	Content         string                           `json:"content,omitempty"`
	Embeds          []messages.Embed                 `json:"embeds,omitempty"`
	AllowedMentions *messages.AllowedMentions        `json:"allowed_mentions,omitempty"`
	Flags           int                              `json:"flags,omitempty"`
	Components      []components.Component           `json:"components,omitempty"`
	Attachments     []messages.Attachment            `json:"attachments,omitempty"`
	Choices         []ApplicationCommandOptionChoice `json:"choices,omitempty"`
	CustomID        string                           `json:"custom_id,omitempty"`
	Title           string                           `json:"title,omitempty"`
	ThreadName      string                           `json:"thread_name,omitempty"`
	AppliedTags     []string                         `json:"applied_tags,omitempty"`
	Poll            *messages.Poll                   `json:"poll,omitempty"`
}

// InteractionResponse represents a response to an interaction.
type InteractionResponse struct {
	Type InteractionCallbackType  `json:"type"`
	Data *InteractionCallbackData `json:"data,omitempty"`
}

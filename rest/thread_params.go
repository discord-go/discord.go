package rest

import "github.com/discord-go/discord.go/messages"

// StartThreadParams represents the parameters for starting a thread.
type StartThreadParams struct {
	Name                string                `json:"name"`
	AutoArchiveDuration int                   `json:"auto_archive_duration,omitempty"`
	Type                int                   `json:"type,omitempty"`
	Invitable           *bool                 `json:"invitable,omitempty"`
	RateLimitPerUser    *int                  `json:"rate_limit_per_user,omitempty"`
	AppliedTags         []string              `json:"applied_tags,omitempty"`
	Message             *messages.MessageSend `json:"message,omitempty"`
}

// StartThreadWithMessageParams represents the parameters for starting a thread from a message.
type StartThreadWithMessageParams struct {
	Name                string                `json:"name"`
	AutoArchiveDuration int                   `json:"auto_archive_duration,omitempty"`
	RateLimitPerUser    *int                  `json:"rate_limit_per_user,omitempty"`
	AppliedTags         []string              `json:"applied_tags,omitempty"`
	Message             *messages.MessageSend `json:"message,omitempty"`
}

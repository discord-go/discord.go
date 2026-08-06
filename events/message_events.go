package events

import (
	"encoding/json"

	"github.com/discord-go/discord.go/emojis"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

// MessageCreate represents the MESSAGE_CREATE event.
type MessageCreate struct {
	messages.Message
	GuildID snowflake.ID `json:"guild_id,string,omitempty"`
}

func (mc *MessageCreate) UnmarshalJSON(data []byte) error {
	if err := mc.Message.UnmarshalJSON(data); err != nil {
		return err
	}
	var aux struct {
		GuildID snowflake.ID `json:"guild_id,string,omitempty"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	mc.GuildID = aux.GuildID
	return nil
}

// MessageUpdate represents the MESSAGE_UPDATE event.
type MessageUpdate struct {
	messages.Message
}

// MessageDelete represents the MESSAGE_DELETE event.
type MessageDelete struct {
	ID        snowflake.ID `json:"id,string"`
	ChannelID snowflake.ID `json:"channel_id,string"`
	GuildID   snowflake.ID `json:"guild_id,string,omitempty"`
}

// MessageReactionAdd represents the MESSAGE_REACTION_ADD event.
type MessageReactionAdd struct {
	UserID    snowflake.ID `json:"user_id,string"`
	ChannelID snowflake.ID `json:"channel_id,string"`
	MessageID snowflake.ID `json:"message_id,string"`
	GuildID   snowflake.ID `json:"guild_id,string,omitempty"`
	Emoji     emojis.Emoji `json:"emoji"`
}

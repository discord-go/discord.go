package guilds

import (
	"encoding/json"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// Widget represents a Discord guild widget.
type Widget struct {
	ID            snowflake.ID    `json:"id,string"`
	Name          string          `json:"name"`
	InstantInvite *string         `json:"instant_invite"`
	Channels      []WidgetChannel `json:"channels"`
	Members       []users.User    `json:"members"`
	PresenceCount int             `json:"presence_count"`
}

// WidgetChannel represents a channel in a guild widget.
type WidgetChannel struct {
	ID       snowflake.ID `json:"id,string"`
	Name     string       `json:"name"`
	Position int          `json:"position"`
}

// WidgetSettings represents the widget settings of a Discord guild.
type WidgetSettings struct {
	Enabled   bool          `json:"enabled"`
	ChannelID *snowflake.ID `json:"channel_id,string"`
}

// UnmarshalJSON unmarshals a Widget.
func (w *Widget) UnmarshalJSON(data []byte) error {
	type Alias Widget
	aux := &struct {
		ID string `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(w),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.ID != "" {
		id, err := snowflake.Parse(aux.ID)
		if err != nil {
			return err
		}
		w.ID = id
	}
	return nil
}

// UnmarshalJSON unmarshals a WidgetChannel.
func (wc *WidgetChannel) UnmarshalJSON(data []byte) error {
	type Alias WidgetChannel
	aux := &struct {
		ID string `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(wc),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.ID != "" {
		id, err := snowflake.Parse(aux.ID)
		if err != nil {
			return err
		}
		wc.ID = id
	}
	return nil
}

// UnmarshalJSON unmarshals WidgetSettings.
func (ws *WidgetSettings) UnmarshalJSON(data []byte) error {
	type Alias WidgetSettings
	aux := &struct {
		ChannelID *string `json:"channel_id"`
		*Alias
	}{
		Alias: (*Alias)(ws),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.ChannelID != nil && *aux.ChannelID != "" {
		id, err := snowflake.Parse(*aux.ChannelID)
		if err != nil {
			return err
		}
		ws.ChannelID = &id
	}
	return nil
}

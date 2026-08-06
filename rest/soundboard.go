package rest

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

type SoundboardSound struct {
	Name      string       `json:"name"`
	SoundID   snowflake.ID `json:"sound_id,string"`
	Volume    float64      `json:"volume"`
	EmojiID   snowflake.ID `json:"emoji_id,string,omitempty"`
	EmojiName string       `json:"emoji_name,omitempty"`
	GuildID   snowflake.ID `json:"guild_id,string,omitempty"`
	Available bool         `json:"available"`
	User      *users.User  `json:"user,omitempty"`
}

type CreateSoundboardSoundParams struct {
	Name      string       `json:"name"`
	Sound     []byte       `json:"-"` // Assuming raw bytes for file
	Volume    float64      `json:"volume,omitempty"`
	EmojiID   snowflake.ID `json:"emoji_id,string,omitempty"`
	EmojiName string       `json:"emoji_name,omitempty"`
}

func (p CreateSoundboardSoundParams) MarshalJSON() ([]byte, error) {
	type payload struct {
		Name      string       `json:"name"`
		Sound     string       `json:"sound,omitempty"`
		Volume    float64      `json:"volume,omitempty"`
		EmojiID   snowflake.ID `json:"emoji_id,string,omitempty"`
		EmojiName string       `json:"emoji_name,omitempty"`
	}
	value := payload{Name: p.Name, Volume: p.Volume, EmojiID: p.EmojiID, EmojiName: p.EmojiName}
	if len(p.Sound) > 0 {
		value.Sound = "data:audio/ogg;base64," + base64.StdEncoding.EncodeToString(p.Sound)
	}
	return json.Marshal(value)
}

type soundboardList struct {
	Items []SoundboardSound `json:"items"`
}

func (s *soundboardList) UnmarshalJSON(data []byte) error {
	var items []SoundboardSound
	if len(data) > 0 && data[0] == '[' {
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		s.Items = items
		return nil
	}
	var payload struct {
		Items []SoundboardSound `json:"items"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	s.Items = payload.Items
	return nil
}

type ModifySoundboardSoundParams struct {
	Name      *string       `json:"name,omitempty"`
	Volume    *float64      `json:"volume,omitempty"`
	EmojiID   *snowflake.ID `json:"emoji_id,string,omitempty"`
	EmojiName *string       `json:"emoji_name,omitempty"`
}

func (c *Client) ListDefaultSoundboardSounds(ctx context.Context) ([]SoundboardSound, error) {
	var payload soundboardList
	err := c.Request(ctx, "GET", "/soundboard-default-sounds", nil, &payload)
	return payload.Items, err
}

func (c *Client) ListGuildSoundboardSounds(ctx context.Context, guildID snowflake.ID) ([]SoundboardSound, error) {
	var payload soundboardList
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/soundboard-sounds", nil, &payload)
	return payload.Items, err
}

func (c *Client) GetGuildSoundboardSound(ctx context.Context, guildID snowflake.ID, soundID snowflake.ID) (*SoundboardSound, error) {
	var sound SoundboardSound
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/soundboard-sounds/"+soundID.String(), nil, &sound)
	if err != nil {
		return nil, err
	}
	return &sound, nil
}

func (c *Client) CreateGuildSoundboardSound(ctx context.Context, guildID snowflake.ID, params CreateSoundboardSoundParams) (*SoundboardSound, error) {
	var sound SoundboardSound
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/soundboard-sounds", params, &sound)
	if err != nil {
		return nil, err
	}
	return &sound, nil
}

func (c *Client) ModifyGuildSoundboardSound(ctx context.Context, guildID snowflake.ID, soundID snowflake.ID, params ModifySoundboardSoundParams) (*SoundboardSound, error) {
	var sound SoundboardSound
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/soundboard-sounds/"+soundID.String(), params, &sound)
	if err != nil {
		return nil, err
	}
	return &sound, nil
}

func (c *Client) DeleteGuildSoundboardSound(ctx context.Context, guildID snowflake.ID, soundID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/soundboard-sounds/"+soundID.String(), nil, nil)
}

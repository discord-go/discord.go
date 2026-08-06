package guilds

import (
	"encoding/json"
	"time"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// Template represents a Discord guild template.
type Template struct {
	Code                  string       `json:"code"`
	Name                  string       `json:"name"`
	Description           *string      `json:"description"`
	UsageCount            int          `json:"usage_count"`
	CreatorID             snowflake.ID `json:"creator_id,string"`
	Creator               users.User   `json:"creator"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
	SourceGuildID         snowflake.ID `json:"source_guild_id,string"`
	SerializedSourceGuild Guild        `json:"serialized_source_guild"`
	IsDirty               *bool        `json:"is_dirty"`
}

// UnmarshalJSON unmarshals a Template from JSON.
func (t *Template) UnmarshalJSON(data []byte) error {
	type Alias Template
	aux := &struct {
		CreatorID     string `json:"creator_id"`
		SourceGuildID string `json:"source_guild_id"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.CreatorID != "" {
		id, err := snowflake.Parse(aux.CreatorID)
		if err != nil {
			return err
		}
		t.CreatorID = id
	}
	if aux.SourceGuildID != "" {
		id, err := snowflake.Parse(aux.SourceGuildID)
		if err != nil {
			return err
		}
		t.SourceGuildID = id
	}

	return nil
}

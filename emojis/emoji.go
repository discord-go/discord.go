package emojis

import (
	"encoding/json"

	"github.com/discord-go/discord.go/roles"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// Emoji represents a custom emoji in Discord.
// https://discord.com/developers/docs/resources/emoji#emoji-object
type Emoji struct {
	ID            snowflake.ID `json:"id,string"`
	Name          *string      `json:"name"`
	Roles         []roles.Role `json:"roles,omitempty"`
	User          *users.User  `json:"user,omitempty"`
	RequireColons bool         `json:"require_colons,omitempty"`
	Managed       bool         `json:"managed,omitempty"`
	Animated      bool         `json:"animated,omitempty"`
	Available     bool         `json:"available,omitempty"`
}

// UnmarshalJSON unmarshals an emoji from JSON, parsing roles properly.
func (e *Emoji) UnmarshalJSON(data []byte) error {
	type Alias Emoji
	aux := &struct {
		Roles []string `json:"roles"`
		ID    *string  `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.ID != nil {
		id, err := snowflake.Parse(*aux.ID)
		if err != nil {
			return err
		}
		e.ID = id
	}
	if aux.Roles != nil {
		e.Roles = make([]roles.Role, len(aux.Roles))
		for i, r := range aux.Roles {
			id, err := snowflake.Parse(r)
			if err != nil {
				return err
			}
			e.Roles[i] = roles.Role{ID: id}
		}
	}
	return nil
}

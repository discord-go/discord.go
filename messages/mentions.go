package messages

import (
	"encoding/json"

	"github.com/discord-go/discord.go/snowflake"
)

// AllowedMentionType represents the type of allowed mentions.
type AllowedMentionType string

const (
	AllowedMentionTypeRole     AllowedMentionType = "roles"
	AllowedMentionTypeUser     AllowedMentionType = "users"
	AllowedMentionTypeEveryone AllowedMentionType = "everyone"
)

// AllowedMentions represents the allowed mentions object.
type AllowedMentions struct {
	Parse       []AllowedMentionType `json:"parse,omitempty"`
	Roles       []snowflake.ID       `json:"roles,omitempty"`
	Users       []snowflake.ID       `json:"users,omitempty"`
	RepliedUser bool                 `json:"replied_user,omitempty"`
}

// UnmarshalJSON unmarshals the AllowedMentions and handles string arrays properly.
func (a *AllowedMentions) UnmarshalJSON(data []byte) error {
	type Alias AllowedMentions
	aux := &struct {
		Roles []string `json:"roles"`
		Users []string `json:"users"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Roles != nil {
		a.Roles = make([]snowflake.ID, len(aux.Roles))
		for i, r := range aux.Roles {
			id, err := snowflake.Parse(r)
			if err != nil {
				return err
			}
			a.Roles[i] = id
		}
	}

	if aux.Users != nil {
		a.Users = make([]snowflake.ID, len(aux.Users))
		for i, u := range aux.Users {
			id, err := snowflake.Parse(u)
			if err != nil {
				return err
			}
			a.Users[i] = id
		}
	}

	return nil
}

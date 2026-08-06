package users

import (
	"encoding/json"
	"time"

	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
)

// Member represents a Discord guild member.
type Member struct {
	User                       *User                  `json:"user,omitempty"`
	Nick                       *string                `json:"nick,omitempty"`
	Avatar                     *string                `json:"avatar,omitempty"`
	AvatarDecorationData       *AvatarDecorationData  `json:"avatar_decoration_data,omitempty"`
	Banner                     *string                `json:"banner,omitempty"`
	Roles                      []snowflake.ID         `json:"roles"`
	JoinedAt                   time.Time              `json:"joined_at"`
	PremiumSince               *time.Time             `json:"premium_since,omitempty"`
	Deaf                       bool                   `json:"deaf"`
	Mute                       bool                   `json:"mute"`
	Pending                    bool                   `json:"pending,omitempty"`
	Permissions                permissions.Permission `json:"permissions,string,omitempty"`
	Flags                      int                    `json:"flags,omitempty"`
	CommunicationDisabledUntil *time.Time             `json:"communication_disabled_until,omitempty"`
}

// UnmarshalJSON unmarshals a member from JSON, parsing roles properly.
func (m *Member) UnmarshalJSON(data []byte) error {
	type Alias Member
	aux := &struct {
		Roles []string `json:"roles"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Roles != nil {
		m.Roles = make([]snowflake.ID, len(aux.Roles))
		for i, r := range aux.Roles {
			id, err := snowflake.Parse(r)
			if err != nil {
				return err
			}
			m.Roles[i] = id
		}
	}
	return nil
}

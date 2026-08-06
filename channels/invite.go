package channels

import (
	"time"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

type Invite struct {
	Code                     string       `json:"code"`
	Guild                    *InviteGuild `json:"guild,omitempty"`
	Channel                  *Channel     `json:"channel"`
	Inviter                  *users.User  `json:"inviter,omitempty"`
	TargetType               int          `json:"target_type,omitempty"`
	TargetUser               *users.User  `json:"target_user,omitempty"`
	TargetApplication        *Application `json:"target_application,omitempty"`
	ApproximatePresenceCount int          `json:"approximate_presence_count,omitempty"`
	ApproximateMemberCount   int          `json:"approximate_member_count,omitempty"`
	ExpiresAt                *time.Time   `json:"expires_at,omitempty"`
}

type InviteGuild struct {
	ID                snowflake.ID `json:"id,string"`
	Name              string       `json:"name"`
	Splash            *string      `json:"splash"`
	Banner            *string      `json:"banner"`
	Description       *string      `json:"description"`
	Icon              *string      `json:"icon"`
	Features          []string     `json:"features"`
	VerificationLevel int          `json:"verification_level"`
	VanityURLCode     *string      `json:"vanity_url_code"`
}

type Application struct {
	ID          snowflake.ID `json:"id,string"`
	Name        string       `json:"name"`
	Icon        *string      `json:"icon"`
	Description string       `json:"description"`
}

package application

import (
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// MembershipState represents the membership state of a team member.
type MembershipState int

const (
	// MembershipStateInvited indicates the user is invited to the team.
	MembershipStateInvited MembershipState = 1
	// MembershipStateAccepted indicates the user has accepted the invitation to the team.
	MembershipStateAccepted MembershipState = 2
)

// Team represents a Discord team.
type Team struct {
	Icon        *string      `json:"icon"`
	ID          snowflake.ID `json:"id,string"`
	Members     []TeamMember `json:"members"`
	Name        string       `json:"name"`
	OwnerUserID snowflake.ID `json:"owner_user_id,string"`
}

// TeamMember represents a member of a Discord team.
type TeamMember struct {
	MembershipState MembershipState `json:"membership_state"`
	Permissions     []string        `json:"permissions"`
	TeamID          snowflake.ID    `json:"team_id,string"`
	User            users.User      `json:"user"`
	Role            string          `json:"role"`
}

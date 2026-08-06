package users

import "github.com/discord-go/discord.go/snowflake"

// ClientStatus represents the client status of a user.
type ClientStatus struct {
	Desktop string `json:"desktop,omitempty"`
	Mobile  string `json:"mobile,omitempty"`
	Web     string `json:"web,omitempty"`
}

// Activity represents a user's activity.
type Activity struct {
	Name          string       `json:"name"`
	Type          int          `json:"type"`
	URL           *string      `json:"url,omitempty"`
	CreatedAt     int64        `json:"created_at"`
	ApplicationID snowflake.ID `json:"application_id,string,omitempty"`
	Details       *string      `json:"details,omitempty"`
	State         *string      `json:"state,omitempty"`
}

// PresenceUpdate represents a presence update.
type PresenceUpdate struct {
	User         User         `json:"user"`
	GuildID      snowflake.ID `json:"guild_id,string"`
	Status       string       `json:"status"`
	Activities   []Activity   `json:"activities"`
	ClientStatus ClientStatus `json:"client_status"`
}

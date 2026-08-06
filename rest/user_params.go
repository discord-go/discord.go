package rest

import "github.com/discord-go/discord.go/snowflake"

// ModifyUserParams contains the parameters for modifying the current user.
type ModifyUserParams struct {
	Username string  `json:"username,omitempty"`
	Avatar   *string `json:"avatar,omitempty"`
	Banner   *string `json:"banner,omitempty"`
}

// ListGuildsParams contains the parameters for listing the current user's guilds.
type ListGuildsParams struct {
	Before     snowflake.ID `json:"-"`
	After      snowflake.ID `json:"-"`
	Limit      int          `json:"-"`
	WithCounts bool         `json:"-"`
}

// QueryString builds the query string for ListGuildsParams.
func (p ListGuildsParams) QueryString() string {
	q := ""
	sep := "?"
	if p.Before != 0 {
		q += sep + "before=" + p.Before.String()
		sep = "&"
	}
	if p.After != 0 {
		q += sep + "after=" + p.After.String()
		sep = "&"
	}
	if p.Limit > 0 {
		q += sep + "limit=" + itoa(p.Limit)
		sep = "&"
	}
	if p.WithCounts {
		q += sep + "with_counts=true"
	}
	return q
}

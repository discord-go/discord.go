package guilds

import (
	"encoding/json"
	"time"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// Integration represents a Discord guild integration.
type Integration struct {
	ID                snowflake.ID            `json:"id,string"`
	Name              string                  `json:"name"`
	Type              string                  `json:"type"`
	Enabled           bool                    `json:"enabled"`
	Syncing           *bool                   `json:"syncing,omitempty"`
	RoleID            *snowflake.ID           `json:"role_id,string"`
	EnableEmoticons   *bool                   `json:"enable_emoticons,omitempty"`
	ExpireBehavior    *int                    `json:"expire_behavior,omitempty"`
	ExpireGracePeriod *int                    `json:"expire_grace_period,omitempty"`
	User              *users.User             `json:"user,omitempty"`
	Account           IntegrationAccount      `json:"account"`
	SyncedAt          *time.Time              `json:"synced_at,omitempty"`
	SubscriberCount   *int                    `json:"subscriber_count,omitempty"`
	Revoked           *bool                   `json:"revoked,omitempty"`
	Application       *IntegrationApplication `json:"application,omitempty"`
	Scopes            []string                `json:"scopes,omitempty"`
}

// IntegrationApplication represents the application of a Discord guild integration.
type IntegrationApplication struct {
	ID          snowflake.ID `json:"id,string"`
	Name        string       `json:"name"`
	Icon        *string      `json:"icon"`
	Description string       `json:"description"`
	Bot         *users.User  `json:"bot,omitempty"`
}

// IntegrationAccount represents a third-party account link for a Discord integration.
type IntegrationAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UnmarshalJSON unmarshals Integration.
func (i *Integration) UnmarshalJSON(data []byte) error {
	type Alias Integration
	aux := &struct {
		ID     string  `json:"id"`
		RoleID *string `json:"role_id"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ID != "" {
		id, err := snowflake.Parse(aux.ID)
		if err != nil {
			return err
		}
		i.ID = id
	}
	if aux.RoleID != nil && *aux.RoleID != "" {
		id, err := snowflake.Parse(*aux.RoleID)
		if err != nil {
			return err
		}
		i.RoleID = &id
	}

	return nil
}

// UnmarshalJSON unmarshals IntegrationApplication.
func (ia *IntegrationApplication) UnmarshalJSON(data []byte) error {
	type Alias IntegrationApplication
	aux := &struct {
		ID string `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(ia),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.ID != "" {
		id, err := snowflake.Parse(aux.ID)
		if err != nil {
			return err
		}
		ia.ID = id
	}
	return nil
}

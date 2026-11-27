package application

import (
	"encoding/json"
	"strconv"

	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// Application represents a Discord application.
type Application struct {
	ID                                snowflake.ID                     `json:"id,string"`
	Name                              string                           `json:"name"`
	Icon                              *string                          `json:"icon"`
	IconHash                          *string                          `json:"icon_hash,omitempty"`
	Description                       string                           `json:"description"`
	RPCOrigins                        []string                         `json:"rpc_origins,omitempty"`
	BotPublic                         bool                             `json:"bot_public"`
	BotRequireCodeGrant               bool                             `json:"bot_require_code_grant"`
	Bot                               *users.User                      `json:"bot,omitempty"`
	TermsOfServiceURL                 string                           `json:"terms_of_service_url,omitempty"`
	PrivacyPolicyURL                  string                           `json:"privacy_policy_url,omitempty"`
	Owner                             *users.User                      `json:"owner,omitempty"`
	VerifyKey                         string                           `json:"verify_key"`
	Team                              *Team                            `json:"team"`
	GuildID                           snowflake.ID                     `json:"guild_id,string,omitempty"`
	PrimarySKUID                      snowflake.ID                     `json:"primary_sku_id,string,omitempty"`
	Slug                              string                           `json:"slug,omitempty"`
	CoverImage                        *string                          `json:"cover_image,omitempty"`
	Flags                             int                              `json:"flags,omitempty"`
	ApproximateGuildCount             int                              `json:"approximate_guild_count,omitempty"`
	RedirectURIs                      []string                         `json:"redirect_uris,omitempty"`
	InteractionsEndpointURL           string                           `json:"interactions_endpoint_url,omitempty"`
	RoleConnectionsVerificationURL    string                           `json:"role_connections_verification_url,omitempty"`
	Tags                              []string                         `json:"tags,omitempty"`
	CustomInstallURL                  string                           `json:"custom_install_url,omitempty"`
	ApproximateUserInstallCount       int                              `json:"approximate_user_install_count,omitempty"`
	ApproximateUserAuthorizationCount int                              `json:"approximate_user_authorization_count,omitempty"`
	IntegrationTypes                  []int                            `json:"integration_types,omitempty"`
	InstallParams                     *InstallParams                   `json:"install_params,omitempty"`
	ExplicitContentFilter             int                              `json:"explicit_content_filter,omitempty"`
	Summary                           string                           `json:"summary,omitempty"`
	Guild                             *struct{}                        `json:"guild,omitempty"`
	FlagsNew                          int                              `json:"flags_new,omitempty"` // Discord string bitfield, limited by int width.
	IntegrationTypesConfig            map[string]IntegrationTypeConfig `json:"integration_types_config,omitempty"`
	EventWebhooksURL                  string                           `json:"event_webhooks_url,omitempty"`
	EventWebhooksStatus               int                              `json:"event_webhooks_status,omitempty"`
	// EventWebhooksTypes contains legacy numeric event values.
	// Deprecated: Discord sends event type names; use EventWebhookTypeNames.
	EventWebhooksTypes []int `json:"event_webhooks_types,omitempty"`
	// EventWebhookTypeNames contains the string event types sent by Discord.
	EventWebhookTypeNames []string `json:"-"`
}

// UnmarshalJSON accepts Discord's string-encoded flags_new as well as the
// numeric representation accepted by earlier versions of this package.
func (a *Application) UnmarshalJSON(data []byte) error {
	type alias Application
	aux := struct {
		FlagsNew     json.RawMessage `json:"flags_new"`
		WebhookTypes json.RawMessage `json:"event_webhooks_types"`
		*alias
	}{alias: (*alias)(a)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.WebhookTypes) > 0 {
		if string(aux.WebhookTypes) == "null" {
			a.EventWebhookTypeNames = nil
			a.EventWebhooksTypes = nil
		} else {
			var names []string
			if err := json.Unmarshal(aux.WebhookTypes, &names); err == nil {
				a.EventWebhookTypeNames = names
				a.EventWebhooksTypes = nil
				if len(names) == 0 {
					a.EventWebhooksTypes = []int{}
				}
			} else {
				var legacy []int
				if err := json.Unmarshal(aux.WebhookTypes, &legacy); err != nil {
					return err
				}
				a.EventWebhooksTypes = legacy
				a.EventWebhookTypeNames = nil
			}
		}
	}
	if len(aux.FlagsNew) == 0 || string(aux.FlagsNew) == "null" {
		return nil
	}
	if aux.FlagsNew[0] == '"' {
		var value string
		if err := json.Unmarshal(aux.FlagsNew, &value); err != nil {
			return err
		}
		flags, err := strconv.ParseInt(value, 10, strconv.IntSize)
		if err != nil {
			return err
		}
		a.FlagsNew = int(flags)
		return nil
	}
	return json.Unmarshal(aux.FlagsNew, &a.FlagsNew)
}

// MarshalJSON retains legacy numeric webhook values while preferring string
// event names when they are provided.
func (a Application) MarshalJSON() ([]byte, error) {
	type alias Application
	if a.EventWebhookTypeNames == nil {
		return json.Marshal(alias(a))
	}
	return json.Marshal(struct {
		WebhookTypes []string `json:"event_webhooks_types,omitempty"`
		*alias
	}{WebhookTypes: a.EventWebhookTypeNames, alias: (*alias)(&a)})
}

type InstallParams struct {
	Scopes      []string               `json:"scopes"`
	Permissions permissions.Permission `json:"permissions,string"`
}

type IntegrationTypeConfig struct {
	OAuth2InstallParams *InstallParams `json:"oauth2_install_params,omitempty"`
}

type ActivityLocation struct {
	Kind      string `json:"kind"`
	GuildID   string `json:"guild_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
}

type ActivityInstance struct {
	ID            string             `json:"id"`
	ApplicationID snowflake.ID       `json:"application_id,string"`
	Activity      map[string]any     `json:"activity,omitempty"`
	Participants  []users.User       `json:"participants,omitempty"`
	Locations     []ActivityLocation `json:"locations,omitempty"`
}

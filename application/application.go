package application

import (
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
	FlagsNew                          int                              `json:"flags_new,omitempty"`
	IntegrationTypesConfig            map[string]IntegrationTypeConfig `json:"integration_types_config,omitempty"`
	EventWebhooksURL                  string                           `json:"event_webhooks_url,omitempty"`
	EventWebhooksStatus               int                              `json:"event_webhooks_status,omitempty"`
	EventWebhooksTypes                []int                            `json:"event_webhooks_types,omitempty"`
}

type InstallParams struct {
	Scopes      []string `json:"scopes"`
	Permissions string   `json:"permissions"`
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

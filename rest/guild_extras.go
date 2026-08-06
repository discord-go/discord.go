package rest

import (
	"context"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

type ModifyWelcomeScreenParams struct {
	Enabled         *bool                          `json:"enabled,omitempty"`
	WelcomeChannels *[]guilds.WelcomeScreenChannel `json:"welcome_channels,omitempty"`
	Description     *string                        `json:"description,omitempty"`
}

type ModifyOnboardingParams struct {
	Prompts           []guilds.OnboardingPrompt `json:"prompts,omitempty"`
	DefaultChannelIDs snowflake.IDs             `json:"default_channel_ids,omitempty"`
	Enabled           *bool                     `json:"enabled,omitempty"`
	Mode              *int                      `json:"mode,omitempty"`
}

func (c *Client) GetGuildPreview(ctx context.Context, guildID snowflake.ID) (*guilds.GuildPreview, error) {
	var preview guilds.GuildPreview
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/preview", nil, &preview)
	if err != nil {
		return nil, err
	}
	return &preview, nil
}

func (c *Client) GetGuildVanityURL(ctx context.Context, guildID snowflake.ID) (*guilds.VanityURL, error) {
	var vanity guilds.VanityURL
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/vanity-url", nil, &vanity)
	if err != nil {
		return nil, err
	}
	return &vanity, nil
}

func (c *Client) GetGuildWelcomeScreen(ctx context.Context, guildID snowflake.ID) (*guilds.WelcomeScreen, error) {
	var screen guilds.WelcomeScreen
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/welcome-screen", nil, &screen)
	if err != nil {
		return nil, err
	}
	return &screen, nil
}

func (c *Client) ModifyGuildWelcomeScreen(ctx context.Context, guildID snowflake.ID, params ModifyWelcomeScreenParams) (*guilds.WelcomeScreen, error) {
	var screen guilds.WelcomeScreen
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/welcome-screen", params, &screen)
	if err != nil {
		return nil, err
	}
	return &screen, nil
}

func (c *Client) GetGuildOnboarding(ctx context.Context, guildID snowflake.ID) (*guilds.Onboarding, error) {
	var onboarding guilds.Onboarding
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/onboarding", nil, &onboarding)
	if err != nil {
		return nil, err
	}
	return &onboarding, nil
}

func (c *Client) ModifyGuildOnboarding(ctx context.Context, guildID snowflake.ID, params ModifyOnboardingParams) (*guilds.Onboarding, error) {
	var onboarding guilds.Onboarding
	err := c.Request(ctx, "PUT", "/guilds/"+guildID.String()+"/onboarding", params, &onboarding)
	if err != nil {
		return nil, err
	}
	return &onboarding, nil
}

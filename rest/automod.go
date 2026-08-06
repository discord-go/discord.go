package rest

import (
	"context"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

// ListAutoModerationRules lists all auto moderation rules for a guild.
func (c *Client) ListAutoModerationRules(ctx context.Context, guildID snowflake.ID) ([]guilds.AutoModerationRule, error) {
	var rules []guilds.AutoModerationRule
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/auto-moderation/rules", nil, &rules)
	return rules, err
}

// GetAutoModerationRule gets a single auto moderation rule.
func (c *Client) GetAutoModerationRule(ctx context.Context, guildID, ruleID snowflake.ID) (*guilds.AutoModerationRule, error) {
	var rule guilds.AutoModerationRule
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/auto-moderation/rules/"+ruleID.String(), nil, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreateAutoModerationRule creates a new auto moderation rule.
func (c *Client) CreateAutoModerationRule(ctx context.Context, guildID snowflake.ID, params CreateAutoModerationRuleParams) (*guilds.AutoModerationRule, error) {
	var rule guilds.AutoModerationRule
	err := c.Request(ctx, "POST", "/guilds/"+guildID.String()+"/auto-moderation/rules", params, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// ModifyAutoModerationRule modifies an existing auto moderation rule.
func (c *Client) ModifyAutoModerationRule(ctx context.Context, guildID, ruleID snowflake.ID, params ModifyAutoModerationRuleParams) (*guilds.AutoModerationRule, error) {
	var rule guilds.AutoModerationRule
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/auto-moderation/rules/"+ruleID.String(), params, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// DeleteAutoModerationRule deletes an auto moderation rule.
func (c *Client) DeleteAutoModerationRule(ctx context.Context, guildID, ruleID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/auto-moderation/rules/"+ruleID.String(), nil, nil)
}

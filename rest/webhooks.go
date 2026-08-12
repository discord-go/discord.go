package rest

import (
	"context"
	"net/url"

	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/webhook"
)

func (c *Client) CreateWebhook(ctx context.Context, channelID snowflake.ID, params CreateWebhookParams) (*webhook.Webhook, error) {
	var wh webhook.Webhook
	err := c.Request(ctx, "POST", "/channels/"+channelID.String()+"/webhooks", params, &wh)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

func (c *Client) ListChannelWebhooks(ctx context.Context, channelID snowflake.ID) ([]webhook.Webhook, error) {
	var webhooks []webhook.Webhook
	err := c.Request(ctx, "GET", "/channels/"+channelID.String()+"/webhooks", nil, &webhooks)
	return webhooks, err
}

func (c *Client) ListGuildWebhooks(ctx context.Context, guildID snowflake.ID) ([]webhook.Webhook, error) {
	var webhooks []webhook.Webhook
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/webhooks", nil, &webhooks)
	return webhooks, err
}

func (c *Client) GetWebhookWithToken(ctx context.Context, webhookID snowflake.ID, token string) (*webhook.Webhook, error) {
	var wh webhook.Webhook
	err := c.RequestNoAuth(ctx, "GET", "/webhooks/"+webhookID.String()+"/"+token, nil, &wh)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

func (c *Client) ModifyWebhookWithToken(ctx context.Context, webhookID snowflake.ID, token string, params ModifyWebhookParams) (*webhook.Webhook, error) {
	var wh webhook.Webhook
	err := c.RequestNoAuth(ctx, "PATCH", "/webhooks/"+webhookID.String()+"/"+token, params, &wh)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

func (c *Client) DeleteWebhookWithToken(ctx context.Context, webhookID snowflake.ID, token string) error {
	return c.RequestNoAuth(ctx, "DELETE", "/webhooks/"+webhookID.String()+"/"+token, nil, nil)
}

func (c *Client) GetWebhook(ctx context.Context, webhookID snowflake.ID) (*webhook.Webhook, error) {
	var wh webhook.Webhook
	err := c.Request(ctx, "GET", "/webhooks/"+webhookID.String(), nil, &wh)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

func (c *Client) ModifyWebhook(ctx context.Context, webhookID snowflake.ID, params ModifyWebhookParams) (*webhook.Webhook, error) {
	var wh webhook.Webhook
	err := c.Request(ctx, "PATCH", "/webhooks/"+webhookID.String(), params, &wh)
	if err != nil {
		return nil, err
	}
	return &wh, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, webhookID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/webhooks/"+webhookID.String(), nil, nil)
}

func (c *Client) ExecuteWebhook(ctx context.Context, webhookID snowflake.ID, webhookToken string, params ExecuteWebhookParams) (*messages.Message, error) {
	return c.ExecuteWebhookWithOptions(ctx, webhookID, webhookToken, params, ExecuteWebhookOptions{Wait: true})
}

type ExecuteWebhookOptions struct {
	Wait           bool
	ThreadID       snowflake.ID
	WithComponents bool
}

func (o ExecuteWebhookOptions) query() string {
	values := url.Values{}
	if o.Wait {
		values.Set("wait", "true")
	}
	if o.ThreadID != 0 {
		values.Set("thread_id", o.ThreadID.String())
	}
	if o.WithComponents {
		values.Set("with_components", "true")
	}
	encoded := values.Encode()
	if encoded == "" {
		return ""
	}
	return "?" + encoded
}

func (c *Client) ExecuteWebhookWithOptions(ctx context.Context, webhookID snowflake.ID, webhookToken string, params ExecuteWebhookParams, options ExecuteWebhookOptions) (*messages.Message, error) {
	var msg messages.Message
	path := "/webhooks/" + webhookID.String() + "/" + webhookToken + options.query()
	var target any = &msg
	if !options.Wait {
		target = nil
	}
	err := c.RequestNoAuth(ctx, "POST", path, params, target)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) ExecuteWebhookWithFiles(ctx context.Context, webhookID snowflake.ID, webhookToken string, params ExecuteWebhookParams, files []File, options ExecuteWebhookOptions) (*messages.Message, error) {
	params.Attachments = append(params.Attachments, AttachmentMetadata(files)...)
	var msg messages.Message
	var target any = &msg
	if !options.Wait {
		target = nil
	}
	err := c.RequestMultipartNoAuth(ctx, "POST", "/webhooks/"+webhookID.String()+"/"+webhookToken+options.query(), params, files, target)
	if err != nil {
		return nil, err
	}
	if !options.Wait {
		return nil, nil
	}
	return &msg, nil
}

// EditWebhookMessage edits a previously-sent webhook message from the same token.
func (c *Client) EditWebhookMessage(ctx context.Context, webhookID snowflake.ID, webhookToken string, messageID snowflake.ID, params EditMessageParams) (*messages.Message, error) {
	return c.EditWebhookMessageWithOptions(ctx, webhookID, webhookToken, messageID, params, ExecuteWebhookOptions{Wait: true})
}

func (c *Client) EditWebhookMessageWithOptions(ctx context.Context, webhookID snowflake.ID, webhookToken string, messageID snowflake.ID, params EditMessageParams, options ExecuteWebhookOptions) (*messages.Message, error) {
	var msg messages.Message
	query := url.Values{}
	if options.ThreadID != 0 {
		query.Set("thread_id", options.ThreadID.String())
	}
	if options.WithComponents {
		query.Set("with_components", "true")
	}
	path := "/webhooks/" + webhookID.String() + "/" + webhookToken + "/messages/" + messageID.String()
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	err := c.RequestNoAuth(ctx, "PATCH", path, params, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) GetWebhookMessage(ctx context.Context, webhookID snowflake.ID, webhookToken string, messageID snowflake.ID, threadID snowflake.ID) (*messages.Message, error) {
	values := url.Values{}
	if threadID != 0 {
		values.Set("thread_id", threadID.String())
	}
	path := "/webhooks/" + webhookID.String() + "/" + webhookToken + "/messages/" + messageID.String()
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var msg messages.Message
	err := c.RequestNoAuth(ctx, "GET", path, nil, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// DeleteWebhookMessage deletes a message that was created by the webhook.
func (c *Client) DeleteWebhookMessage(ctx context.Context, webhookID snowflake.ID, webhookToken string, messageID snowflake.ID) error {
	return c.DeleteWebhookMessageWithOptions(ctx, webhookID, webhookToken, messageID, 0)
}

func (c *Client) DeleteWebhookMessageWithOptions(ctx context.Context, webhookID snowflake.ID, webhookToken string, messageID, threadID snowflake.ID) error {
	values := url.Values{}
	if threadID != 0 {
		values.Set("thread_id", threadID.String())
	}
	path := "/webhooks/" + webhookID.String() + "/" + webhookToken + "/messages/" + messageID.String()
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return c.RequestNoAuth(ctx, "DELETE", path, nil, nil)
}

func (c *Client) ExecuteSlackWebhook(ctx context.Context, webhookID snowflake.ID, webhookToken string, payload any, wait bool) error {
	path := "/webhooks/" + webhookID.String() + "/" + webhookToken + "/slack"
	if wait {
		path += "?wait=true"
	}
	return c.RequestNoAuth(ctx, "POST", path, payload, nil)
}

func (c *Client) ExecuteSlackWebhookWithOptions(ctx context.Context, webhookID snowflake.ID, webhookToken string, payload any, options ExecuteWebhookOptions) error {
	path := "/webhooks/" + webhookID.String() + "/" + webhookToken + "/slack" + options.query()
	return c.RequestNoAuth(ctx, "POST", path, payload, nil)
}

func (c *Client) ExecuteGitHubWebhook(ctx context.Context, webhookID snowflake.ID, webhookToken string, payload any, wait bool) error {
	path := "/webhooks/" + webhookID.String() + "/" + webhookToken + "/github"
	if wait {
		path += "?wait=true"
	}
	return c.RequestNoAuth(ctx, "POST", path, payload, nil)
}

func (c *Client) ExecuteGitHubWebhookWithOptions(ctx context.Context, webhookID snowflake.ID, webhookToken string, payload any, options ExecuteWebhookOptions) error {
	path := "/webhooks/" + webhookID.String() + "/" + webhookToken + "/github" + options.query()
	return c.RequestNoAuth(ctx, "POST", path, payload, nil)
}

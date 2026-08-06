package rest

import (
	"context"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

func (c *Client) CreateInteractionResponse(ctx context.Context, interactionID snowflake.ID, interactionToken string, response interactions.InteractionResponse) error {
	return c.RequestNoAuth(ctx, "POST", "/interactions/"+interactionID.String()+"/"+interactionToken+"/callback", response, nil)
}

// CreateInteractionResponseWithFiles sends an initial callback as multipart.
func (c *Client) CreateInteractionResponseWithFiles(ctx context.Context, interactionID snowflake.ID, interactionToken string, response interactions.InteractionResponse, files []File) error {
	response = withAttachmentMetadata(response, files)
	return c.RequestMultipartNoAuth(ctx, "POST", "/interactions/"+interactionID.String()+"/"+interactionToken+"/callback", response, files, nil)
}

// CreateInteractionResponseWithResult requests Discord's callback response
// resource using with_response=true.
func (c *Client) CreateInteractionResponseWithResult(ctx context.Context, interactionID snowflake.ID, interactionToken string, response interactions.InteractionResponse) (map[string]any, error) {
	var result map[string]any
	err := c.RequestNoAuth(ctx, "POST", "/interactions/"+interactionID.String()+"/"+interactionToken+"/callback?with_response=true", response, &result)
	return result, err
}

func (c *Client) GetOriginalInteractionResponse(ctx context.Context, applicationID snowflake.ID, interactionToken string) (*messages.Message, error) {
	var msg messages.Message
	err := c.RequestNoAuth(ctx, "GET", "/webhooks/"+applicationID.String()+"/"+interactionToken+"/messages/@original", nil, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) EditOriginalInteractionResponse(ctx context.Context, applicationID snowflake.ID, interactionToken string, params EditMessageParams) (*messages.Message, error) {
	var msg messages.Message
	err := c.RequestNoAuth(ctx, "PATCH", "/webhooks/"+applicationID.String()+"/"+interactionToken+"/messages/@original", params, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// EditOriginalInteractionResponseWithFiles edits the original response as multipart.
func (c *Client) EditOriginalInteractionResponseWithFiles(ctx context.Context, applicationID snowflake.ID, interactionToken string, params EditMessageParams, files []File) (*messages.Message, error) {
	params = withEditAttachmentMetadata(params, files)
	var msg messages.Message
	err := c.RequestMultipartNoAuth(ctx, "PATCH", "/webhooks/"+applicationID.String()+"/"+interactionToken+"/messages/@original", params, files, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *Client) DeleteOriginalInteractionResponse(ctx context.Context, applicationID snowflake.ID, interactionToken string) error {
	return c.RequestNoAuth(ctx, "DELETE", "/webhooks/"+applicationID.String()+"/"+interactionToken+"/messages/@original", nil, nil)
}

// CreateFollowupMessage creates a followup message for an interaction.
func (c *Client) CreateFollowupMessage(ctx context.Context, applicationID snowflake.ID, interactionToken string, params ExecuteWebhookParams) (*messages.Message, error) {
	var msg messages.Message
	err := c.RequestNoAuth(ctx, "POST", "/webhooks/"+applicationID.String()+"/"+interactionToken, params, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// CreateFollowupMessageWithFiles sends a multipart follow-up message.
func (c *Client) CreateFollowupMessageWithFiles(ctx context.Context, applicationID snowflake.ID, interactionToken string, params ExecuteWebhookParams, files []File) (*messages.Message, error) {
	params.Attachments = append(params.Attachments, AttachmentMetadata(files)...)
	var msg messages.Message
	err := c.RequestMultipartNoAuth(ctx, "POST", "/webhooks/"+applicationID.String()+"/"+interactionToken, params, files, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetFollowupMessage gets a followup message for an interaction.
func (c *Client) GetFollowupMessage(ctx context.Context, applicationID snowflake.ID, interactionToken string, messageID snowflake.ID) (*messages.Message, error) {
	var msg messages.Message
	err := c.RequestNoAuth(ctx, "GET", "/webhooks/"+applicationID.String()+"/"+interactionToken+"/messages/"+messageID.String(), nil, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// EditFollowupMessage edits a followup message for an interaction.
func (c *Client) EditFollowupMessage(ctx context.Context, applicationID snowflake.ID, interactionToken string, messageID snowflake.ID, params EditMessageParams) (*messages.Message, error) {
	var msg messages.Message
	err := c.RequestNoAuth(ctx, "PATCH", "/webhooks/"+applicationID.String()+"/"+interactionToken+"/messages/"+messageID.String(), params, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// DeleteFollowupMessage deletes a followup message for an interaction.
func (c *Client) DeleteFollowupMessage(ctx context.Context, applicationID snowflake.ID, interactionToken string, messageID snowflake.ID) error {
	return c.RequestNoAuth(ctx, "DELETE", "/webhooks/"+applicationID.String()+"/"+interactionToken+"/messages/"+messageID.String(), nil, nil)
}

func withAttachmentMetadata(response interactions.InteractionResponse, files []File) interactions.InteractionResponse {
	if response.Data == nil || len(files) == 0 {
		return response
	}
	data := *response.Data
	data.Attachments = append(append([]messages.Attachment(nil), data.Attachments...), AttachmentMetadata(files)...)
	response.Data = &data
	return response
}

func withEditAttachmentMetadata(params EditMessageParams, files []File) EditMessageParams {
	if len(files) == 0 {
		return params
	}
	attachments := AttachmentMetadata(files)
	if params.Attachments != nil {
		attachments = append(append([]messages.Attachment(nil), (*params.Attachments)...), attachments...)
	}
	params.Attachments = &attachments
	return params
}

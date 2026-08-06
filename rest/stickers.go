package rest

import (
	"context"

	"github.com/discord-go/discord.go/emojis"
	"github.com/discord-go/discord.go/snowflake"
)

func (c *Client) GetSticker(ctx context.Context, stickerID snowflake.ID) (*emojis.Sticker, error) {
	var sticker emojis.Sticker
	err := c.Request(ctx, "GET", "/stickers/"+stickerID.String(), nil, &sticker)
	if err != nil {
		return nil, err
	}
	return &sticker, nil
}

func (c *Client) ListStickerPacks(ctx context.Context) ([]emojis.StickerPack, error) {
	var packs struct {
		StickerPacks []emojis.StickerPack `json:"sticker_packs"`
	}
	err := c.Request(ctx, "GET", "/sticker-packs", nil, &packs)
	return packs.StickerPacks, err
}

type Sticker struct {
	ID          snowflake.ID `json:"id,string"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tags        string       `json:"tags"`
	Type        int          `json:"type"`
	FormatType  int          `json:"format_type"`
	Available   bool         `json:"available"`
	GuildID     snowflake.ID `json:"guild_id,string,omitempty"`
}

type CreateStickerParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	File        []byte `json:"-"` // Assuming raw bytes for file or a proper multipart structure in Request.
}

type ModifyStickerParams struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Tags        *string `json:"tags,omitempty"`
}

func (c *Client) ListGuildStickers(ctx context.Context, guildID snowflake.ID) ([]Sticker, error) {
	var stickers []Sticker
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/stickers", nil, &stickers)
	return stickers, err
}

func (c *Client) GetGuildSticker(ctx context.Context, guildID snowflake.ID, stickerID snowflake.ID) (*Sticker, error) {
	var sticker Sticker
	err := c.Request(ctx, "GET", "/guilds/"+guildID.String()+"/stickers/"+stickerID.String(), nil, &sticker)
	if err != nil {
		return nil, err
	}
	return &sticker, nil
}

func (c *Client) CreateGuildSticker(ctx context.Context, guildID snowflake.ID, params CreateStickerParams) (*Sticker, error) {
	var sticker Sticker
	file := FileFromBytes("sticker", params.File)
	err := c.RequestMultipartFormNamedFile(ctx, "POST", "/guilds/"+guildID.String()+"/stickers", map[string]string{
		"name": params.Name, "description": params.Description, "tags": params.Tags,
	}, "file", []File{file}, &sticker)
	if err != nil {
		return nil, err
	}
	return &sticker, nil
}

func (c *Client) ModifyGuildSticker(ctx context.Context, guildID snowflake.ID, stickerID snowflake.ID, params ModifyStickerParams) (*Sticker, error) {
	var sticker Sticker
	err := c.Request(ctx, "PATCH", "/guilds/"+guildID.String()+"/stickers/"+stickerID.String(), params, &sticker)
	if err != nil {
		return nil, err
	}
	return &sticker, nil
}

func (c *Client) DeleteGuildSticker(ctx context.Context, guildID snowflake.ID, stickerID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/guilds/"+guildID.String()+"/stickers/"+stickerID.String(), nil, nil)
}

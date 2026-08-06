package emojis

import (
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// StickerType represents the type of a sticker.
type StickerType int

const (
	StickerTypeStandard StickerType = 1
	StickerTypeGuild    StickerType = 2
)

// StickerFormatType represents the format of a sticker.
type StickerFormatType int

const (
	StickerFormatTypePNG    StickerFormatType = 1
	StickerFormatTypeAPNG   StickerFormatType = 2
	StickerFormatTypeLottie StickerFormatType = 3
	StickerFormatTypeGIF    StickerFormatType = 4
)

// Sticker represents a Discord sticker.
// https://discord.com/developers/docs/resources/sticker#sticker-object
type Sticker struct {
	ID          snowflake.ID      `json:"id,string"`
	PackID      snowflake.ID      `json:"pack_id,string,omitempty"`
	Name        string            `json:"name"`
	Description *string           `json:"description"`
	Tags        string            `json:"tags"`
	Type        StickerType       `json:"type"`
	FormatType  StickerFormatType `json:"format_type"`
	Available   bool              `json:"available,omitempty"`
	GuildID     snowflake.ID      `json:"guild_id,string,omitempty"`
	User        *users.User       `json:"user,omitempty"`
	SortValue   *int              `json:"sort_value,omitempty"`
}

// StickerItem represents a sticker item within a message.
// https://discord.com/developers/docs/resources/sticker#sticker-item-object
type StickerItem struct {
	ID         snowflake.ID      `json:"id,string"`
	Name       string            `json:"name"`
	FormatType StickerFormatType `json:"format_type"`
}

type StickerPack struct {
	ID          snowflake.ID `json:"id,string"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Stickers    []Sticker    `json:"stickers"`
	SKU         snowflake.ID `json:"sku_id,string"`
	BannerAsset *string      `json:"banner_asset_id,string,omitempty"`
	CoverAsset  *string      `json:"cover_sticker_id,string,omitempty"`
}

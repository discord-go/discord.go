package messages

import (
	"time"
)

// EmbedBuilder provides a fluent interface for constructing Embeds.
type EmbedBuilder struct {
	embed Embed
}

// NewEmbedBuilder creates a new EmbedBuilder.
func NewEmbedBuilder() *EmbedBuilder {
	return &EmbedBuilder{
		embed: Embed{},
	}
}

// SetTitle sets the title of the embed.
func (b *EmbedBuilder) SetTitle(title string) *EmbedBuilder {
	b.embed.Title = title
	return b
}

// SetDescription sets the description of the embed.
func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder {
	b.embed.Description = desc
	return b
}

// SetURL sets the URL of the embed.
func (b *EmbedBuilder) SetURL(url string) *EmbedBuilder {
	b.embed.URL = url
	return b
}

// SetTimestamp sets the timestamp of the embed.
func (b *EmbedBuilder) SetTimestamp(t time.Time) *EmbedBuilder {
	b.embed.Timestamp = t.Format(time.RFC3339)
	return b
}

// SetColor sets the color of the embed.
func (b *EmbedBuilder) SetColor(color int) *EmbedBuilder {
	b.embed.Color = color
	return b
}

// SetFooter sets the footer of the embed.
func (b *EmbedBuilder) SetFooter(text, iconURL string) *EmbedBuilder {
	b.embed.Footer = &EmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return b
}

// SetImage sets the image of the embed.
func (b *EmbedBuilder) SetImage(url string) *EmbedBuilder {
	b.embed.Image = &EmbedImage{
		URL: url,
	}
	return b
}

// SetThumbnail sets the thumbnail of the embed.
func (b *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder {
	b.embed.Thumbnail = &EmbedImage{
		URL: url,
	}
	return b
}

// SetAuthor sets the author of the embed.
func (b *EmbedBuilder) SetAuthor(name, url, iconURL string) *EmbedBuilder {
	b.embed.Author = &EmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return b
}

// AddField adds a field to the embed.
func (b *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder {
	b.embed.Fields = append(b.embed.Fields, EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return b
}

// Build returns the constructed Embed.
func (b *EmbedBuilder) Build() Embed {
	return b.embed
}

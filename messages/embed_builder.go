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

// SetFooter sets the footer of the embed. The icon URL is optional; pass an
// empty string to omit it.
func (b *EmbedBuilder) SetFooter(text, iconURL string) *EmbedBuilder {
	b.embed.Footer = &EmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return b
}

// SetFooterText sets the footer text of the embed without an icon.
func (b *EmbedBuilder) SetFooterText(text string) *EmbedBuilder {
	b.embed.Footer = &EmbedFooter{Text: text}
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

// SetAuthor sets the author of the embed. The URL and icon URL are optional;
// pass empty strings to omit them.
func (b *EmbedBuilder) SetAuthor(name, url, iconURL string) *EmbedBuilder {
	b.embed.Author = &EmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return b
}

// SetAuthorName sets the author name of the embed without a URL or icon.
func (b *EmbedBuilder) SetAuthorName(name string) *EmbedBuilder {
	b.embed.Author = &EmbedAuthor{Name: name}
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

// AddInlineField adds a field to the embed with inline set to true.
func (b *EmbedBuilder) AddInlineField(name, value string) *EmbedBuilder {
	return b.AddField(name, value, true)
}

// ClearFields removes all fields from the embed.
func (b *EmbedBuilder) ClearFields() *EmbedBuilder {
	b.embed.Fields = nil
	return b
}

// SetTimestampNow sets the embed timestamp to the current time.
func (b *EmbedBuilder) SetTimestampNow() *EmbedBuilder {
	return b.SetTimestamp(time.Now())
}

// SetColorHex parses a hex color string such as "#5865F2" or "0x5865F2" and
// sets the embed color. Invalid input leaves the color unchanged and returns
// the builder for chaining.
func (b *EmbedBuilder) SetColorHex(hex string) *EmbedBuilder {
	if color, ok := parseHexColor(hex); ok {
		b.embed.Color = color
	}
	return b
}

// Build returns the constructed Embed.
func (b *EmbedBuilder) Build() Embed {
	return b.embed
}

// parseHexColor converts "#RRGGBB" or "0xRRGGBB" into an RGB integer.
func parseHexColor(hex string) (int, bool) {
	if len(hex) == 0 {
		return 0, false
	}
	if hex[0] == '#' {
		hex = hex[1:]
	} else if len(hex) > 2 && hex[0] == '0' && (hex[1] == 'x' || hex[1] == 'X') {
		hex = hex[2:]
	}
	if len(hex) != 6 {
		return 0, false
	}
	var color int
	for _, c := range hex {
		color <<= 4
		switch {
		case c >= '0' && c <= '9':
			color |= int(c - '0')
		case c >= 'a' && c <= 'f':
			color |= int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			color |= int(c-'A') + 10
		default:
			return 0, false
		}
	}
	return color, true
}

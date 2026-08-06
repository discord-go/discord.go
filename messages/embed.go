package messages

import (
	"fmt"
)

// Embed represents a Discord embed.
type Embed struct {
	Title       string         `json:"title,omitempty"`
	Type        string         `json:"type,omitempty"`
	Description string         `json:"description,omitempty"`
	URL         string         `json:"url,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
	Color       int            `json:"color,omitempty"`
	Footer      *EmbedFooter   `json:"footer,omitempty"`
	Image       *EmbedImage    `json:"image,omitempty"`
	Thumbnail   *EmbedImage    `json:"thumbnail,omitempty"`
	Video       *EmbedVideo    `json:"video,omitempty"`
	Provider    *EmbedProvider `json:"provider,omitempty"`
	Author      *EmbedAuthor   `json:"author,omitempty"`
	Fields      []EmbedField   `json:"fields,omitempty"`
}

// EmbedFooter represents the footer of an embed.
type EmbedFooter struct {
	Text         string `json:"text"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedImage represents the image of an embed.
type EmbedImage struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// EmbedVideo represents the video of an embed.
type EmbedVideo struct {
	URL      string `json:"url,omitempty"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// EmbedProvider represents the provider of an embed.
type EmbedProvider struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// EmbedAuthor represents the author of an embed.
type EmbedAuthor struct {
	Name         string `json:"name"`
	URL          string `json:"url,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedField represents a field of an embed.
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// Validate ensures the embed constraints are met before sending to Discord.
func (e *Embed) Validate() error {
	totalChars := 0

	if len(e.Title) > 256 {
		return fmt.Errorf("embed title exceeds 256 characters")
	}
	totalChars += len(e.Title)

	if len(e.Description) > 4096 {
		return fmt.Errorf("embed description exceeds 4096 characters")
	}
	totalChars += len(e.Description)

	if e.Footer != nil {
		if len(e.Footer.Text) > 2048 {
			return fmt.Errorf("embed footer text exceeds 2048 characters")
		}
		totalChars += len(e.Footer.Text)
	}

	if e.Author != nil {
		if len(e.Author.Name) > 256 {
			return fmt.Errorf("embed author name exceeds 256 characters")
		}
		totalChars += len(e.Author.Name)
	}

	if len(e.Fields) > 25 {
		return fmt.Errorf("embed fields exceed 25")
	}

	for _, field := range e.Fields {
		if len(field.Name) > 256 {
			return fmt.Errorf("embed field name exceeds 256 characters")
		}
		if len(field.Value) > 1024 {
			return fmt.Errorf("embed field value exceeds 1024 characters")
		}
		totalChars += len(field.Name) + len(field.Value)
	}

	if totalChars > 6000 {
		return fmt.Errorf("embed total characters exceed 6000")
	}

	return nil
}

package messages

import (
	"github.com/discord-go/discord.go/snowflake"
)

// Attachment represents a Discord message attachment.
type Attachment struct {
	ID           snowflake.ID `json:"id,string"`
	Filename     string       `json:"filename"`
	Title        string       `json:"title,omitempty"`
	Description  string       `json:"description,omitempty"`
	ContentType  string       `json:"content_type,omitempty"`
	Size         int          `json:"size"`
	URL          string       `json:"url"`
	ProxyURL     string       `json:"proxy_url"`
	Height       int          `json:"height,omitempty"`
	Width        int          `json:"width,omitempty"`
	Ephemeral    bool         `json:"ephemeral,omitempty"`
	DurationSecs float64      `json:"duration_secs,omitempty"`
	Waveform     string       `json:"waveform,omitempty"`
	Flags        int          `json:"flags,omitempty"`
}

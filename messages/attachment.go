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

// AttachmentParams is Discord's attachment request structure: what a Message
// Create or Edit (and the webhook and interaction equivalents) may send,
// as opposed to the full Attachment object Discord returns. An existing
// attachment is referenced with ID alone, a new multipart upload carries the
// file index as ID plus Filename, and the remaining fields are optional
// metadata. Spoiler state is requested with IsSpoiler rather than the
// response-only attachment flag.
type AttachmentParams struct {
	ID           snowflake.ID `json:"id,string"`
	Filename     string       `json:"filename,omitempty"`
	Title        string       `json:"title,omitempty"`
	Description  string       `json:"description,omitempty"`
	DurationSecs float64      `json:"duration_secs,omitempty"`
	Waveform     string       `json:"waveform,omitempty"`
	IsSpoiler    bool         `json:"is_spoiler,omitempty"`
}

// NewAttachmentParams describes a multipart upload: id is the zero-based
// index of the file in the request, and name is the file name Discord shows.
func NewAttachmentParams(index int, name string) AttachmentParams {
	return AttachmentParams{ID: snowflake.ID(index), Filename: name}
}

package rest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

// MaxMessagesPerRequest is the Discord API limit for a single
// GetChannelMessages call.
const MaxMessagesPerRequest = 100

// FetchAllMessages retrieves every message in a channel by paginating
// backwards through the message history using the `before` cursor. Messages
// are returned in chronological order (oldest first).
//
// The max parameter caps the total number of messages fetched. Pass 0 or a
// negative value for no cap — the function will fetch until the channel's
// history is exhausted.
//
// The ctx is checked between each page request, so a cancelled context
// aborts the loop promptly.
func (c *Client) FetchAllMessages(ctx context.Context, channelID snowflake.ID, max int) ([]messages.Message, error) {
	var all []messages.Message
	before := (*snowflake.ID)(nil)

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if max > 0 && len(all) >= max {
			break
		}

		limit := MaxMessagesPerRequest
		if max > 0 {
			remaining := max - len(all)
			if remaining < limit {
				limit = remaining
			}
		}

		page, err := c.GetChannelMessages(ctx, channelID, GetMessagesParams{
			Before: before,
			Limit:  &limit,
		})
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}

		// Discord returns messages newest-first; accumulate as-is and
		// reverse once at the end for chronological order.
		all = append(all, page...)

		if len(page) < limit {
			break
		}
		before = &page[len(page)-1].ID
	}

	// Reverse to chronological order (oldest first).
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}

	// Trim to max if the last page exceeded it.
	if max > 0 && len(all) > max {
		all = all[:max]
	}

	return all, nil
}

// TranscriptOptions controls how a text transcript is formatted.
type TranscriptOptions struct {
	// ChannelName appears in the transcript header. If empty, the channel
	// ID is used.
	ChannelName string
	// ChannelID is included in the header metadata.
	ChannelID snowflake.ID
	// GeneratedAt overrides the timestamp in the header. If zero,
	// time.Now() is used.
	GeneratedAt time.Time
}

// FormatTranscript formats a slice of messages into a human-readable plain
// text transcript. Each message is rendered as:
//
//	[timestamp] Username (userID): content
//
// Attachments and embeds are noted on subsequent lines. Messages are
// expected to be in chronological order (oldest first), as produced by
// FetchAllMessages.
func FormatTranscript(msgs []messages.Message, opts TranscriptOptions) string {
	var b strings.Builder

	generatedAt := opts.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}

	channelName := opts.ChannelName
	if channelName == "" {
		channelName = opts.ChannelID.String()
	}

	fmt.Fprintf(&b, "Transcript for #%s\n", channelName)
	fmt.Fprintf(&b, "Channel ID: %s\n", opts.ChannelID.String())
	fmt.Fprintf(&b, "Generated: %s\n", generatedAt.UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "Messages: %d\n", len(msgs))
	b.WriteString("----------------------------------------\n\n")

	for _, msg := range msgs {
		timestamp := msg.Timestamp.UTC().Format("2006-01-02 15:04:05")

		authorName := "Unknown"
		authorID := ""
		if msg.Author != nil {
			authorName = msg.Author.Username
			if msg.Author.GlobalName != nil && *msg.Author.GlobalName != "" {
				authorName = *msg.Author.GlobalName
			}
			authorID = msg.Author.ID.String()
		}

		fmt.Fprintf(&b, "[%s] %s (%s):", timestamp, authorName, authorID)
		if msg.Content != "" {
			b.WriteString(" ")
			b.WriteString(msg.Content)
		}
		b.WriteString("\n")

		for _, attachment := range msg.Attachments {
			fmt.Fprintf(&b, "  + Attachment: %s (%s)\n", attachment.Filename, attachment.URL)
		}

		if len(msg.Embeds) > 0 {
			for i, embed := range msg.Embeds {
				title := embed.Title
				if title == "" {
					title = fmt.Sprintf("embed #%d", i+1)
				}
				fmt.Fprintf(&b, "  + Embed: %s\n", title)
			}
		}

		b.WriteString("\n")
	}

	return b.String()
}

// GenerateTranscript fetches all messages in a channel, formats them into a
// text transcript, and returns the result as a File suitable for uploading
// via CreateMessageComplexWithFiles or similar multipart endpoints.
//
// The max parameter caps the total number of messages fetched. Pass 0 or a
// negative value for no cap.
//
// The filename controls the attachment name. If empty, "transcript.txt" is
// used.
func (c *Client) GenerateTranscript(ctx context.Context, channelID snowflake.ID, max int, filename string, opts TranscriptOptions) (File, error) {
	if filename == "" {
		filename = "transcript.txt"
	}

	msgs, err := c.FetchAllMessages(ctx, channelID, max)
	if err != nil {
		return File{}, err
	}

	opts.ChannelID = channelID
	content := FormatTranscript(msgs, opts)

	return FileFromBytes(filename, []byte(content)), nil
}

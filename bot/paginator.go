package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
)

// Paginator manages a paginated message with button navigation.
// Each page is rendered by a user-provided function that returns the
// content and embeds to display.
type Paginator struct {
	bot       *Bot
	channelID snowflake.ID
	userID    snowflake.ID // optional: restrict to one user
	pages     []PaginatorPage
	timeout   time.Duration
}

// PaginatorPage holds the content for a single page.
type PaginatorPage struct {
	Content string
	Embeds  []messages.Embed
}

// PaginatorOption configures a Paginator.
type PaginatorOption func(*Paginator)

// WithPaginatorTimeout sets how long the paginator stays interactive
// before its buttons are removed. Defaults to 5 minutes.
func WithPaginatorTimeout(d time.Duration) PaginatorOption {
	return func(p *Paginator) { p.timeout = d }
}

// WithPaginatorUser restricts button interaction to a single user.
func WithPaginatorUser(userID snowflake.ID) PaginatorOption {
	return func(p *Paginator) { p.userID = userID }
}

const (
	paginatorPrevID = "paginator:prev"
	paginatorNextID = "paginator:next"
	paginatorStopID = "paginator:stop"
)

// NewPaginator creates a Paginator bound to a channel. Call Send to
// create the initial message and start listening for button presses.
func (b *Bot) NewPaginator(channelID snowflake.ID, pages []PaginatorPage, opts ...PaginatorOption) *Paginator {
	p := &Paginator{
		bot:       b,
		channelID: channelID,
		pages:     pages,
		timeout:   5 * time.Minute,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Send creates the initial message and starts the interaction loop.
// It blocks until the timeout expires, the user clicks stop, or ctx
// is cancelled. The message is edited in-place as the user navigates.
func (p *Paginator) Send(ctx context.Context) error {
	if len(p.pages) == 0 {
		return fmt.Errorf("paginator: no pages")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	current := 0
	msg, err := p.bot.Rest.CreateMessageComplex(ctx, p.channelID, p.renderPage(current, false))
	if err != nil {
		return fmt.Errorf("paginator: failed to send message: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	for {
		interaction, err := p.bot.AwaitInteraction(timeoutCtx, func(ic *InteractionContext) bool {
			if ic.Message == nil || ic.Message.ID != msg.ID {
				return false
			}
			if p.userID != 0 {
				if ic.Member != nil && ic.Member.User != nil && ic.Member.User.ID != p.userID {
					return false
				}
				if u := ic.User(); u != nil && u.ID != p.userID {
					return false
				}
			}
			customID := ic.CustomID()
			return customID == paginatorPrevID || customID == paginatorNextID || customID == paginatorStopID
		})
		if err != nil {
			break
		}

		customID := interaction.CustomID()

		// Acknowledge the interaction to prevent "interaction failed".
		_, _ = p.bot.Rest.EditOriginalInteractionResponse(timeoutCtx, p.bot.appID, interaction.Token, rest.EditMessageParams{})

		switch customID {
		case paginatorPrevID:
			if current > 0 {
				current--
			}
		case paginatorNextID:
			if current < len(p.pages)-1 {
				current++
			}
		case paginatorStopID:
			_, _ = p.bot.Rest.EditMessage(timeoutCtx, p.channelID, msg.ID, p.toEditParams(current, true))
			return nil
		}

		_, _ = p.bot.Rest.EditMessage(timeoutCtx, p.channelID, msg.ID, p.toEditParams(current, false))
	}

	// Remove buttons on timeout.
	_, _ = p.bot.Rest.EditMessage(ctx, p.channelID, msg.ID, p.toEditParams(current, true))
	return nil
}

func (p *Paginator) renderPage(index int, done bool) messages.MessageSend {
	page := p.pages[index]
	send := messages.MessageSend{
		Content: page.Content,
		Embeds:  page.Embeds,
	}
	if done {
		return send
	}

	prevDisabled := index == 0
	nextDisabled := index == len(p.pages)-1

	buttons := []components.Component{
		components.Button{
			Style:    components.ButtonStyleSecondary,
			Label:    "◀",
			CustomID: paginatorPrevID,
			Disabled: prevDisabled,
		},
		components.Button{
			Style:    components.ButtonStyleDanger,
			Label:    "✕",
			CustomID: paginatorStopID,
		},
		components.Button{
			Style:    components.ButtonStyleSecondary,
			Label:    "▶",
			CustomID: paginatorNextID,
			Disabled: nextDisabled,
		},
	}

	send.Components = []components.Component{
		components.ActionRow{Components: buttons},
	}
	return send
}

func (p *Paginator) toEditParams(index int, done bool) rest.EditMessageParams {
	page := p.pages[index]
	content := page.Content
	embeds := page.Embeds
	params := rest.EditMessageParams{
		Content: &content,
		Embeds:  &embeds,
	}
	if done {
		return params
	}

	prevDisabled := index == 0
	nextDisabled := index == len(p.pages)-1

	buttons := []components.Component{
		components.Button{
			Style:    components.ButtonStyleSecondary,
			Label:    "◀",
			CustomID: paginatorPrevID,
			Disabled: prevDisabled,
		},
		components.Button{
			Style:    components.ButtonStyleDanger,
			Label:    "✕",
			CustomID: paginatorStopID,
		},
		components.Button{
			Style:    components.ButtonStyleSecondary,
			Label:    "▶",
			CustomID: paginatorNextID,
			Disabled: nextDisabled,
		},
	}

	comps := []components.Component{
		components.ActionRow{Components: buttons},
	}
	params.Components = &comps
	return params
}

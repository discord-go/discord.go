package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
)

// Paginator manages a paginated message with button navigation.
// Each page supplies the content and embeds to display.
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
// On exit it attempts to remove the buttons, even if ctx was cancelled.
// A timeout or stop returns nil if cleanup succeeds; cancellation and any
// acknowledgement, edit, or cleanup errors are returned to the caller.
// For non-blocking use from an interaction handler, call SendAsync.
func (p *Paginator) Send(ctx context.Context) (resultErr error) {
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

	// Remove controls even when the caller has cancelled. Bound cleanup so a
	// failed Discord request cannot keep the paginator alive indefinitely.
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if _, err := p.bot.Rest.EditMessage(cleanupCtx, p.channelID, msg.ID, p.toEditParams(current, true)); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("paginator: failed to remove buttons: %w", err))
		}
	}()

	timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	interactions, stop := p.bot.CollectInteractions(timeoutCtx, func(ic *InteractionContext) bool {
		if ic == nil || ic.Interaction == nil || ic.Message == nil || ic.Message.ID != msg.ID {
			return false
		}
		if p.userID != 0 {
			user := ic.User()
			if user == nil || user.ID != p.userID {
				return false
			}
		}
		customID := ic.CustomID()
		return customID == paginatorPrevID || customID == paginatorNextID || customID == paginatorStopID
	})
	defer stop()

	for {
		var interaction *InteractionContext
		select {
		case <-timeoutCtx.Done():
			if ctx.Err() != nil {
				return fmt.Errorf("paginator: failed waiting for interaction: %w", ctx.Err())
			}
			return nil
		case interaction = <-interactions:
			if interaction == nil {
				if ctx.Err() != nil {
					return fmt.Errorf("paginator: failed waiting for interaction: %w", ctx.Err())
				}
				return nil
			}
		}

		if err := interaction.DeferUpdate(); err != nil {
			return fmt.Errorf("paginator: failed to acknowledge button interaction: %w", err)
		}
		switch interaction.CustomID() {
		case paginatorPrevID:
			if current > 0 {
				current--
			}
		case paginatorNextID:
			if current < len(p.pages)-1 {
				current++
			}
		case paginatorStopID:
			return nil
		}
		if _, err := p.bot.Rest.EditMessage(timeoutCtx, p.channelID, msg.ID, p.toEditParams(current, false)); err != nil {
			return fmt.Errorf("paginator: failed to update page: %w", err)
		}
	}
}

// SendAsync launches Send in a goroutine and returns immediately. It is
// the non-blocking variant of Send — use it from interaction handlers so
// the handler goroutine is not occupied for the paginator's lifetime.
// The error (if any) is delivered to the optional onError callback rather
// than returned, since the call returns before Send completes.
func (p *Paginator) SendAsync(ctx context.Context, onError func(error)) {
	go func() {
		if err := p.Send(ctx); err != nil && onError != nil {
			onError(err)
		}
	}()
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
		comps := []components.Component{}
		params.Components = &comps
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

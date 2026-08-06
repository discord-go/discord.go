package bot

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/discord-go/discord.go/gateway"
)

// Activity is an activity shown in the bot's Discord presence.
type Activity struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	URL  string `json:"url,omitempty"`
}

// PresenceUpdate is the gateway OP 3 payload used to update bot presence.
type PresenceUpdate struct {
	Since      *int64     `json:"since"`
	Activities []Activity `json:"activities"`
	Status     string     `json:"status"`
	AFK        bool       `json:"afk"`
}

// SetPresence updates the bot presence and reapplies it after reconnects.
func (b *Bot) SetPresence(ctx context.Context, presence PresenceUpdate) error {
	data, err := json.Marshal(presence)
	if err != nil {
		return err
	}
	b.stateMu.RLock()
	client := b.gwClient
	manager := b.shardManager
	b.stateMu.RUnlock()
	if client == nil && manager == nil {
		return errors.New("bot: gateway is not running")
	}
	payload := gateway.GatewayPayload{Op: gateway.OpcodePresenceUpdate, Data: data}
	if manager != nil {
		if err := manager.Broadcast(nonNilContext(ctx), payload); err != nil {
			return err
		}
	} else if err := client.Send(nonNilContext(ctx), payload); err != nil {
		return err
	}
	b.presenceMu.Lock()
	b.presence = &presence
	b.presenceMu.Unlock()
	return nil
}

// SetStatus updates only the status while preserving current activities.
func (b *Bot) SetStatus(ctx context.Context, status string) error {
	presence := b.currentPresence()
	presence.Status = status
	return b.SetPresence(ctx, presence)
}

// SetActivity updates the displayed activity while preserving status.
func (b *Bot) SetActivity(ctx context.Context, activity Activity) error {
	presence := b.currentPresence()
	presence.Activities = []Activity{activity}
	return b.SetPresence(ctx, presence)
}

func (b *Bot) currentPresence() PresenceUpdate {
	b.presenceMu.Lock()
	defer b.presenceMu.Unlock()
	if b.presence == nil {
		return PresenceUpdate{Status: "online", Activities: []Activity{}}
	}
	presence := *b.presence
	presence.Activities = append([]Activity(nil), b.presence.Activities...)
	return presence
}

func (b *Bot) reapplyPresence() {
	presence := b.currentPresence()
	if len(presence.Activities) == 0 && presence.Status == "" {
		return
	}
	_ = b.SetPresence(context.Background(), presence)
}

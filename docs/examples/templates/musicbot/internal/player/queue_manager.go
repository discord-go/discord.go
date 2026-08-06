// discord.go code
package player

import (
	"context"
	"log"
	"sync"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"
)

// Track represents an audio track.
type Track struct {
	Title     string
	StreamURL string
	Requested string
}

// GuildPlayer manages music for a single guild.
type GuildPlayer struct {
	mu sync.Mutex

	GuildID   snowflake.ID
	ChannelID snowflake.ID

	Queue   []*Track
	Playing bool

	VoiceClient *voice.Client

	cancel context.CancelFunc
}

// QueueManager manages players for multiple guilds.
type QueueManager struct {
	mu      sync.Mutex
	players map[snowflake.ID]*GuildPlayer
}

// NewQueueManager initializes a new QueueManager.
func NewQueueManager() *QueueManager {
	return &QueueManager{
		players: make(map[snowflake.ID]*GuildPlayer),
	}
}

// GetPlayer retrieves or creates a player for the given guild.
func (m *QueueManager) GetPlayer(guildID snowflake.ID) *GuildPlayer {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.players[guildID]
	if !ok {
		p = &GuildPlayer{
			GuildID: guildID,
			Queue:   make([]*Track, 0),
		}
		m.players[guildID] = p
	}
	return p
}

// Enqueue adds a track to the queue.
func (p *GuildPlayer) Enqueue(t *Track) {
	p.mu.Lock()
	p.Queue = append(p.Queue, t)
	p.mu.Unlock()
}

// Pop removes and returns the next track in the queue.
func (p *GuildPlayer) Pop() *Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.Queue) == 0 {
		return nil
	}
	t := p.Queue[0]
	p.Queue = p.Queue[1:]
	return t
}

// GetQueue returns a copy of the current queue.
func (p *GuildPlayer) GetQueue() []*Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	q := make([]*Track, len(p.Queue))
	copy(q, p.Queue)
	return q
}

// Stop stops the current track and clears the queue.
func (p *GuildPlayer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Queue = nil
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	p.Playing = false
	if p.VoiceClient != nil {
		_ = p.VoiceClient.Disconnect()
		p.VoiceClient = nil
	}
}

// PlayNext begins playing tracks from the queue on the provided voice client.
func (p *GuildPlayer) PlayNext(vc *voice.Client) {
	p.mu.Lock()
	if p.Playing {
		p.mu.Unlock()
		return
	}
	p.Playing = true
	p.VoiceClient = vc
	p.mu.Unlock()

	go func() {
		for {
			track := p.Pop()
			if track == nil {
				break
			}

			playCtx, cancel := context.WithCancel(context.Background())
			p.mu.Lock()
			p.cancel = cancel
			p.mu.Unlock()

			log.Printf("Playing track %s", track.Title)
			if err := StreamAudio(playCtx, vc, track.StreamURL); err != nil {
				log.Printf("Stream error: %v", err)
			}
			cancel()
		}

		p.mu.Lock()
		p.Playing = false
		p.mu.Unlock()
	}()
}

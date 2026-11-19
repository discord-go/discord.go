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

// GuildPlayer manages the playback queue for one guild.
type GuildPlayer struct {
	mu          sync.Mutex
	GuildID     snowflake.ID
	channelID   snowflake.ID
	queue       []*Track
	playing     bool
	voiceClient *voice.Client
	cancel      context.CancelFunc
	generation  uint64
}

// QueueManager manages players for multiple guilds.
type QueueManager struct {
	mu      sync.Mutex
	players map[snowflake.ID]*GuildPlayer
}

func NewQueueManager() *QueueManager {
	return &QueueManager{players: make(map[snowflake.ID]*GuildPlayer)}
}

func (m *QueueManager) GetPlayer(guildID snowflake.ID) *GuildPlayer {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.players[guildID] == nil {
		m.players[guildID] = &GuildPlayer{GuildID: guildID}
	}
	return m.players[guildID]
}

func (p *GuildPlayer) SetChannelID(id snowflake.ID) {
	p.mu.Lock()
	p.channelID = id
	p.mu.Unlock()
}

func (p *GuildPlayer) ChannelID() snowflake.ID {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.channelID
}

// Playback returns a synchronized snapshot for command handlers.
func (p *GuildPlayer) Playback() (bool, *voice.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing, p.voiceClient
}

func (p *GuildPlayer) Enqueue(t *Track) {
	p.mu.Lock()
	p.queue = append(p.queue, t)
	p.mu.Unlock()
}

func (p *GuildPlayer) GetQueue() []*Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]*Track(nil), p.queue...)
}

// Stop cancels the entire queue and disconnects the current voice client.
func (p *GuildPlayer) Stop() {
	p.mu.Lock()
	p.generation++
	p.queue = nil
	cancel, vc := p.cancel, p.voiceClient
	p.cancel = nil
	p.voiceClient = nil
	p.playing = false
	p.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if vc != nil {
		_ = vc.Disconnect()
	}
}

// PlayNext starts a single worker for the queue. It marks the queue idle under
// the same lock used by Enqueue, so an enqueue at the end cannot be stranded.
func (p *GuildPlayer) PlayNext(vc *voice.Client) {
	p.mu.Lock()
	if p.playing || vc == nil {
		p.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel, p.voiceClient, p.playing = cancel, vc, true
	generation := p.generation
	p.mu.Unlock()
	go func() {
		defer cancel()
		for {
			p.mu.Lock()
			if p.generation != generation {
				p.mu.Unlock()
				return
			}
			if len(p.queue) == 0 {
				p.playing = false
				p.cancel = nil
				p.mu.Unlock()
				return
			}
			track := p.queue[0]
			p.queue[0] = nil
			p.queue = p.queue[1:]
			p.mu.Unlock()
			log.Printf("Playing track %s", track.Title)
			if err := StreamAudio(ctx, vc, track.StreamURL); err != nil && ctx.Err() == nil {
				log.Printf("Stream error: %v", err)
			}
		}
	}()
}

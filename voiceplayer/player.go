// Package voiceplayer provides paced playback of Opus frames over a connected
// Discord voice client. It intentionally leaves decoding and media acquisition
// to the application.
package voiceplayer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/discord-go/discord.go/voice"
)

var (
	ErrAlreadyPlaying = errors.New("voiceplayer: player is already active")
	ErrNotPlaying     = errors.New("voiceplayer: player is not active")
)

// FrameInterval is the wall-clock interval for Discord's standard 20ms Opus
// packet cadence. It matches voice.Client's RTP timestamp increment.
const FrameInterval = 20 * time.Millisecond

// Sender is the subset of a voice client needed to play audio.
type Sender interface {
	SendOpus([]byte) error
	SetSpeaking(voice.SpeakingFlag) error
}

// Source returns one Opus packet at a time. Implementations must return when
// ctx is canceled so Stop can release a blocked source read.
type Source interface {
	ReadFrame(ctx context.Context) ([]byte, error)
	Close() error
}

// State reports the current player state.
type State uint8

const (
	Idle State = iota
	Playing
	Paused
)

func (s State) String() string {
	switch s {
	case Idle:
		return "idle"
	case Playing:
		return "playing"
	case Paused:
		return "paused"
	default:
		return fmt.Sprintf("State(%d)", s)
	}
}

// Player sends source frames at Discord's usual 20ms Opus frame interval.
type Player struct {
	mu      sync.Mutex
	sender  Sender
	state   State
	changed chan struct{}
	cancel  context.CancelFunc
	run     *playback
}

type playback struct {
	done chan struct{}
	err  error
}

// New creates a player that uses Discord's standard 20ms Opus packet cadence.
func New(sender Sender) (*Player, error) {
	if sender == nil {
		return nil, errors.New("voiceplayer: sender is required")
	}
	return &Player{sender: sender, state: Idle, changed: make(chan struct{})}, nil
}

// Play starts source playback asynchronously. The caller can observe the
// result with Wait, pause or resume it, or stop it. A source is closed when
// playback finishes for any reason.
func (p *Player) Play(ctx context.Context, source Source) error {
	if source == nil {
		return errors.New("voiceplayer: source is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state != Idle {
		return ErrAlreadyPlaying
	}
	playCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	run := &playback{done: make(chan struct{})}
	p.run = run
	p.state = Playing
	p.notifyLocked()
	go p.playRun(playCtx, source, run)
	return nil
}

// Pause suspends packet sending. The source read is context-aware and may
// finish one frame already in progress; that frame is held until playback
// resumes.
func (p *Player) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state != Playing {
		return ErrNotPlaying
	}
	p.state = Paused
	p.notifyLocked()
	return nil
}

// Resume resumes a paused player.
func (p *Player) Resume() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.state != Paused {
		return ErrNotPlaying
	}
	p.state = Playing
	p.notifyLocked()
	return nil
}

// Stop cancels playback and waits for the source and worker to exit.
func (p *Player) Stop() error {
	p.mu.Lock()
	if p.state == Idle {
		p.mu.Unlock()
		return nil
	}
	cancel, done := p.cancel, p.run.done
	p.mu.Unlock()
	cancel()
	<-done
	return nil
}

// Wait blocks until the current playback ends and returns its error. It
// returns ErrNotPlaying when no playback has been started.
func (p *Player) Wait() error {
	p.mu.Lock()
	run := p.run
	if run == nil {
		p.mu.Unlock()
		return ErrNotPlaying
	}
	p.mu.Unlock()
	<-run.done
	return run.err
}

// State returns the current state.
func (p *Player) State() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *Player) playRun(ctx context.Context, source Source, run *playback) {
	err := p.play(ctx, source)
	closeErr := source.Close()
	if err == nil {
		err = closeErr
	}

	p.mu.Lock()
	cancel := p.cancel
	p.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if speakingErr := p.sender.SetSpeaking(0); err == nil {
		err = speakingErr
	}

	p.mu.Lock()
	run.err = err
	p.state = Idle
	p.cancel = nil
	p.notifyLocked()
	close(run.done)
	p.mu.Unlock()
}

func (p *Player) play(ctx context.Context, source Source) error {
	if err := p.sender.SetSpeaking(voice.Microphone); err != nil {
		return err
	}
	ticker := time.NewTicker(FrameInterval)
	defer ticker.Stop()
	for {
		if err := p.waitUntilPlaying(ctx); err != nil {
			return err
		}
		frame, err := source.ReadFrame(ctx)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if len(frame) == 0 {
			return errors.New("voiceplayer: source returned an empty Opus frame")
		}
		if err := p.waitUntilPlaying(ctx); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		if err := p.waitUntilPlaying(ctx); err != nil {
			return err
		}
		if err := p.sender.SendOpus(frame); err != nil {
			return err
		}
	}
}

func (p *Player) waitUntilPlaying(ctx context.Context) error {
	for {
		p.mu.Lock()
		state, changed := p.state, p.changed
		p.mu.Unlock()
		if state == Playing {
			return nil
		}
		if state == Idle {
			return context.Canceled
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

func (p *Player) notifyLocked() {
	close(p.changed)
	p.changed = make(chan struct{})
}

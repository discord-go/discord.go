// discord.go code
package player

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/discord-go/discord.go/voice"
	"github.com/discord-go/discord.go/voiceplayer"
	"layeh.com/gopus"
)

// StreamAudio converts a media input and lets voiceplayer pace its Opus frames.
func StreamAudio(ctx context.Context, vc *voice.Client, input string) error {
	source, err := NewFFmpegSource(ctx, input)
	if err != nil {
		return err
	}
	playback, err := voiceplayer.New(vc)
	if err != nil {
		_ = source.Close()
		return err
	}
	if err := playback.Play(ctx, source); err != nil {
		_ = source.Close()
		return err
	}
	return playback.Wait()
}

// StreamTestTone sends a short synthetic tone for end-to-end voice tests.
func StreamTestTone(ctx context.Context, vc *voice.Client, duration time.Duration) error {
	if vc.DaveSession != nil {
		readyCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		err := WaitVoiceReady(readyCtx, vc)
		cancel()
		if err != nil {
			return fmt.Errorf("DAVE session did not become ready: %w", err)
		}
	}
	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		return err
	}
	if err := vc.SetSpeaking(voice.Microphone); err != nil {
		return err
	}
	defer func() { _ = vc.SetSpeaking(0) }()

	const frameSize = 960
	pcm := make([]int16, frameSize*2)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	frames := int(duration / (20 * time.Millisecond))
	for frame := 0; frame < frames; frame++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		for sample := 0; sample < frameSize; sample++ {
			value := int16(math.Sin(2*math.Pi*440*float64(frame*frameSize+sample)/48000) * 5000)
			pcm[sample*2], pcm[sample*2+1] = value, value
		}
		opus, err := encoder.Encode(pcm, frameSize, 4000)
		if err != nil {
			return err
		}
		if err := vc.SendOpus(opus); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	return nil
}

// WaitVoiceReady waits for the active DAVE session before sending audio.
func WaitVoiceReady(ctx context.Context, vc *voice.Client) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for vc.DaveSession != nil && !vc.DaveSession.Ready() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	return nil
}

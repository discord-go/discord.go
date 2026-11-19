package voiceplayer_test

import (
	"context"
	"io"
	"sync"

	"github.com/discord-go/discord.go/voice"
	"github.com/discord-go/discord.go/voiceplayer"
)

type exampleSender struct {
	mu     sync.Mutex
	frames int
}

func (s *exampleSender) SendOpus([]byte) error {
	s.mu.Lock()
	s.frames++
	s.mu.Unlock()
	return nil
}

func (s *exampleSender) SetSpeaking(voice.SpeakingFlag) error { return nil }

type sliceSource struct {
	frames [][]byte
}

func (s *sliceSource) ReadFrame(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(s.frames) == 0 {
		return nil, io.EOF
	}
	frame := s.frames[0]
	s.frames = s.frames[1:]
	return frame, nil
}

func (s *sliceSource) Close() error { return nil }

// ExamplePlayer_Play shows the Source contract. A real source can decode a
// local file or an FFmpeg pipe; the musicbot template is the reference
// implementation.
func ExamplePlayer_Play() {
	sender := &exampleSender{}
	player, err := voiceplayer.New(sender)
	if err != nil {
		panic(err)
	}
	source := &sliceSource{frames: [][]byte{{0x01}, {0x02}}}
	if err := player.Play(context.Background(), source); err != nil {
		panic(err)
	}
	if err := player.Wait(); err != nil {
		panic(err)
	}
	// Output:
}

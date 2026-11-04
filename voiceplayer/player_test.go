package voiceplayer

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/discord-go/discord.go/voice"
)

type testSender struct {
	mu       sync.Mutex
	frames   [][]byte
	speaking []voice.SpeakingFlag
}

func (s *testSender) SendOpus(frame []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.frames = append(s.frames, append([]byte(nil), frame...))
	return nil
}

func (s *testSender) SetSpeaking(flag voice.SpeakingFlag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.speaking = append(s.speaking, flag)
	return nil
}

type channelSource struct {
	frames chan []byte
	closed chan struct{}
	once   sync.Once
}

func newChannelSource() *channelSource {
	return &channelSource{frames: make(chan []byte), closed: make(chan struct{})}
}

func (s *channelSource) ReadFrame(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.closed:
		return nil, io.EOF
	case frame := <-s.frames:
		return frame, nil
	}
}

func (s *channelSource) Close() error {
	s.once.Do(func() { close(s.closed) })
	return nil
}

func TestPlayerPlayCompletesAndSignalsSpeaking(t *testing.T) {
	sender := &testSender{}
	player, err := New(sender)
	if err != nil {
		t.Fatal(err)
	}
	source := newChannelSource()
	if err := player.Play(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	go func() {
		source.frames <- []byte{1, 2, 3}
		source.Close()
	}()
	if err := player.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if got := player.State(); got != Idle {
		t.Fatalf("State() = %v, want idle", got)
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if len(sender.frames) != 1 || string(sender.frames[0]) != string([]byte{1, 2, 3}) {
		t.Fatalf("sent frames = %v, want one Opus frame", sender.frames)
	}
	if len(sender.speaking) != 2 || sender.speaking[0] != voice.Microphone || sender.speaking[1] != 0 {
		t.Fatalf("speaking transitions = %v", sender.speaking)
	}
}

func TestPlayerPauseAndResume(t *testing.T) {
	sender := &testSender{}
	player, err := New(sender)
	if err != nil {
		t.Fatal(err)
	}
	source := newChannelSource()
	if err := player.Play(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if err := player.Pause(); err != nil {
		t.Fatal(err)
	}
	if err := player.Pause(); !errors.Is(err, ErrNotPlaying) {
		t.Fatalf("second Pause() = %v, want ErrNotPlaying", err)
	}
	go func() {
		source.frames <- []byte{9}
		source.Close()
	}()
	time.Sleep(10 * time.Millisecond)
	sender.mu.Lock()
	pausedSent := len(sender.frames)
	sender.mu.Unlock()
	if pausedSent != 0 {
		t.Fatalf("sent %d frames while paused", pausedSent)
	}
	if err := player.Resume(); err != nil {
		t.Fatal(err)
	}
	if err := player.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if len(sender.frames) != 1 {
		t.Fatalf("sent %d frames after resume, want 1", len(sender.frames))
	}
}

func TestPlayerStopCancelsBlockedSource(t *testing.T) {
	player, err := New(&testSender{})
	if err != nil {
		t.Fatal(err)
	}
	source := newChannelSource()
	if err := player.Play(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if err := player.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := player.Wait(); !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait() = %v, want context.Canceled", err)
	}
	select {
	case <-source.closed:
	default:
		t.Fatal("Stop did not close the source")
	}
}

func TestPlayerValidation(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("New accepted a nil sender")
	}
	player, err := New(&testSender{})
	if err != nil {
		t.Fatal(err)
	}
	if err := player.Play(context.Background(), nil); err == nil {
		t.Fatal("Play accepted a nil source")
	}
	if err := player.Wait(); !errors.Is(err, ErrNotPlaying) {
		t.Fatalf("Wait() = %v, want ErrNotPlaying", err)
	}
}

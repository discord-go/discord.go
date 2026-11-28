package ytdlp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRequestedInput(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"  discord music  ", "ytsearch1:discord music"},
		{"https://www.youtube.com/watch?v=abc", "https://www.youtube.com/watch?v=abc"},
		{"https://youtu.be/abc", "https://youtu.be/abc"},
		{`C:\media\test.wav`, `ytsearch1:C:\media\test.wav`},
		{"https://example.net/video", ""},
		{"file:///media/song.wav", ""},
		{"", ""},
	} {
		got, err := requestedInput(tt.input)
		if got != tt.want || (err != nil) != (tt.want == "") {
			t.Errorf("requestedInput(%q) = %q, %v; want %q", tt.input, got, err, tt.want)
		}
	}
}

func TestOperatorCanSelectLocalMedia(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.wav")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	got, title, err := ExtractAudioInfo(context.Background(), path)
	if err != nil || got != path || title != "sample.wav" {
		t.Fatalf("local media = %q, %q, %v", got, title, err)
	}
}

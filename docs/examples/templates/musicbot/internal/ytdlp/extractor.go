// discord.go code
package ytdlp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExtractAudioInfo resolves media chosen by the bot operator, including
// existing local files used for offline tests.
func ExtractAudioInfo(ctx context.Context, videoURL string) (streamURL string, title string, err error) {
	if info, statErr := os.Stat(videoURL); statErr == nil && !info.IsDir() {
		return videoURL, filepath.Base(videoURL), nil
	}
	if !strings.HasPrefix(videoURL, "http://") && !strings.HasPrefix(videoURL, "https://") {
		videoURL = "ytsearch1:" + videoURL
	}
	return extract(ctx, videoURL)
}

// ExtractRequestedAudioInfo resolves a Discord user's YouTube URL or search.
func ExtractRequestedAudioInfo(ctx context.Context, input string) (streamURL string, title string, err error) {
	videoURL, err := requestedInput(input)
	if err != nil {
		return "", "", err
	}
	return extract(ctx, videoURL)
}

func requestedInput(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("enter a YouTube URL or search term")
	}
	if !strings.Contains(input, "://") {
		return "ytsearch1:" + input, nil
	}
	u, err := url.Parse(input)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.User == nil && u.Port() == "" {
		switch strings.ToLower(u.Hostname()) {
		case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com", "youtu.be":
			return u.String(), nil
		}
	}
	return "", fmt.Errorf("only YouTube URLs are supported")
}

func extract(ctx context.Context, videoURL string) (streamURL string, title string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// yt-dlp -f bestaudio -g -e <url> prints title then stream URL.
	out, err := exec.CommandContext(ctx, "yt-dlp", "--no-playlist", "-f", "bestaudio", "-g", "-e", "--", videoURL).Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to run yt-dlp: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("unexpected yt-dlp output")
	}

	title = strings.TrimSpace(lines[len(lines)-2])
	streamURL = strings.TrimSpace(lines[len(lines)-1])
	return streamURL, title, nil
}

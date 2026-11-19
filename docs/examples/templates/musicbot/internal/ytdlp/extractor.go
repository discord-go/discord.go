// discord.go code
package ytdlp

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtractAudioInfo extracts the audio stream URL and title using yt-dlp.
// Existing local files are returned as-is so the template can play test media
// without a network round-trip.
func ExtractAudioInfo(videoURL string) (streamURL string, title string, err error) {
	if info, statErr := os.Stat(videoURL); statErr == nil && !info.IsDir() {
		return videoURL, filepath.Base(videoURL), nil
	}
	if !strings.HasPrefix(videoURL, "http://") && !strings.HasPrefix(videoURL, "https://") {
		videoURL = "ytsearch1:" + videoURL
	}

	// yt-dlp -f bestaudio -g -e <url> prints title then stream URL
	cmd := exec.Command("yt-dlp", "--no-playlist", "-f", "bestaudio", "-g", "-e", videoURL)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("failed to run yt-dlp: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 2 {
		return "", "", fmt.Errorf("unexpected yt-dlp output")
	}

	title = lines[len(lines)-2]
	streamURL = lines[len(lines)-1]
	return streamURL, title, nil
}

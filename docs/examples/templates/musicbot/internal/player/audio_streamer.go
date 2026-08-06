// discord.go code
package player

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/discord-go/discord.go/voice"
	"layeh.com/gopus"
)

// StreamAudio starts ffmpeg to stream audio from streamURL, encodes to Opus, and sends to vc.
func StreamAudio(ctx context.Context, vc *voice.Client, streamURL string) error {
	// Send Speaking before audio
	if err := vc.SetSpeaking(voice.Microphone); err != nil {
		log.Printf("SetSpeaking error: %v", err)
	}

	// Start ffmpeg
	cmd := exec.CommandContext(ctx, "ffmpeg", "-reconnect", "1", "-reconnect_streamed", "1",
		"-i", streamURL,
		"-f", "s16le", "-ar", "48000", "-ac", "2", "-loglevel", "error", "pipe:1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	log.Printf("AUDIO: ffmpeg started, streaming audio...")

	// Wait for process to finish
	defer func() {
		_ = cmd.Wait()
		// Stop speaking when done
		_ = vc.SetSpeaking(0)
		log.Printf("AUDIO: Playback finished")
	}()

	// Initialize Opus encoder
	// 48000 Hz, 2 channels, application audio
	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		return fmt.Errorf("failed to create opus encoder: %w", err)
	}

	// 20ms of audio at 48000Hz and 2 channels (16-bit)
	// 48000 * 0.02 = 960 samples per channel
	frameSize := 960
	bufferSize := frameSize * 2 * 2 // 2 channels, 2 bytes per sample
	pcmBuf := make([]int16, frameSize*2)
	byteBuf := make([]byte, bufferSize)

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	frameCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, err := io.ReadFull(stdout, byteBuf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return fmt.Errorf("read error: %w", err)
		}

		// Convert bytes to int16 (little-endian)
		for i := 0; i < len(pcmBuf); i++ {
			pcmBuf[i] = int16(binary.LittleEndian.Uint16(byteBuf[i*2 : (i+1)*2]))
		}

		// Encode to opus
		opusData, err := encoder.Encode(pcmBuf, frameSize, 4000)
		if err != nil {
			log.Printf("Opus encode error: %v", err)
			continue
		}

		// Wait for the next 20ms tick
		<-ticker.C

		// Send to discord voice connection
		if err := vc.SendOpus(opusData); err != nil {
			log.Printf("SendOpus error: %v", err)
			return err
		}

		frameCount++
		if frameCount == 1 {
			log.Printf("AUDIO: First frame sent successfully!")
		}
	}

	return nil
}

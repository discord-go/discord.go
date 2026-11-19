package player

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"layeh.com/gopus"
)

const opusSamples = 960 // 20ms at 48 kHz

// FFmpegSource converts a local file or media URL to 20ms Opus packets. It is
// an example adapter; FFmpeg and a working C toolchain for gopus are required.
// The URL is passed as an argument, never interpreted by a shell.
type FFmpegSource struct {
	cmd         *exec.Cmd
	stdout      io.ReadCloser
	encoder     *gopus.Encoder
	finished    bool
	terminalErr error
}

func NewFFmpegSource(ctx context.Context, input string) (*FFmpegSource, error) {
	if ctx == nil || input == "" {
		return nil, errors.New("ffmpeg: context and input are required")
	}
	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		return nil, fmt.Errorf("create Opus encoder: %w", err)
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", "-loglevel", "error", "-i", input,
		"-vn", "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}
	return &FFmpegSource{cmd: cmd, stdout: stdout, encoder: encoder}, nil
}

func (s *FFmpegSource) ReadFrame(ctx context.Context) ([]byte, error) {
	if s.finished {
		if s.terminalErr != nil {
			err := s.terminalErr
			s.terminalErr = nil
			return nil, err
		}
		return nil, io.EOF
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Canceling a blocked pipe read must also stop FFmpeg.
	stop := context.AfterFunc(ctx, func() { _ = s.cmd.Process.Kill() })
	defer stop()
	data := make([]byte, opusSamples*2*2)
	n, err := io.ReadFull(s.stdout, data)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, err
		}
		s.finished = true
		if waitErr := s.cmd.Wait(); waitErr != nil {
			s.terminalErr = fmt.Errorf("ffmpeg: %w", waitErr)
		}
		if n == 0 {
			if s.terminalErr != nil {
				err := s.terminalErr
				s.terminalErr = nil
				return nil, err
			}
			return nil, io.EOF
		}
	}
	pcm := make([]int16, opusSamples*2)
	for i := range pcm {
		pcm[i] = int16(binary.LittleEndian.Uint16(data[i*2:]))
	}
	return s.encoder.Encode(pcm, opusSamples, 4000)
}

func (s *FFmpegSource) Close() error {
	if s.finished {
		return nil
	}
	s.finished = true
	_ = s.stdout.Close()
	_ = s.cmd.Process.Kill()
	_ = s.cmd.Wait()
	return nil
}

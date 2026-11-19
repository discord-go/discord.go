package player

import (
	"context"
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFFmpegSourceLocalWAV(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is not installed")
	}
	// A 30ms PCM file exercises both a full frame and a padded final frame.
	const samples = 1440 * 2
	pcm := make([]byte, samples*2)
	for i := 0; i < samples; i++ {
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(i%1000))
	}
	header := make([]byte, 44)
	copy(header, "RIFF")
	binary.LittleEndian.PutUint32(header[4:], uint32(36+len(pcm)))
	copy(header[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(header[16:], 16)
	binary.LittleEndian.PutUint16(header[20:], 1)
	binary.LittleEndian.PutUint16(header[22:], 2)
	binary.LittleEndian.PutUint32(header[24:], 48000)
	binary.LittleEndian.PutUint32(header[28:], 192000)
	binary.LittleEndian.PutUint16(header[32:], 4)
	binary.LittleEndian.PutUint16(header[34:], 16)
	copy(header[36:], "data")
	binary.LittleEndian.PutUint32(header[40:], uint32(len(pcm)))
	path := filepath.Join(t.TempDir(), "sample.wav")
	if err := os.WriteFile(path, append(header, pcm...), 0o600); err != nil {
		t.Fatal(err)
	}
	source, err := NewFFmpegSource(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	for i := 0; i < 2; i++ {
		frame, err := source.ReadFrame(context.Background())
		if err != nil || len(frame) == 0 {
			t.Fatalf("frame %d: %d bytes, %v", i, len(frame), err)
		}
	}
	if _, err := source.ReadFrame(context.Background()); err != io.EOF {
		t.Fatalf("end: %v", err)
	}
}

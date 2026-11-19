package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"layeh.com/gopus"
)

func TestRecorderWritesFinalizedWAV(t *testing.T) {
	dir := t.TempDir()
	r := newRecorder(dir)
	if err := r.start(); err != nil {
		t.Fatal(err)
	}
	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		t.Fatal(err)
	}
	opus, err := encoder.Encode(make([]int16, 1920), 960, 4000)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.write("speaker", opus); err != nil {
		t.Fatal(err)
	}
	r.stop()

	data, err := os.ReadFile(filepath.Join(dir, "speaker.wav"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) <= 44 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" || string(data[12:16]) != "fmt " {
		t.Fatalf("not a WAV recording: %d bytes", len(data))
	}
	if got, want := binary.LittleEndian.Uint32(data[4:8]), uint32(len(data)-8); got != want {
		t.Fatalf("RIFF size = %d, want %d", got, want)
	}
	if got := binary.LittleEndian.Uint16(data[20:22]); got != 1 {
		t.Fatalf("audio format = %d, want PCM", got)
	}
	if got := binary.LittleEndian.Uint16(data[22:24]); got != 2 {
		t.Fatalf("channels = %d, want 2", got)
	}
	if got := binary.LittleEndian.Uint32(data[24:28]); got != 48000 {
		t.Fatalf("sample rate = %d, want 48000", got)
	}
	if got, want := binary.LittleEndian.Uint32(data[40:44]), uint32(len(data)-44); got != want {
		t.Fatalf("data chunk size = %d, want %d", got, want)
	}
}

func TestRecorderDoesNotCreateFilesWhenInactive(t *testing.T) {
	r := newRecorder(t.TempDir())
	if err := r.write("speaker", []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if len(r.files) != 0 {
		t.Fatal("inactive recorder created an output file")
	}
}

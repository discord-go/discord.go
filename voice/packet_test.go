package voice

import (
	"bytes"
	"testing"
)

func TestRTPHeader(t *testing.T) {
	header := NewRTPHeader(100, 200, 300)
	marshaled := header.Marshal()

	if len(marshaled) != 12 {
		t.Errorf("Expected marshal length 12, got %d", len(marshaled))
	}

	parsed, err := ParseRTPHeader(marshaled)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Sequence != 100 || parsed.Timestamp != 200 || parsed.SSRC != 300 {
		t.Errorf("Parsed header mismatch: %+v", parsed)
	}
	if parsed.PayloadType != 120 {
		t.Errorf("Expected payload type 120, got %d", parsed.PayloadType)
	}
	if parsed.Version != 2 {
		t.Errorf("Expected version 2, got %d", parsed.Version)
	}

	audioPacket := BuildAudioPacket(header, []byte{1, 2, 3})
	if len(audioPacket) != 15 {
		t.Errorf("Expected audio packet length 15, got %d", len(audioPacket))
	}
	if !bytes.Equal(audioPacket[12:], []byte{1, 2, 3}) {
		t.Errorf("Audio payload mismatch")
	}
}

func TestParseRTPHeaderError(t *testing.T) {
	_, err := ParseRTPHeader([]byte{1, 2, 3})
	if err == nil {
		t.Error("Expected error for short packet")
	}
}

package voice

import "testing"

func TestParseRTPHeaderMarkerAndPadding(t *testing.T) {
	header := NewRTPHeader(1, 2, 3)
	header.Marker = true
	header.Padding = true
	packet := append(header.Marshal(), 1, 2, 3, 1)
	parsed, err := ParseRTPHeader(packet)
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.Marker || !parsed.Padding || parsed.PayloadType != 120 {
		t.Fatalf("parsed RTP header = %#v", parsed)
	}
}

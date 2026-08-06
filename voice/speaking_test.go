package voice

import (
	"encoding/json"
	"testing"
)

func TestSpeaking(t *testing.T) {
	if Microphone != 1 || Soundshare != 2 || Priority != 4 {
		t.Error("Speaking flags are incorrect")
	}

	payload := SpeakingPayload{
		Speaking: Microphone,
		Delay:    0,
		SSRC:     12345,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	var parsed SpeakingPayload
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	if parsed.Speaking != Microphone || parsed.SSRC != 12345 {
		t.Error("Payload mismatch after marshal/unmarshal")
	}
}

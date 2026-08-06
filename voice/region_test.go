package voice

import (
	"encoding/json"
	"testing"
)

func TestVoiceRegion_JSON(t *testing.T) {
	raw := `{
		"id": "us-west",
		"name": "US West",
		"optimal": true,
		"deprecated": false,
		"custom": false
	}`

	var region VoiceRegion
	if err := json.Unmarshal([]byte(raw), &region); err != nil {
		t.Fatalf("failed to unmarshal VoiceRegion: %v", err)
	}

	if region.ID != "us-west" {
		t.Errorf("expected ID %q, got %q", "us-west", region.ID)
	}
	if region.Name != "US West" {
		t.Errorf("expected Name %q, got %q", "US West", region.Name)
	}
	if !region.Optimal {
		t.Error("expected Optimal to be true")
	}
	if region.Deprecated {
		t.Error("expected Deprecated to be false")
	}
	if region.Custom {
		t.Error("expected Custom to be false")
	}
}

func TestVoiceRegion_Marshal(t *testing.T) {
	region := VoiceRegion{
		ID:         "eu-central",
		Name:       "EU Central",
		Optimal:    false,
		Deprecated: true,
		Custom:     true,
	}

	data, err := json.Marshal(region)
	if err != nil {
		t.Fatalf("failed to marshal VoiceRegion: %v", err)
	}

	var decoded VoiceRegion
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal marshaled VoiceRegion: %v", err)
	}

	if decoded.ID != region.ID {
		t.Errorf("expected ID %q, got %q", region.ID, decoded.ID)
	}
	if decoded.Name != region.Name {
		t.Errorf("expected Name %q, got %q", region.Name, decoded.Name)
	}
	if decoded.Optimal != region.Optimal {
		t.Errorf("expected Optimal %v, got %v", region.Optimal, decoded.Optimal)
	}
	if decoded.Deprecated != region.Deprecated {
		t.Errorf("expected Deprecated %v, got %v", region.Deprecated, decoded.Deprecated)
	}
	if decoded.Custom != region.Custom {
		t.Errorf("expected Custom %v, got %v", region.Custom, decoded.Custom)
	}
}

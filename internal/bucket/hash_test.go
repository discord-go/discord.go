package bucket

import (
	"testing"
)

func TestRouteHash_Basic(t *testing.T) {
	h := RouteHash("GET", "/channels/123456789012345678/messages")
	if h == "" {
		t.Fatal("expected non-empty hash")
	}
	// Should be a hex-encoded SHA-256 (64 chars).
	if len(h) != 64 {
		t.Errorf("hash length = %d, want 64", len(h))
	}
}

func TestRouteHash_Deterministic(t *testing.T) {
	h1 := RouteHash("GET", "/channels/123456789012345678/messages")
	h2 := RouteHash("GET", "/channels/123456789012345678/messages")
	if h1 != h2 {
		t.Errorf("same input produced different hashes: %q != %q", h1, h2)
	}
}

func TestRouteHash_NormalizesSnowflakes(t *testing.T) {
	// Two different channel IDs should produce the same hash because
	// snowflakes are replaced by a placeholder.
	h1 := RouteHash("GET", "/channels/123456789012345678/messages")
	h2 := RouteHash("GET", "/channels/987654321098765432/messages")
	if h1 != h2 {
		t.Errorf("different snowflakes should produce same hash: %q != %q", h1, h2)
	}
}

func TestRouteHash_DifferentMethodsDifferentHash(t *testing.T) {
	hGet := RouteHash("GET", "/channels/123456789012345678/messages")
	hDelete := RouteHash("DELETE", "/channels/123456789012345678/messages")
	if hGet == hDelete {
		t.Error("different methods should produce different hashes")
	}
}

func TestRouteHash_DifferentRoutesDifferentHash(t *testing.T) {
	h1 := RouteHash("GET", "/channels/123456789012345678/messages")
	h2 := RouteHash("GET", "/guilds/123456789012345678/members")
	if h1 == h2 {
		t.Error("different routes should produce different hashes")
	}
}

func TestRouteHash_CaseInsensitiveMethod(t *testing.T) {
	h1 := RouteHash("get", "/channels/123456789012345678/messages")
	h2 := RouteHash("GET", "/channels/123456789012345678/messages")
	if h1 != h2 {
		t.Errorf("method case should be normalized: %q != %q", h1, h2)
	}
}

func TestRouteHash_EmptyMethod(t *testing.T) {
	h := RouteHash("", "/channels/123456789012345678/messages")
	if h != "" {
		t.Errorf("empty method should return empty string, got %q", h)
	}
}

func TestRouteHash_EmptyPath(t *testing.T) {
	h := RouteHash("GET", "")
	if h != "" {
		t.Errorf("empty path should return empty string, got %q", h)
	}
}

func TestRouteHash_BothEmpty(t *testing.T) {
	h := RouteHash("", "")
	if h != "" {
		t.Errorf("both empty should return empty string, got %q", h)
	}
}

func TestRouteHash_MultipleSnowflakes(t *testing.T) {
	// A path with multiple snowflakes should normalize all of them.
	h1 := RouteHash("DELETE", "/channels/123456789012345678/messages/876543210987654321")
	h2 := RouteHash("DELETE", "/channels/111111111111111111/messages/222222222222222222")
	if h1 != h2 {
		t.Errorf("multiple snowflakes should be normalized: %q != %q", h1, h2)
	}
}

func TestRouteHash_NoSnowflakes(t *testing.T) {
	// A path without snowflakes should still produce a valid hash.
	h := RouteHash("GET", "/gateway/bot")
	if h == "" {
		t.Fatal("expected non-empty hash for path without snowflakes")
	}
	if len(h) != 64 {
		t.Errorf("hash length = %d, want 64", len(h))
	}
}

func TestRouteHash_ShortNumbers(t *testing.T) {
	// Numbers shorter than 17 digits should NOT be treated as snowflakes.
	h1 := RouteHash("GET", "/channels/12345/messages")
	h2 := RouteHash("GET", "/channels/67890/messages")
	if h1 == h2 {
		t.Error("short numbers should not be normalized as snowflakes")
	}
}

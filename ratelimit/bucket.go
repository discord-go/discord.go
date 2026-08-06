package ratelimit

import (
	"net/http"
	"strconv"
	"time"
)

// Info contains parsed rate limit headers from the Discord API.
type Info struct {
	Bucket     string
	Remaining  int
	Reset      time.Time
	ResetAfter time.Duration
	Global     bool
	Scope      string
}

// ParseHeaders extracts rate limit information from HTTP response headers.
func ParseHeaders(header http.Header) Info {
	var info Info
	info.Bucket = header.Get("X-RateLimit-Bucket")
	info.Scope = header.Get("X-RateLimit-Scope")

	if r := header.Get("X-RateLimit-Remaining"); r != "" {
		if remaining, err := strconv.Atoi(r); err == nil {
			info.Remaining = remaining
		}
	}

	if ra := header.Get("X-RateLimit-Reset-After"); ra != "" {
		if resetAfter, err := strconv.ParseFloat(ra, 64); err == nil {
			info.ResetAfter = time.Duration(resetAfter * float64(time.Second))
		}
	}

	if r := header.Get("X-RateLimit-Reset"); r != "" {
		if reset, err := strconv.ParseFloat(r, 64); err == nil {
			info.Reset = time.Unix(0, int64(reset*float64(time.Second)))
		}
	}

	if g := header.Get("X-RateLimit-Global"); g != "" {
		if global, err := strconv.ParseBool(g); err == nil {
			info.Global = global
		}
	}

	return info
}

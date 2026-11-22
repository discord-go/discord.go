package bucket

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

// snowflakePattern matches Discord snowflake IDs (numeric strings of 17-20 digits).
var snowflakePattern = regexp.MustCompile(`\d{17,20}`)

// RouteHash computes a deterministic hash for a normalized API route.
//
// RouteHash replaces snowflake IDs in the path with a placeholder before
// hashing. The result is not a Discord rate limit bucket key: it omits major
// resource IDs and is not the bucket identifier provided by Discord.
//
// The method is included because Discord sometimes uses different buckets for
// different HTTP methods on the same route (e.g. DELETE /channels/{id}/messages/{id}
// has a different bucket from GET on the same path).
//
// Both method and path are required; an empty method or path returns an empty string.
func RouteHash(method, path string) string {
	if method == "" || path == "" {
		return ""
	}

	method = strings.ToUpper(method)

	// Replace snowflake IDs with a fixed placeholder so that different
	// resource IDs produce the same bucket key.
	normalized := snowflakePattern.ReplaceAllString(path, ":id")

	raw := fmt.Sprintf("%s:%s", method, normalized)

	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", hash)
}

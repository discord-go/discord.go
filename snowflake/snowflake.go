package snowflake

import (
	"encoding/json"
	"strconv"
)

// ID represents a Discord snowflake ID.
type ID uint64

// String returns the string representation of the ID.
func (id ID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// MarshalJSON encodes the ID as a JSON string. Discord serializes
// snowflakes as strings in API payloads; a numeric encoding is rejected
// with error 50035 (Invalid Form Body), most visibly in
// allowed_mentions.users and allowed_mentions.roles.
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

// IsZero reports whether the ID is the zero value (0), indicating an unset
// or missing snowflake.
func (id ID) IsZero() bool {
	return id == 0
}

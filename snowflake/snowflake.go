package snowflake

import "strconv"

// ID represents a Discord snowflake ID.
type ID uint64

// String returns the string representation of the ID.
func (id ID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// IsZero reports whether the ID is the zero value (0), indicating an unset
// or missing snowflake.
func (id ID) IsZero() bool {
	return id == 0
}

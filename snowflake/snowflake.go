package snowflake

import "strconv"

// ID represents a Discord snowflake ID.
type ID uint64

// String returns the string representation of the ID.
func (id ID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

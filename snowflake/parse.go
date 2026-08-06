package snowflake

import "strconv"

// Parse parses a string representation of a Discord snowflake and returns an ID.
// Returns an error if the string is not a valid 64-bit unsigned integer.
func Parse(s string) (ID, error) {
	parsed, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return ID(parsed), nil
}

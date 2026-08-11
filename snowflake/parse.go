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

// MustParse parses a string representation of a Discord snowflake and panics
// on error. Use it only for compile-time constants or config values known
// to be valid; for untrusted input, use Parse.
func MustParse(s string) ID {
	id, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

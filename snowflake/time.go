package snowflake

import "time"

// DiscordEpoch is the Discord epoch (2015-01-01T00:00:00Z) in milliseconds.
const DiscordEpoch = 1420070400000

// Time returns the time.Time corresponding to the timestamp part of the snowflake ID.
func (id ID) Time() time.Time {
	// The timestamp is the first 42 bits of the snowflake, representing the number of milliseconds since DiscordEpoch.
	timestampMs := (uint64(id) >> 22) + DiscordEpoch
	sec := timestampMs / 1000
	nsec := (timestampMs % 1000) * 1000000
	return time.Unix(int64(sec), int64(nsec)).UTC()
}

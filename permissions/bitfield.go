package permissions

import "strconv"

// Permission represents a 64-bit integer containing multiple Discord permission flags.
type Permission uint64

// String returns the decimal string representation of the permission bitfield.
// This is the format Discord expects in channel permission overwrites.
func (p Permission) String() string {
	return strconv.FormatUint(uint64(p), 10)
}

// Add adds the given permissions to the bitfield.
func (p *Permission) Add(permissions Permission) {
	*p |= permissions
}

// Remove removes the given permissions from the bitfield.
func (p *Permission) Remove(permissions Permission) {
	*p &= ^permissions
}

// Has checks if the bitfield contains ANY of the given permissions.
func (p Permission) Has(permissions Permission) bool {
	return (p & permissions) != 0
}

// HasAll checks if the bitfield contains ALL of the given permissions.
func (p Permission) HasAll(permissions Permission) bool {
	return (p & permissions) == permissions
}

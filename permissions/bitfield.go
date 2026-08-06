package permissions

// Permission represents a 64-bit integer containing multiple Discord permission flags.
type Permission uint64

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

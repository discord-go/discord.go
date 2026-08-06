package permissions

// Build creates a new Permission bitfield from multiple permission flags.
func Build(permissions ...Permission) Permission {
	var p Permission
	for _, perm := range permissions {
		p |= perm
	}
	return p
}

// Builder is a struct to chain permission building.
type Builder struct {
	perm Permission
}

// NewBuilder creates a new Builder for chaining permission flags.
func NewBuilder(initial ...Permission) *Builder {
	b := &Builder{}
	b.Add(initial...)
	return b
}

// Add adds permissions and returns the builder for chaining.
func (b *Builder) Add(permissions ...Permission) *Builder {
	for _, p := range permissions {
		b.perm |= p
	}
	return b
}

// Remove removes permissions and returns the builder for chaining.
func (b *Builder) Remove(permissions ...Permission) *Builder {
	for _, p := range permissions {
		b.perm &= ^p
	}
	return b
}

// Build returns the final Permission bitfield.
func (b *Builder) Build() Permission {
	return b.perm
}

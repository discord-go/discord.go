package intents

// Add adds the given intents to the current intent.
func (i *Intent) Add(intent Intent) {
	*i |= intent
}

// Remove removes the given intents from the current intent.
func (i *Intent) Remove(intent Intent) {
	*i &^= intent
}

// Has checks if the current intent has the given intent.
func (i Intent) Has(intent Intent) bool {
	return i&intent == intent
}

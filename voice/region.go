package voice

// VoiceRegion represents a Discord voice region.
type VoiceRegion struct {
	// ID is the unique identifier for the region.
	ID string `json:"id"`
	// Name is the human-readable name of the region.
	Name string `json:"name"`
	// Optimal indicates whether this is the closest region to the client.
	Optimal bool `json:"optimal"`
	// Deprecated indicates whether this is a deprecated voice region.
	Deprecated bool `json:"deprecated"`
	// Custom indicates whether this is a custom voice region (used for events, etc.).
	Custom bool `json:"custom"`
}

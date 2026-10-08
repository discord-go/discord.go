package gateway

import (
	"github.com/discord-go/discord.go/intents"
)

// Capability is a bit in the Identify payload's capabilities field. The
// capabilities bitfield opts a bot client into gateway behaviors; it is
// separate from intents, which select delivered event families.
type Capability int

const (
	// CapabilityChannelObfuscation opts the client into receiving obfuscated
	// channel metadata over the Gateway for channels it cannot view. Discord
	// documents this as a temporary, testing-only opt-in for channel
	// obfuscation: the mechanism will change before the feature reaches
	// general availability, after which obfuscation applies to all bots
	// automatically. The same bit can also be enabled through the
	// developer portal's Private Channel Obfuscation toggle.
	CapabilityChannelObfuscation Capability = 1 << 15
)

// IdentifyProperties represents the properties of an Identify payload.
type IdentifyProperties struct {
	OS      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

// Identify represents the Identify payload.
type Identify struct {
	Token          string             `json:"token"`
	Properties     IdentifyProperties `json:"properties"`
	Compress       bool               `json:"compress,omitempty"`
	LargeThreshold int                `json:"large_threshold,omitempty"`
	Shard          []int              `json:"shard,omitempty"`
	Presence       interface{}        `json:"presence,omitempty"`
	Intents        intents.Intent     `json:"intents"`
	Capabilities   Capability         `json:"capabilities,omitempty"`
}

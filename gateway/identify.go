package gateway

import (
	"github.com/discord-go/discord.go/intents"
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
}

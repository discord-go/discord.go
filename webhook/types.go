package webhook

// Type represents the type of a webhook.
type Type int

const (
	// TypeIncoming is an incoming webhook that can be used to send messages.
	TypeIncoming Type = 1
	// TypeChannelFollower is a channel follower webhook that cross-posts
	// from a source channel.
	TypeChannelFollower Type = 2
	// TypeApplication is an application webhook used for interactions.
	TypeApplication Type = 3
)

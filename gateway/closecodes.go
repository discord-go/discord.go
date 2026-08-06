package gateway

// CloseCode represents a Discord gateway close event code.
type CloseCode int

const (
	CloseCodeUnknownError         CloseCode = 4000
	CloseCodeUnknownOpcode        CloseCode = 4001
	CloseCodeDecodeError          CloseCode = 4002
	CloseCodeNotAuthenticated     CloseCode = 4003
	CloseCodeAuthenticationFailed CloseCode = 4004
	CloseCodeAlreadyAuthenticated CloseCode = 4005
	CloseCodeInvalidSeq           CloseCode = 4007
	CloseCodeRateLimited          CloseCode = 4008
	CloseCodeSessionTimedOut      CloseCode = 4009
	CloseCodeInvalidShard         CloseCode = 4010
	CloseCodeShardingRequired     CloseCode = 4011
	CloseCodeInvalidAPIVersion    CloseCode = 4012
	CloseCodeInvalidIntents       CloseCode = 4013
	CloseCodeDisallowedIntents    CloseCode = 4014
)

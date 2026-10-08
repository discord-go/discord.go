package gateway

// Opcode represents a Discord gateway opcode.
type Opcode int

const (
	OpcodeDispatch            Opcode = 0
	OpcodeHeartbeat           Opcode = 1
	OpcodeIdentify            Opcode = 2
	OpcodePresenceUpdate      Opcode = 3
	OpcodeVoiceStateUpdate    Opcode = 4
	OpcodeResume              Opcode = 6
	OpcodeReconnect           Opcode = 7
	OpcodeRequestGuildMembers Opcode = 8
	OpcodeInvalidSession      Opcode = 9
	OpcodeHello               Opcode = 10
	OpcodeHeartbeatACK        Opcode = 11
	// OpcodeRequestChannelInfo asks Discord for a Channel Info event carrying
	// ephemeral per-channel values (voice channel status, voice start time)
	// that are not present on the channel object (opcode 43, added 2026).
	OpcodeRequestChannelInfo Opcode = 43
)

package voice

type SpeakingFlag uint8

const (
	Microphone SpeakingFlag = 1
	Soundshare SpeakingFlag = 2
	Priority   SpeakingFlag = 4
)

type SpeakingPayload struct {
	Speaking SpeakingFlag `json:"speaking"`
	Delay    int          `json:"delay"`
	SSRC     uint32       `json:"ssrc"`
}

package gateway

// Resume represents a Resume payload.
type Resume struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int    `json:"seq"`
}

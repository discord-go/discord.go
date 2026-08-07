package interactions

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// InteractionHandler is called for each verified interaction. The handler
// receives the parsed interaction and must return an InteractionResponse
// that will be JSON-encoded as the HTTP response body. For a ping (type 1)
// the server automatically responds with a pong unless the handler returns
// a non-nil response.
type InteractionHandler func(interaction *Interaction) *InteractionResponse

// Server is an http.Handler that receives Discord interaction webhooks,
// verifies their Ed25519 signature and timestamp freshness, and dispatches
// them to a user-provided handler.
//
// Create one with NewServer and register it on your http.ServeMux:
//
//	srv := interactions.NewServer(publicKey, func(i *interactions.Interaction) *interactions.InteractionResponse {
//	    // handle interaction
//	    return nil // let the server auto-respond
//	})
//	http.Handle("/interactions", srv)
type Server struct {
	publicKey string
	handler   InteractionHandler
	now       func() time.Time
}

// NewServer creates a new interaction webhook server. publicKey is the
// hex-encoded Ed25519 public key from the Discord Developer Portal.
func NewServer(publicKey string, handler InteractionHandler) *Server {
	return &Server{
		publicKey: publicKey,
		handler:   handler,
		now:       time.Now,
	}
}

// ServeHTTP implements http.Handler. It reads the request body, verifies
// the signature and timestamp, and dispatches to the handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	timestamp := r.Header.Get("X-Signature-Timestamp")
	signature := r.Header.Get("X-Signature-Ed25519")
	if timestamp == "" || signature == "" {
		http.Error(w, "missing signature headers", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	if !VerifyRequest(s.publicKey, timestamp, signature, body, s.now()) {
		http.Error(w, "invalid request signature", http.StatusUnauthorized)
		return
	}

	var interaction Interaction
	if err := json.Unmarshal(body, &interaction); err != nil {
		http.Error(w, "failed to parse interaction", http.StatusBadRequest)
		return
	}

	// Auto-respond to pings so the bot appears online in the developer portal.
	if interaction.Type == InteractionTypePing {
		respond(w, &InteractionResponse{Type: InteractionCallbackTypePong})
		return
	}

	resp := s.handler(&interaction)
	if resp == nil {
		// Discord requires a response within 3 seconds. If the handler
		// returns nil, acknowledge with a deferred response so the
		// interaction does not time out.
		resp = &InteractionResponse{Type: InteractionCallbackTypeDeferredChannelMessageWithSource}
	}

	respond(w, resp)
}

func respond(w http.ResponseWriter, resp *InteractionResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

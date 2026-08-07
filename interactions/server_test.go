package interactions

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServerRejectsNonPost(t *testing.T) {
	srv := NewServer("whatever", func(i *Interaction) *InteractionResponse { return nil })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestServerRejectsMissingHeaders(t *testing.T) {
	srv := NewServer("whatever", func(i *Interaction) *InteractionResponse { return nil })
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestServerRejectsInvalidSignature(t *testing.T) {
	srv := NewServer("0000000000000000000000000000000000000000000000000000000000000000", func(i *Interaction) *InteractionResponse { return nil })
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set("X-Signature-Timestamp", "1700000000")
	req.Header.Set("X-Signature-Ed25519", "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestServerHandlesPing(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	publicKeyHex := hex.EncodeToString(publicKey)

	now := time.Unix(1700000000, 0)
	srv := NewServer(publicKeyHex, func(i *Interaction) *InteractionResponse { return nil })
	srv.now = func() time.Time { return now }

	body := []byte(`{"type":1,"token":"test","version":1}`)
	timestamp := "1700000000"
	message := append([]byte(timestamp), body...)
	signature := ed25519.Sign(privateKey, message)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("X-Signature-Timestamp", timestamp)
	req.Header.Set("X-Signature-Ed25519", hex.EncodeToString(signature))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp InteractionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Type != InteractionCallbackTypePong {
		t.Fatalf("expected pong (type 1), got type %d", resp.Type)
	}
}

func TestServerDispatchesToHandler(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	publicKeyHex := hex.EncodeToString(publicKey)

	now := time.Unix(1700000000, 0)
	var received *Interaction
	srv := NewServer(publicKeyHex, func(i *Interaction) *InteractionResponse {
		received = i
		return &InteractionResponse{
			Type: InteractionCallbackTypeChannelMessageWithSource,
			Data: &InteractionCallbackData{Content: "Hello!"},
		}
	})
	srv.now = func() time.Time { return now }

	body := []byte(`{"type":2,"token":"test","version":1,"data":{"name":"ping"}}`)
	timestamp := "1700000000"
	message := append([]byte(timestamp), body...)
	signature := ed25519.Sign(privateKey, message)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("X-Signature-Timestamp", timestamp)
	req.Header.Set("X-Signature-Ed25519", hex.EncodeToString(signature))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if received == nil {
		t.Fatal("handler was not called")
	}
	if received.Type != InteractionTypeApplicationCommand {
		t.Fatalf("expected type 2, got %d", received.Type)
	}

	var resp InteractionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Type != InteractionCallbackTypeChannelMessageWithSource {
		t.Fatalf("expected type 4, got %d", resp.Type)
	}
	if resp.Data == nil || resp.Data.Content != "Hello!" {
		t.Fatalf("unexpected response data: %+v", resp.Data)
	}
}

func TestServerDefaultsToDeferredWhenHandlerReturnsNil(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	publicKeyHex := hex.EncodeToString(publicKey)

	now := time.Unix(1700000000, 0)
	srv := NewServer(publicKeyHex, func(i *Interaction) *InteractionResponse { return nil })
	srv.now = func() time.Time { return now }

	body := []byte(`{"type":2,"token":"test","version":1}`)
	timestamp := "1700000000"
	message := append([]byte(timestamp), body...)
	signature := ed25519.Sign(privateKey, message)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("X-Signature-Timestamp", timestamp)
	req.Header.Set("X-Signature-Ed25519", hex.EncodeToString(signature))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp InteractionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Type != InteractionCallbackTypeDeferredChannelMessageWithSource {
		t.Fatalf("expected deferred (type 5), got type %d", resp.Type)
	}
}

func TestServerRejectsReplay(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	publicKeyHex := hex.EncodeToString(publicKey)

	// The request was signed at t=1700000000, but the server's clock is
	// 10 minutes later — past the MaxTimestampSkew window.
	srv := NewServer(publicKeyHex, func(i *Interaction) *InteractionResponse { return nil })
	srv.now = func() time.Time { return time.Unix(1700000600, 0) }

	body := []byte(`{"type":1,"token":"test","version":1}`)
	timestamp := "1700000000"
	message := append([]byte(timestamp), body...)
	signature := ed25519.Sign(privateKey, message)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	req.Header.Set("X-Signature-Timestamp", timestamp)
	req.Header.Set("X-Signature-Ed25519", hex.EncodeToString(signature))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for replay, got %d", w.Code)
	}

	// Ensure the body was fully consumed so the connection can be reused.
	resp := w.Result()
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
}

package gateway

import (
	"sync"
	"sync/atomic"
)

// Session tracks the state of a gateway session for resuming connections.
// It is safe for concurrent use.
type Session struct {
	mu        sync.RWMutex
	sessionID string
	resumeURL string
	sequence  atomic.Int64
}

// NewSession creates a new empty Session.
func NewSession() *Session {
	return &Session{}
}

// SetSessionID sets the session ID received from the READY event.
func (s *Session) SetSessionID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionID = id
}

// SessionID returns the current session ID.
func (s *Session) SessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionID
}

// SetResumeURL sets the resume gateway URL received from the READY event.
func (s *Session) SetResumeURL(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resumeURL = url
}

// ResumeURL returns the resume gateway URL.
func (s *Session) ResumeURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.resumeURL
}

// UpdateSequence updates the latest sequence number from a dispatch event.
func (s *Session) UpdateSequence(seq int64) {
	s.sequence.Store(seq)
}

// Sequence returns the last received sequence number.
func (s *Session) Sequence() int64 {
	return s.sequence.Load()
}

// Reset clears all session state. Called when a session is invalidated
// and a fresh identify is required.
func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionID = ""
	s.resumeURL = ""
	s.sequence.Store(0)
}

// CanResume reports whether the session has enough state to attempt a resume.
func (s *Session) CanResume() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionID != "" && s.resumeURL != ""
}

// ToResume creates a Resume payload from the current session state.
func (s *Session) ToResume(token string) Resume {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Resume{
		Token:     token,
		SessionID: s.sessionID,
		Seq:       int(s.sequence.Load()),
	}
}

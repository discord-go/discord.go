// Package session implements a DAVE session for the Discord voice client.
// It processes MLS events and encrypts and decrypts media frames.
//
// A session is intended for one voice connection. Call [Session.Close] when
// it is discarded to stop its background recovery work. [Session.Ready] and
// [Session.WaitReady] report encryption readiness; [Session.ShouldHoldFrames]
// distinguishes a pending handshake from an unencrypted protocol session.
// [CreateFunc] and [WithSessionHook] connect sessions to a voice manager.
package session

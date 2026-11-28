package compression

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"fmt"
	"io"
)

// NewReader creates a new zlib decompression reader that wraps the provided
// io.Reader. The returned io.ReadCloser decompresses data on the fly.
//
// The caller is responsible for closing the returned reader when done.
//
// This is used by the Gateway client to decompress zlib-compressed payloads
// sent by Discord when transport compression is enabled.
func NewReader(r io.Reader) (io.ReadCloser, error) {
	if r == nil {
		return nil, fmt.Errorf("compression: reader must not be nil")
	}

	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("compression: failed to create zlib reader: %w", err)
	}

	return zr, nil
}

// Decompress reads all data from r, decompresses it via zlib, and returns
// the decompressed bytes. It is a convenience wrapper around NewReader for
// cases where the entire payload is available in memory.
func Decompress(r io.Reader) ([]byte, error) {
	zr, err := NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("compression: failed to read decompressed data: %w", err)
	}

	return data, nil
}

// A DEFLATE back-reference can reach at most 32 KiB into the decoded stream.
const historySize = 32 * 1024

// Stream handles stateful zlib decompression for one Discord connection.
// It buffers compressed chunks until the Z_SYNC_FLUSH suffix arrives.
type Stream struct {
	pending  bytes.Buffer
	history  []byte
	inflater io.ReadCloser
}

// NewStream creates a new stateful Stream.
func NewStream() *Stream {
	return &Stream{}
}

// Write appends a compressed chunk to the stream.
// If the chunk ends with the sync flush suffix (00 00 ff ff), it decompresses
// the available data and returns the newly decompressed payload.
func (s *Stream) Write(chunk []byte) ([]byte, error) {
	s.pending.Write(chunk)

	if !bytes.HasSuffix(s.pending.Bytes(), []byte{0x00, 0x00, 0xff, 0xff}) {
		return nil, nil
	}

	reader, err := s.payloadReader()
	if err != nil {
		return nil, fmt.Errorf("compression: failed to start stream payload: %w", err)
	}

	payload, err := io.ReadAll(reader)
	_ = reader.Close()
	s.pending.Reset()
	// A sync flush ends the block, while the DEFLATE stream is still open.
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("compression: stream read error: %w", err)
	}

	if s.history == nil {
		s.history = make([]byte, 0, historySize)
	}
	recent := payload[max(0, len(payload)-historySize):]
	if drop := len(s.history) + len(recent) - historySize; drop > 0 {
		copy(s.history, s.history[drop:])
		s.history = s.history[:len(s.history)-drop]
	}
	s.history = append(s.history, recent...)
	return payload, nil
}

func (s *Stream) payloadReader() (io.ReadCloser, error) {
	if s.history == nil {
		return zlib.NewReader(&s.pending)
	}
	// The first payload consumed the zlib header. Each sync flush ends on a
	// byte boundary, so later blocks can resume from the saved dictionary.
	if s.inflater == nil {
		s.inflater = flate.NewReaderDict(&s.pending, s.history)
		return s.inflater, nil
	}
	return s.inflater, s.inflater.(flate.Resetter).Reset(&s.pending, s.history)
}

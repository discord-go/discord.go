package compression

import (
	"bytes"
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

// Stream handles stateful zlib decompression for Discord's stream compression.
// It buffers compressed chunks and inflates them when the Z_SYNC_FLUSH suffix is received.
type Stream struct {
	buf        bytes.Buffer
	readOffset int64
}

// NewStream creates a new stateful Stream.
func NewStream() *Stream {
	return &Stream{}
}

// Write appends a compressed chunk to the stream.
// If the chunk ends with the sync flush suffix (00 00 ff ff), it decompresses
// the available data and returns the newly decompressed payload.
func (s *Stream) Write(chunk []byte) ([]byte, error) {
	s.buf.Write(chunk)

	if len(chunk) < 4 || !bytes.HasSuffix(chunk, []byte{0x00, 0x00, 0xff, 0xff}) {
		return nil, nil // Need more data
	}

	zr, err := zlib.NewReader(bytes.NewReader(s.buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("compression: failed to create stream zlib reader: %w", err)
	}
	defer zr.Close()

	all, err := io.ReadAll(zr)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("compression: stream read error: %w", err)
	}

	if int64(len(all)) < s.readOffset {
		return nil, fmt.Errorf("compression: stream invalid read offset")
	}

	ret := all[s.readOffset:]
	s.readOffset = int64(len(all))
	return ret, nil
}

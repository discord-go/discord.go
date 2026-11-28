package compression

import (
	"bytes"
	"compress/zlib"
	"io"
	"math/rand"
	"strings"
	"testing"
)

// zlibCompress is a test helper that zlib-compresses data.
func zlibCompress(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		t.Fatalf("failed to write zlib data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zlib writer: %v", err)
	}
	return buf.Bytes()
}

func TestNewReader_ValidData(t *testing.T) {
	original := []byte("hello, discord gateway!")
	compressed := zlibCompress(t, original)

	reader, err := NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !bytes.Equal(got, original) {
		t.Errorf("decompressed = %q, want %q", got, original)
	}
}

func TestNewReader_NilReader(t *testing.T) {
	reader, err := NewReader(nil)
	if err == nil {
		t.Fatal("NewReader(nil) should return an error")
	}
	if reader != nil {
		t.Error("NewReader(nil) should return nil reader")
	}
	if !strings.Contains(err.Error(), "must not be nil") {
		t.Errorf("error message should mention nil: %v", err)
	}
}

func TestNewReader_InvalidData(t *testing.T) {
	reader, err := NewReader(bytes.NewReader([]byte("not zlib data")))
	if err == nil {
		if reader != nil {
			reader.Close()
		}
		t.Fatal("NewReader with invalid data should return an error")
	}
	if !strings.Contains(err.Error(), "failed to create zlib reader") {
		t.Errorf("error should mention zlib reader creation: %v", err)
	}
}

func TestNewReader_EmptyCompressed(t *testing.T) {
	original := []byte("")
	compressed := zlibCompress(t, original)

	reader, err := NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected empty result, got %q", got)
	}
}

func TestNewReader_LargePayload(t *testing.T) {
	// Simulate a large Gateway payload.
	original := bytes.Repeat([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"), 10000)
	compressed := zlibCompress(t, original)

	reader, err := NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !bytes.Equal(got, original) {
		t.Errorf("large payload: decompressed length = %d, want %d", len(got), len(original))
	}
}

func TestDecompress_ValidData(t *testing.T) {
	original := []byte(`{"op":0,"d":{"content":"test"},"s":1,"t":"MESSAGE_CREATE"}`)
	compressed := zlibCompress(t, original)

	got, err := Decompress(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("Decompress() error = %v", err)
	}

	if !bytes.Equal(got, original) {
		t.Errorf("Decompress() = %q, want %q", got, original)
	}
}

func TestDecompress_NilReader(t *testing.T) {
	_, err := Decompress(nil)
	if err == nil {
		t.Fatal("Decompress(nil) should return an error")
	}
}

func TestDecompress_InvalidData(t *testing.T) {
	_, err := Decompress(bytes.NewReader([]byte("garbage")))
	if err == nil {
		t.Fatal("Decompress with invalid data should return an error")
	}
}

func streamChunks(t testing.TB, payloads [][]byte) [][]byte {
	t.Helper()
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	defer w.Close()
	chunks := make([][]byte, len(payloads))
	for i, payload := range payloads {
		if _, err := w.Write(payload); err != nil {
			t.Fatal(err)
		}
		if err := w.Flush(); err != nil {
			t.Fatal(err)
		}
		chunks[i] = bytes.Clone(buf.Bytes())
		buf.Reset()
	}
	return chunks
}

func TestStreamRetainsDictionaryAcrossSplitFlush(t *testing.T) {
	large := make([]byte, historySize+1024)
	_, _ = rand.New(rand.NewSource(1)).Read(large)
	repeat := bytes.Clone(large[len(large)-16*1024:])
	payloads := [][]byte{large, repeat, nil, repeat, []byte("final payload")}
	chunks := streamChunks(t, payloads)
	if len(chunks[1]) >= len(repeat)/2 {
		t.Fatal("second payload did not use the preceding payload as a dictionary")
	}
	s := NewStream()
	for i, chunk := range chunks {
		split := len(chunk) - 2
		if data, err := s.Write(chunk[:split]); data != nil || err != nil {
			t.Fatalf("chunk %d before flush: %d bytes, error %v", i, len(data), err)
		}
		data, err := s.Write(chunk[split:])
		if err != nil || !bytes.Equal(data, payloads[i]) {
			t.Fatalf("chunk %d after flush: %d bytes, error %v", i, len(data), err)
		}
		// The returned slice belongs to the caller.
		clear(data)
	}
}

func BenchmarkStream(b *testing.B) {
	payload := make([]byte, 1024)
	_, _ = rand.New(rand.NewSource(1)).Read(payload)
	payloads := make([][]byte, 128)
	for i := range payloads {
		payloads[i] = payload
	}
	chunks := streamChunks(b, payloads)
	b.SetBytes(int64(len(payload) * len(chunks)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewStream()
		for _, chunk := range chunks {
			if data, err := s.Write(chunk); err != nil || len(data) != len(payload) {
				b.Fatalf("got %d bytes, error %v", len(data), err)
			}
		}
	}
}

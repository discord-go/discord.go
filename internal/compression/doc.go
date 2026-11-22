// Package compression provides zlib stream decompression helpers for Gateway payloads.
//
// The Discord Gateway can send zlib-compressed payloads; this package wraps
// the standard library's compress/zlib to provide a convenient reader.
package compression

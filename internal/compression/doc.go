// Package compression provides zlib stream decompression helpers for Gateway payloads.
//
// This is an internal package and MUST NOT be imported by external consumers.
// The Discord Gateway can send zlib-compressed payloads; this package wraps
// the standard library's compress/zlib to provide a convenient reader.
package compression

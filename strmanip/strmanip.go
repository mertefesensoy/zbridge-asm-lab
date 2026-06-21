// Package strmanip demonstrates string-header access and length-prefixed
// buffer construction in Plan 9 assembly on amd64. It is exercise 3 of
// zbridge-asm-lab; see the repo root README for the broader context and the
// roadmap toward WTO via SVC 35 from Go assembly on z/OS.
//
// The distinctive skill over the ebcdic exercise is the Go string header.
// A string is a two-word header (data pointer, length), unlike a slice's
// three-word header (data, length, capacity), so the argument offsets in the
// assembly differ.
//
// WrapLengthPrefixed rehearses the WTO parameter list shape: a 2-byte length
// field followed by the message text. The length is stored big-endian, which
// matches z/Architecture byte order, so the on-storage layout is identical on
// s390x.
package strmanip

// StrLen returns the length of s by reading the length word of the Go string
// header directly in assembly. It is equivalent to len(s); it exists to
// isolate string-header access as its own step before WrapLengthPrefixed uses
// the same access pattern.
//
//go:noescape
func StrLen(s string) int

// WrapLengthPrefixed writes a 2-byte big-endian length header followed by the
// bytes of msg into dst, and returns the total number of bytes written, which
// is len(msg)+2.
//
// dst must have capacity for at least len(msg)+2 bytes. msg must be at most
// 65535 bytes, the maximum the 2-byte length field can hold; longer input
// truncates the stored length and is the caller's responsibility to avoid.
//
// The output shape mirrors the WTO parameter list: a length field followed by
// the (EBCDIC) message text. The EBCDIC conversion itself is the ebcdic
// exercise; this exercise builds the surrounding length-prefixed buffer.
//
//go:noescape
func WrapLengthPrefixed(dst []byte, msg string) int

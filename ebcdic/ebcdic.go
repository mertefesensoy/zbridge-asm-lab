// Package ebcdic provides byte-level conversion between
// ISO-8859-1 (treated as ASCII-compatible by the lookup tables)
// and EBCDIC IBM-1047, the code page used by z/OS USS.
//
// The conversion is table-driven and implemented in Plan 9
// assembly on amd64. The s390x port collapses the lookup loop
// into a single TR instruction; see ebcdic_s390x.s.
//
// This is exercise 2 of zbridge-asm-lab; see the repo root README
// for the broader context and the roadmap toward WTO via SVC 35
// from Go assembly on z/OS.
package ebcdic

// AtoE translates len(src) bytes from src (ISO-8859-1) into
// EBCDIC IBM-1047, writing the result to dst. dst must have
// at least len(src) capacity. Bytes outside the printable
// range are mapped per the IBM-1047 table; no validation is
// performed.
//
//go:noescape
func AtoE(dst, src []byte)

// EtoA is the inverse of AtoE: EBCDIC IBM-1047 to ISO-8859-1.
//
//go:noescape
func EtoA(dst, src []byte)

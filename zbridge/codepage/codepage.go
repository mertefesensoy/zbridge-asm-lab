// Package codepage converts between ISO-8859-1 (treated as ASCII-compatible by
// the lookup tables) and EBCDIC IBM-1047, the default code page on z/OS USS.
//
// It is the one package in this module that is finished, verified, and useful
// on its own. Every other package here waits on an unretired unknown; this one
// does not, because character conversion depends on a documented table rather
// than on an undocumented parameter-list layout.
//
// # Implementation
//
// Table-driven, in Plan 9 assembly on both supported architectures:
//
//	amd64  a byte-at-a-time lookup loop
//	s390x  MVC to copy, then TR (Translate) to translate a 256-byte block in
//	       one instruction, with EXRL applying a runtime length to the tail
//
// TR is emitted as literal bytes because Go's s390x assembler has no TR
// mnemonic — 729 mnemonics in cmd/internal/obj/s390x/anames.go and the translate
// family is absent. The encoding was confirmed three ways: GNU objdump decodes
// the bytes as "tr 0(256,%r2),0(%r3)", a differential test compares the TR path
// against a byte-loop implementation at every block boundary, and a third test
// checks all 256 table entries directly. See
// docs/evidence/E-L-s390x-port-qemu-2026-07-25.md.
//
// Note that TR translates IN PLACE, so a two-buffer API costs MVC then TR — two
// block instructions per 256 bytes, not the single instruction the project's
// roadmap describes. The per-byte loop is genuinely gone; "one instruction" is
// not literally true.
//
// # Provenance and licence
//
// The tables derive from ibmruntimes/go-recordio (BSD-3-Clause) and were
// verified against the ICU mapping file ibm-1047_P100-1995.ucm from
// unicode-org/icu. Attribution lives in LICENSES.md in this directory and must
// survive any refactor.
package codepage

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

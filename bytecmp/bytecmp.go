// Package bytecmp demonstrates byte-sequence comparison in Plan 9 assembly on
// amd64: memory loads from two pointers, a counted loop, and branching on the
// condition codes a byte compare sets. It is exercise 5 of zbridge-asm-lab;
// see the repo root README for the broader context and the roadmap toward WTO
// via SVC 35 from Go assembly on z/OS.
//
// Both functions match the semantics of their standard-library equivalents,
// bytes.Equal and bytes.Compare, which makes the standard library a ready
// oracle for the tests and a baseline for the benchmark.
//
// On s390x both functions collapse onto the CLC instruction (Compare Logical
// Characters), which compares two storage operands and sets the condition code
// in one operation; see bytecmp_s390x.s.
package bytecmp

// Equal reports whether a and b are the same length and hold the same bytes.
// It matches bytes.Equal.
//
//go:noescape
func Equal(a, b []byte) bool

// Compare returns an integer comparing a and b lexicographically: 0 if they
// are equal, -1 if a sorts before b, and +1 if a sorts after b. Bytes are
// compared as unsigned values, and if one slice is a prefix of the other the
// shorter slice sorts first. It matches bytes.Compare.
//
//go:noescape
func Compare(a, b []byte) int

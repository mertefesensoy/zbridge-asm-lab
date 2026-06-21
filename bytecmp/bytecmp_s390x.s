// STUB - not implemented. This file documents the s390x port path for the
// Equal and Compare functions.
//
// Both collapse onto a single instruction family, the same way the ebcdic
// lookup loop collapses into TR: CLC, Compare Logical Characters.
//
// CLC compares two storage operands byte by byte, left to right, and sets the
// condition code:
//
//   CC = 0   operands equal
//   CC = 1   first operand low  (sorts before)
//   CC = 2   first operand high (sorts after)
//
// So the byte-at-a-time loop and the condition-code branching here become one
// CLC plus a branch on CC:
//
//   Equal:   after the length check, CLC the two buffers; equal iff CC == 0.
//   Compare: CLC over min(len(a), len(b)) bytes, then branch on CC for the
//            ordering; if CC == 0 the common prefix matched, so decide by
//            length, exactly as the amd64 bylen path does.
//
// CLC is an SS-format instruction whose length is encoded in the instruction
// text, so it handles up to 256 bytes directly. For a length known only at
// run time, the standard trick is EX / EXRL to patch the length byte, or a
// loop issuing CLC in 256-byte strides for longer buffers.
//
// Implementation is deferred until Phase 1b of the roadmap, when LinuxONE
// Community Cloud access enables real s390x testing.
//
// The stub TEXT directives below cause a "not implemented" build failure if
// this file is accidentally targeted; this is intentional.

#include "textflag.h"

TEXT ·Equal(SB), NOSPLIT, $0-49
	UNDEF

TEXT ·Compare(SB), NOSPLIT, $0-56
	UNDEF

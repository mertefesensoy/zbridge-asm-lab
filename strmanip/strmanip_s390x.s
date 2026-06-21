// STUB - not implemented. This file documents the s390x port path for the
// StrLen and WrapLengthPrefixed functions.
//
// Two amd64 details collapse on s390x, for the same reason ebcdic's lookup
// loop collapses into a single TR instruction: the hardware does the work.
//
//  1. The big-endian length header.
//     On amd64 (little-endian) the 16-bit length is split into a high byte and
//     a low byte by hand (SHRQ + two MOVB stores). On s390x the architecture
//     is big-endian, so a Store Halfword (STH, Go asm MOVH) writes the two
//     bytes in big-endian order in ONE instruction. No manual split.
//
//  2. The byte copy loop.
//     On amd64 the message bytes are copied one at a time in a loop. On s390x
//     the MVC instruction (Move Characters) copies up to 256 bytes in one
//     operation, so the loop collapses for short messages, which is every
//     single-line WTO message. (Longer buffers issue MVC in a loop with
//     256-byte strides.)
//
// String-header access is identical in shape: a string is still a two-word
// header (data pointer, length); only the load instruction name differs
// (MOVD on s390x).
//
// Implementation is deferred until Phase 1b of the roadmap, when LinuxONE
// Community Cloud access enables real s390x testing.
//
// The stub TEXT directives below cause a "not implemented" build failure if
// this file is accidentally targeted; this is intentional.

#include "textflag.h"

TEXT ·StrLen(SB), NOSPLIT, $0-24
	UNDEF

TEXT ·WrapLengthPrefixed(SB), NOSPLIT, $0-48
	UNDEF

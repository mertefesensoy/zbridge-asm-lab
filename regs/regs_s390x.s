// STUB - not implemented. This file documents the s390x port path for the
// GetSP, GetBP, and GetFramedSP functions.
//
// The register file is the headline difference. amd64 has named registers
// (AX, BX, SP, BP, ...). s390x has sixteen general registers R0 through R15
// and sixteen floating registers F0 through F15, and the calling convention
// assigns fixed roles to several of them:
//
//   R15  Stack pointer. This is the analog of amd64's SP, so GetSP becomes a
//        read of R15 (Go asm: MOVD R15, ...).
//   R14  Return address (link register). A call via BASR or BRASL leaves the
//        return point here, where amd64 places it on the stack.
//   R13  Save-area / base register in standard MVS linkage. There is no single
//        "frame pointer" register the way amd64 uses BP; the equivalent role
//        is carried by the save-area chain anchored in R13.
//
// These are exactly the registers the WTO path will touch: parameters go in
// R0 and R1, SVC 35 is issued, and the return code is read from R15 after the
// call (R15 doubles as the return-code register on return from a service).
//
// GetFramedSP maps the same way as GetSP (read R15); the frame is allocated by
// the s390x prologue rather than the amd64 one, but the observation that a
// declared frame lowers the stack pointer is identical.
//
// Implementation is deferred until Phase 1b of the roadmap, when LinuxONE
// Community Cloud access enables real s390x testing.
//
// The stub TEXT directives below cause a "not implemented" build failure if
// this file is accidentally targeted; this is intentional.

#include "textflag.h"

TEXT ·GetSP(SB), NOSPLIT, $0-8
	UNDEF

TEXT ·GetBP(SB), NOSPLIT, $0-8
	UNDEF

TEXT ·GetFramedSP(SB), NOSPLIT, $256-8
	UNDEF

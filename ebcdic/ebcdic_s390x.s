// STUB - not implemented. This file documents the s390x port
// path for the AtoE and EtoA functions.
//
// On s390x, the entire amd64 lookup loop above collapses into
// a single hardware instruction: TR (Translate), opcode DC.
//
//   TR  D1(L,B1),D2(B2)
//
// The TR instruction translates up to 256 bytes in one operation
// by using the source buffer as an index into a 256-byte
// translation table. For longer buffers, TR is issued in a loop
// with 256-byte strides, or a TRT-prefixed variant is used for
// length-aware variants. This is the same pattern go-recordio
// uses in utils/utils.s for its AtoE/EtoA helpers.
//
// Implementation is deferred until Phase 1b of the roadmap, when
// LinuxONE Community Cloud access enables real s390x testing.
//
// Stub TEXT directives below cause "not implemented" build
// failure if accidentally targeted; this is intentional.

#include "textflag.h"

TEXT ·AtoE(SB), NOSPLIT, $0-48
	UNDEF

TEXT ·EtoA(SB), NOSPLIT, $0-48
	UNDEF

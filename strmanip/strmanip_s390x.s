// s390x implementation of StrLen and WrapLengthPrefixed.
//
// This file replaces the former UNDEF stub. Two amd64 details collapse here,
// and one of them is the single most direct piece of evidence in this repo that
// the target architecture is a better fit for the WTO parameter list than the
// machine we develop on.
//
// ---------------------------------------------------------------------------
// 1. The big-endian length header is one instruction.
// ---------------------------------------------------------------------------
//
// amd64 is little-endian, so strmanip_amd64.s builds the 16-bit header by hand:
//
//	MOVQ	CX, AX
//	SHRQ	$8, AX
//	MOVB	AX, (DI)	// high byte first
//	MOVB	CX, 1(DI)	// then low byte
//
// z/Architecture is big-endian, so a halfword store already puts the high byte
// at the lower address. Four instructions become one:
//
//	MOVH	R5, 0(R2)
//
// MOVH is the Go assembler's spelling of Store Halfword (STH). It stores the
// low 16 bits of the register, which is the same truncation the amd64 path
// performs, so lengths above 65535 wrap identically on both architectures — the
// documented caller responsibility in strmanip.go is unchanged.
//
// This matters beyond the exercise: the WTO parameter list begins with a
// two-byte big-endian length. On z/Architecture that field is written by the
// natural store instruction, with no byte-order code at all.
//
// ---------------------------------------------------------------------------
// 2. The copy loop becomes MVC.
// ---------------------------------------------------------------------------
//
// MVC (Move Characters) copies up to 256 bytes in one SS-format instruction.
// Its length is encoded in the instruction text, so a length known only at run
// time is applied with EXRL (Execute Relative Long), which ORs the low 8 bits
// of a register into the length field. The register therefore holds len-1.
//
// This is the encoding runtime/memmove_s390x.s uses for the same problem, which
// is what H002's threat #1 asks for: prefer constructs attested in the Go
// standard library's own s390x assembly.
//
// Every single-line WTO message fits in one MVC — the maximum WTO text is well
// under 256 bytes — so on the real call path the block loop below never runs a
// second iteration. It is here because WrapLengthPrefixed is documented to
// accept up to 65535 bytes and must be correct at that size too.
//
// ---------------------------------------------------------------------------
// 3. String headers are unchanged.
// ---------------------------------------------------------------------------
//
// A Go string is a two-word header (data pointer, length) on every
// architecture, so the argument offsets are identical to amd64 and only the
// load mnemonic differs: MOVD instead of MOVQ. That non-difference is worth
// stating, because it is the part of the ABI the port did NOT have to change.
//
// Registers (R0 is not usable as a base; R10/R11 REGTMP, R12 REGCTXT, R13 g,
// R14 link, R15 stack pointer):
//
//	R2  destination pointer, then the return value
//	R4  message pointer
//	R5  message length, then bytes remaining
//	R6  scratch: len-1 for EXRL

#include "textflag.h"

// func StrLen(s string) int
//
// Argument layout (FP): s_base+0, s_len+8, ret+16.
// A string is TWO words, so the length is at +8 and there is no capacity word.
TEXT ·StrLen(SB), NOSPLIT|NOFRAME, $0-24
	MOVD	s_len+8(FP), R2
	MOVD	R2, ret+16(FP)
	RET

// func WrapLengthPrefixed(dst []byte, msg string) int
//
// Argument layout (FP):
//	dst []byte (3 words): dst_base+0, dst_len+8, dst_cap+16
//	msg string (2 words): msg_base+24, msg_len+32
//	return int:           ret+40
TEXT ·WrapLengthPrefixed(SB), NOSPLIT|NOFRAME, $0-48
	MOVD	dst_base+0(FP), R2
	MOVD	msg_base+24(FP), R4
	MOVD	msg_len+32(FP), R5

	// The whole big-endian story, in one instruction.
	MOVH	R5, 0(R2)
	LA	2(R2), R2		// advance past the header

	CMPBEQ	R5, $0, done		// empty message: header only

blocks:
	CMP	R5, $256
	BLT	tail
	MVC	$256, 0(R4), 0(R2)
	LA	256(R4), R4
	LA	256(R2), R2
	SUB	$256, R5
	CMPBNE	R5, $0, blocks
	BR	done

tail:
	SUB	$1, R5, R6		// EXRL ORs R6's low byte into the length
					// field, which encodes len-1
	EXRL	$strmanip_exrl_mvc<>(SB), R6

done:
	// Return len(msg)+2. R5 has been consumed by the loop, so re-read it.
	MOVD	msg_len+32(FP), R2
	ADD	$2, R2
	MOVD	R2, ret+40(FP)
	RET

// DO NOT CALL - target for EXRL. NOFRAME is required: with a prologue, EXRL
// would execute the prologue instead of the MVC.
TEXT strmanip_exrl_mvc<>(SB), NOSPLIT|NOFRAME, $0-0
	MVC	$1, 0(R4), 0(R2)	// length byte is patched by EXRL
	MOVD	R0, 0(R0)		// guard: fault rather than fall through
	RET

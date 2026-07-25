// s390x implementation of Equal and Compare.
//
// This file replaces the former UNDEF stub. Of the five ported exercises this
// is the one with the least invention in it: the Go standard library already
// solves exactly this problem on exactly this architecture, and H002's threat
// #1 says to prefer stdlib-attested encodings over ones we reasoned out.
// The structure below deliberately mirrors, and credits, those files:
//
//	internal/bytealg/equal_s390x.s    memeqbody<> / memeqbodyclc<>
//	internal/bytealg/compare_s390x.s  cmpbody<>   / cmpbodyclc<>
//
// The differences from those files are only the ones our exercise requires: we
// keep the ABI0 stack-argument form rather than their ABIInternal register
// form, we do not need their tiny-length special cases, and Compare returns via
// the stack slot rather than R2.
//
// ---------------------------------------------------------------------------
// CLC, and what the condition code buys
// ---------------------------------------------------------------------------
//
// CLC (Compare Logical Characters) compares two storage operands left to right
// and sets the condition code:
//
//	CC = 0	operands equal
//	CC = 1	first operand low   (sorts before)
//	CC = 2	first operand high  (sorts after)
//
// So the amd64 byte-at-a-time loop with its per-byte CMPQ/JNE collapses to one
// instruction per 256-byte block plus a branch. Note what CLC does NOT need: no
// index register, no per-byte load, no loop counter for the compared bytes.
// The comparison IS the instruction.
//
// Go spells it "CLC $length, first, second", and the condition code describes
// the first operand relative to the second — which is why BLT below means
// "a sorts before b".
//
// ---------------------------------------------------------------------------
// EXRL for the tail
// ---------------------------------------------------------------------------
//
// CLC encodes its length in the instruction text, so lengths known only at run
// time go through EXRL (Execute Relative Long), which ORs the low 8 bits of a
// register into the length field. The register holds len-1. Both Equal and
// Compare share one EXRL target because both keep the operands in the same
// registers.
//
// EX (the non-relative form) is not available: Go's s390x assembler has EXRL
// but no EX. EXRL requires z10 or later, which is far below any machine this
// project will run on, but it is a real constraint and worth recording.
//
// Registers (R0 is not usable as a base; R10/R11 REGTMP, R12 REGCTXT, R13 g,
// R14 link, R15 stack pointer):
//
//	R2  a pointer     (CLC first operand)
//	R3  b pointer     (CLC second operand)
//	R4  len(a)
//	R5  len(b), then bytes remaining in Equal
//	R6  bytes remaining in Compare
//	R7  scratch: len-1 for EXRL

#include "textflag.h"

// func Equal(a, b []byte) bool
//
// Argument layout (FP):
//	a []byte: a_base+0,  a_len+8,  a_cap+16
//	b []byte: b_base+24, b_len+32, b_cap+40
//	ret bool: ret+48
TEXT ·Equal(SB), NOSPLIT|NOFRAME, $0-49
	MOVD	a_len+8(FP), R5
	MOVD	b_len+32(FP), R4
	CMPBNE	R5, R4, notequal	// different lengths are never equal

	MOVD	a_base+0(FP), R2
	MOVD	b_base+24(FP), R3

	CMPBEQ	R5, $0, equal		// both empty
	CMPBEQ	R2, R3, equal		// same pointer and same length

eqblocks:
	CMP	R5, $256
	BLT	eqtail
	CLC	$256, 0(R2), 0(R3)
	BNE	notequal
	LA	256(R2), R2
	LA	256(R3), R3
	SUB	$256, R5
	CMPBNE	R5, $0, eqblocks
	BR	equal

eqtail:
	SUB	$1, R5, R7
	EXRL	$bytecmp_exrl_clc<>(SB), R7
	BNE	notequal

equal:
	MOVD	$1, R2
	MOVB	R2, ret+48(FP)
	RET

notequal:
	MOVD	$0, R2
	MOVB	R2, ret+48(FP)
	RET

// func Compare(a, b []byte) int
//
// Argument layout (FP):
//	a []byte: a_base+0,  a_len+8,  a_cap+16
//	b []byte: b_base+24, b_len+32, b_cap+40
//	ret int:  ret+48
//
// Compare the first min(len(a), len(b)) bytes; at the first difference return
// -1 or +1. If the common prefix matches, the shorter slice sorts first. CLC is
// an unsigned (logical) compare, which is the ordering bytes.Compare specifies.
TEXT ·Compare(SB), NOSPLIT|NOFRAME, $0-56
	MOVD	a_base+0(FP), R2
	MOVD	a_len+8(FP), R4
	MOVD	b_base+24(FP), R3
	MOVD	b_len+32(FP), R5

	// R6 = n = min(len(a), len(b))
	MOVD	R4, R6
	CMPBLE	R4, R5, haven
	MOVD	R5, R6

haven:
	CMPBEQ	R6, $0, bylen		// nothing to compare: decide by length
	CMPBEQ	R2, R3, bylen		// same pointer: common prefix is equal

cmpblocks:
	CMP	R6, $256
	BLT	cmptail
	CLC	$256, 0(R2), 0(R3)
	BLT	less			// CC = 1: a low
	BGT	greater			// CC = 2: a high
	LA	256(R2), R2
	LA	256(R3), R3
	SUB	$256, R6
	CMPBNE	R6, $0, cmpblocks
	BR	bylen

cmptail:
	SUB	$1, R6, R7
	EXRL	$bytecmp_exrl_clc<>(SB), R7
	BLT	less
	BGT	greater

bylen:
	CMP	R4, R5
	BEQ	same
	BLT	less
	BR	greater

same:
	MOVD	$0, R2
	MOVD	R2, ret+48(FP)
	RET

less:
	MOVD	$-1, R2
	MOVD	R2, ret+48(FP)
	RET

greater:
	MOVD	$1, R2
	MOVD	R2, ret+48(FP)
	RET

// DO NOT CALL - target for EXRL, shared by Equal and Compare. NOFRAME is
// required: with a prologue, EXRL would execute the prologue instead of the
// CLC. The length byte assembled here is 0 (meaning one byte); EXRL ORs len-1
// into it at execution time.
TEXT bytecmp_exrl_clc<>(SB), NOSPLIT|NOFRAME, $0-0
	CLC	$1, 0(R2), 0(R3)
	MOVD	R0, 0(R0)		// guard: fault rather than fall through
	RET

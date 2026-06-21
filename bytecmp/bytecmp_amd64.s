#include "textflag.h"

// func Equal(a, b []byte) bool
//
// Argument layout on the stack (FP):
//   a []byte: a_base+0(FP),  a_len+8(FP),  a_cap+16(FP)
//   b []byte: b_base+24(FP), b_len+32(FP), b_cap+40(FP)
//   return bool: ret+48(FP)
//
// Different lengths are unequal immediately. Otherwise walk both buffers,
// comparing one byte at a time, and bail on the first mismatch.

TEXT ·Equal(SB), NOSPLIT, $0-49
	MOVQ	a_len+8(FP), CX
	MOVQ	b_len+32(FP), DX
	CMPQ	CX, DX
	JNE	notequal		// lengths differ -> not equal
	MOVQ	a_base+0(FP), SI
	MOVQ	b_base+24(FP), DI
	TESTQ	CX, CX
	JZ	equal			// both empty -> equal
eqloop:
	MOVBQZX	(SI), AX
	MOVBQZX	(DI), DX
	CMPQ	AX, DX
	JNE	notequal
	INCQ	SI
	INCQ	DI
	DECQ	CX
	JNZ	eqloop
equal:
	MOVB	$1, ret+48(FP)
	RET
notequal:
	MOVB	$0, ret+48(FP)
	RET

// func Compare(a, b []byte) int
//
// Argument layout on the stack (FP):
//   a []byte: a_base+0(FP),  a_len+8(FP),  a_cap+16(FP)
//   b []byte: b_base+24(FP), b_len+32(FP), b_cap+40(FP)
//   return int: ret+48(FP)
//
// Compare the first min(len(a), len(b)) bytes. At the first difference return
// -1 or +1. If the common prefix matches, the shorter slice sorts first.
// Bytes are zero-extended, so they are in 0..255 and the signed JLT/JGT give
// the same ordering as an unsigned byte compare.

TEXT ·Compare(SB), NOSPLIT, $0-56
	MOVQ	a_base+0(FP), SI
	MOVQ	a_len+8(FP), CX		// CX = len(a)
	MOVQ	b_base+24(FP), DI
	MOVQ	b_len+32(FP), DX		// DX = len(b)

	// BX = n = min(len(a), len(b))
	MOVQ	CX, BX
	CMPQ	DX, BX
	JGE	haven			// len(b) >= len(a) -> n = len(a)
	MOVQ	DX, BX			// else n = len(b)
haven:
	TESTQ	BX, BX
	JZ	bylen			// no bytes to compare -> decide by length
	XORQ	R8, R8			// R8 = i = 0
cmploop:
	MOVBQZX	(SI)(R8*1), AX		// a[i]
	MOVBQZX	(DI)(R8*1), R9		// b[i]
	CMPQ	AX, R9
	JLT	less
	JGT	greater
	INCQ	R8
	CMPQ	R8, BX
	JLT	cmploop
bylen:
	CMPQ	CX, DX			// common prefix equal; shorter sorts first
	JLT	less
	JGT	greater
	MOVQ	$0, ret+48(FP)		// equal
	RET
less:
	MOVQ	$-1, ret+48(FP)
	RET
greater:
	MOVQ	$1, ret+48(FP)
	RET

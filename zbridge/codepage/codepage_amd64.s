#include "textflag.h"

// func AtoE(dst, src []byte)
//
// dst layout on stack: dst_base+0(FP), dst_len+8(FP), dst_cap+16(FP)
// src layout on stack: src_base+24(FP), src_len+32(FP), src_cap+40(FP)
//
// Algorithm:
//   - load atoeTable address into BX
//   - load src base into SI, len into CX
//   - load dst base into DI
//   - loop: load byte from (SI), index into table at (BX), store to (DI)
//   - decrement CX, advance SI and DI, loop until CX == 0
//
// No frame needed (NOSPLIT, $0).

TEXT ·AtoE(SB), NOSPLIT, $0-48
	MOVQ	src_base+24(FP), SI
	MOVQ	src_len+32(FP), CX
	MOVQ	dst_base+0(FP), DI
	LEAQ	·atoeTable(SB), BX
	TESTQ	CX, CX
	JZ	done_atoe
loop_atoe:
	MOVBQZX	(SI), AX
	MOVB	(BX)(AX*1), DL
	MOVB	DL, (DI)
	INCQ	SI
	INCQ	DI
	DECQ	CX
	JNZ	loop_atoe
done_atoe:
	RET

// func EtoA(dst, src []byte)
// Identical structure, etoaTable instead.

TEXT ·EtoA(SB), NOSPLIT, $0-48
	MOVQ	src_base+24(FP), SI
	MOVQ	src_len+32(FP), CX
	MOVQ	dst_base+0(FP), DI
	LEAQ	·etoaTable(SB), BX
	TESTQ	CX, CX
	JZ	done_etoa
loop_etoa:
	MOVBQZX	(SI), AX
	MOVB	(BX)(AX*1), DL
	MOVB	DL, (DI)
	INCQ	SI
	INCQ	DI
	DECQ	CX
	JNZ	loop_etoa
done_etoa:
	RET

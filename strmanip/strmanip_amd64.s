#include "textflag.h"

// func StrLen(s string) int
//
// Go string header on the stack (FP):
//   s_base+0(FP)   data pointer (8 bytes)
//   s_len+8(FP)    length       (8 bytes)
// return:
//   ret+16(FP)     int          (8 bytes)
//
// A string is a TWO-word header, so the length sits at +8 and there is no
// capacity word. (A slice is three words: base, len, cap.) This is the whole
// point of the exercise: read the string header directly.

TEXT ·StrLen(SB), NOSPLIT, $0-24
	MOVQ	s_len+8(FP), AX
	MOVQ	AX, ret+16(FP)
	RET

// func WrapLengthPrefixed(dst []byte, msg string) int
//
// Argument layout on the stack (FP):
//   dst []byte (3 words): dst_base+0(FP), dst_len+8(FP), dst_cap+16(FP)
//   msg string (2 words): msg_base+24(FP), msg_len+32(FP)
//   return int:           ret+40(FP)
//
// Algorithm:
//   - write the message length as a 2-byte BIG-ENDIAN header into dst[0:2]
//     amd64 is little-endian, so the 16-bit length must be split by hand:
//     high byte first, then low byte
//   - copy len(msg) bytes from msg into dst[2:]
//   - return len(msg)+2
//
// No frame needed (NOSPLIT, $0); registers only, calls nothing.

TEXT ·WrapLengthPrefixed(SB), NOSPLIT, $0-48
	MOVQ	msg_base+24(FP), SI	// SI = source pointer
	MOVQ	msg_len+32(FP), CX	// CX = length
	MOVQ	dst_base+0(FP), DI	// DI = destination pointer

	// 2-byte big-endian length header: high byte, then low byte.
	MOVQ	CX, AX
	SHRQ	$8, AX			// AX = len >> 8
	MOVB	AX, (DI)		// dst[0] = high byte (AL)
	MOVB	CX, 1(DI)		// dst[1] = low byte  (CL)
	ADDQ	$2, DI			// advance past the header

	// copy CX bytes from (SI) to (DI)
	TESTQ	CX, CX
	JZ	done
loop:
	MOVB	(SI), AX
	MOVB	AX, (DI)
	INCQ	SI
	INCQ	DI
	DECQ	CX
	JNZ	loop
done:
	// return total bytes written = original length + 2
	MOVQ	msg_len+32(FP), AX
	ADDQ	$2, AX
	MOVQ	AX, ret+40(FP)
	RET

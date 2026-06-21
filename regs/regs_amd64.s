#include "textflag.h"

// func GetSP() uintptr
//
// Bare SP is the HARDWARE stack pointer register. (By contrast, "x+0(SP)",
// with a symbol, would be the pseudo stack pointer, a frame-relative virtual
// register. They are different things; this is the whole point of regs.)

TEXT ·GetSP(SB), NOSPLIT, $0-8
	MOVQ	SP, AX
	MOVQ	AX, ret+0(FP)
	RET

// func GetBP() uintptr
//
// BP is the hardware frame pointer on amd64. Go maintains it as a chain of
// call frames, so it is non-zero during normal execution.

TEXT ·GetBP(SB), NOSPLIT, $0-8
	MOVQ	BP, AX
	MOVQ	AX, ret+0(FP)
	RET

// func GetFramedSP() uintptr
//
// Same body as GetSP, but declared with a 256-byte frame. The function
// prologue lowers the hardware SP by the frame size before the read, so the
// value returned here is lower than GetSP's. Note that ret+0(FP) still
// addresses the result correctly: FP is the pseudo frame pointer and is
// independent of the declared frame size.

TEXT ·GetFramedSP(SB), NOSPLIT, $256-8
	MOVQ	SP, AX
	MOVQ	AX, ret+0(FP)
	RET

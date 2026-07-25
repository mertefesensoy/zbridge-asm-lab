// s390x implementation of GetSP, GetFramedSP, GetLinkRegister and GetGReg.
//
// This file replaces the former UNDEF stub. It does NOT implement GetBP: s390x
// has no frame-pointer register, and the package doc in regs.go explains why
// inventing one would have been the wrong answer.
//
// ---------------------------------------------------------------------------
// The register file, and why the exercise's central point survives the port
// ---------------------------------------------------------------------------
//
// amd64 has named registers with the stack pointer among them, so the hardware
// read in regs_amd64.s is "MOVQ SP, AX" — bare SP, no symbol, meaning the
// machine register.
//
// s390x has sixteen general registers R0-R15 and no register spelled SP. Go's
// own assignment (cmd/internal/obj/s390x/a.out.go) fixes the ones that matter:
//
//	R0	not usable as a base or index register: in an address, "base R0"
//		means "no base register", which is why no code here uses it
//	R10	REGTMP    scratch for the assembler and linker
//	R11	REGTMP2   second assembler scratch
//	R12	REGCTXT   closure context
//	R13	REGG      the goroutine pointer g
//	R14	the link register: return addresses live here, not on the stack
//	R15	REGSP     the stack pointer
//
// So the hardware stack-pointer read becomes "MOVD R15, R2".
//
// The pseudo-register distinction this exercise exists to teach is not softened
// by the port — it is sharpened. On s390x, bare SP in Go assembly is ONLY the
// pseudo stack pointer; the hardware register must be named R15. A line like
// "MOVD SP, R2" does not read the machine's stack pointer. The two things amd64
// spells with the same two letters have different names here, and code that
// confused them on amd64 will not even assemble the same way.
//
// ---------------------------------------------------------------------------
// Why R14 and R13 are the interesting registers for this project
// ---------------------------------------------------------------------------
//
// R14 is how control comes back. Standard MVS linkage returns through R14, and
// SVC 35 is issued in the middle of a Go assembly routine that must still be
// able to return to Go. GetLinkRegister makes that register observable from a
// Go test.
//
// R13 is the register that breaks hand-written Go assembly. It holds g. Clobber
// it and the runtime cannot find the current goroutine. GetGReg makes it
// observable so a test can assert what it should be: stable within a goroutine,
// different between goroutines.

#include "textflag.h"

// func GetSP() uintptr
//
// R15 is the HARDWARE stack pointer. (By contrast "x+0(SP)", with a symbol,
// is the pseudo stack pointer, a frame-relative virtual register. They are
// different things; this is the whole point of regs.)
//
// NOFRAME keeps the prologue from moving R15 before the read, so the value
// returned is this function's entry stack pointer.
TEXT ·GetSP(SB), NOSPLIT|NOFRAME, $0-8
	MOVD	R15, R2
	MOVD	R2, ret+0(FP)
	RET

// func GetFramedSP() uintptr
//
// Same body as GetSP, but declared with a 256-byte frame and deliberately
// WITHOUT NOFRAME, so the prologue lowers R15 by the frame size before the
// read. The value returned is therefore lower than GetSP's.
//
// ret+0(FP) still addresses the result correctly: FP is the pseudo frame
// pointer and is independent of the declared frame size.
TEXT ·GetFramedSP(SB), NOSPLIT, $256-8
	MOVD	R15, R2
	MOVD	R2, ret+0(FP)
	RET

// func GetLinkRegister() uintptr
//
// R14 holds the return address. NOFRAME matters here for a second reason: a
// frame would make the prologue spill R14 to the stack, and while it would
// still hold the return address at the point of the read, the read would no
// longer be demonstrating the leaf-call convention this function is here to
// show.
TEXT ·GetLinkRegister(SB), NOSPLIT|NOFRAME, $0-8
	MOVD	R14, R2
	MOVD	R2, ret+0(FP)
	RET

// func GetGReg() uintptr
//
// R13 holds g. Written as "g" rather than "R13" because that is the spelling
// Go's own runtime assembly uses and the assembler recognises; the register
// number is R13 either way.
TEXT ·GetGReg(SB), NOSPLIT|NOFRAME, $0-8
	MOVD	g, R2
	MOVD	R2, ret+0(FP)
	RET

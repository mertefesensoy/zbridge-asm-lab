// Package regs demonstrates reading hardware registers and observing stack
// frame layout in Plan 9 assembly. It is exercise 4 of zbridge-asm-lab; see the
// repo root README for the broader context and the roadmap toward WTO via
// SVC 35 from Go assembly on z/OS.
//
// The distinction this exercise makes concrete is between the two stack
// pointers in Go assembly:
//
//   - Bare SP (no symbol), as in "MOVQ SP, AX" on amd64 or "MOVD R15, R2" on
//     s390x, is the HARDWARE stack pointer register and holds a real address.
//   - A reference like "x-8(SP)", with a symbol and an offset, is the PSEUDO
//     stack pointer: a virtual register the assembler resolves to a slot in
//     the function's frame. Its value is unrelated to the hardware SP.
//
// Conflating the two is the most common Go-assembly mistake. GetSP returns the
// hardware value; the pseudo registers FP and SP are what the add, ebcdic, and
// strmanip exercises used to address their arguments.
//
// # Why this package's API is not identical on both architectures
//
// GetSP and GetFramedSP exist everywhere: every architecture Go supports has a
// hardware stack pointer, and every one of them grows the stack downward when a
// frame is declared.
//
// The frame POINTER is a different story. On amd64, Go maintains BP as a linked
// chain of call frames, so GetBP is meaningful and is declared in regs_amd64.go.
// On s390x there is no frame-pointer register at all — Go's own register
// assignment (cmd/internal/obj/s390x/a.out.go) gives R13 to the goroutine
// pointer g, R14 to the return address, and R15 to the stack pointer, and
// nothing to a frame pointer. Rather than return R13 from a function called
// GetBP and quietly redefine what "BP" means, the s390x port declares what it
// actually has: GetLinkRegister and GetGReg, in regs_s390x.go.
//
// That split is the honest answer to a real architectural difference, and it
// makes this exercise MORE relevant to the endgame rather than less: R14 is the
// register the SVC 35 call path returns through, and R13 is the register a Go
// assembly routine must not clobber if it wants to return to a working runtime.
package regs

// GetSP returns the value of the hardware stack pointer register, read inside
// GetSP's own frame. On amd64 that register is SP; on s390x it is R15.
func GetSP() uintptr

// GetFramedSP is identical to GetSP except that it is declared with a 256-byte
// stack frame. Because that frame is allocated below the return address, the
// hardware SP read inside it is lower than GetSP's by at least the frame size.
// It exists to make the effect of a declared frame size on the stack pointer
// directly observable, and the effect is architecture-independent: both amd64
// and s390x grow the stack toward lower addresses.
func GetFramedSP() uintptr

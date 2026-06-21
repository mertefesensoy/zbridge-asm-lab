// Package regs demonstrates reading hardware registers and observing stack
// frame layout in Plan 9 assembly on amd64. It is exercise 4 of
// zbridge-asm-lab; see the repo root README for the broader context and the
// roadmap toward WTO via SVC 35 from Go assembly on z/OS.
//
// The distinction this exercise makes concrete is between the two stack
// pointers in Go assembly:
//
//   - Bare SP (no symbol), as in "MOVQ SP, AX", is the HARDWARE stack pointer
//     register and holds a real address.
//   - A reference like "x-8(SP)", with a symbol and an offset, is the PSEUDO
//     stack pointer: a virtual register the assembler resolves to a slot in
//     the function's frame. Its value is unrelated to the hardware SP.
//
// Conflating the two is the most common Go-assembly mistake. GetSP returns the
// hardware value; the pseudo registers FP and SP are what the add, ebcdic, and
// strmanip exercises used to address their arguments.
package regs

// GetSP returns the value of the hardware stack pointer register, read inside
// GetSP's own frame.
func GetSP() uintptr

// GetBP returns the value of the hardware base pointer register (BP), which Go
// maintains as the frame pointer on amd64. It links call frames together.
func GetBP() uintptr

// GetFramedSP is identical to GetSP except that it is declared with a 256-byte
// stack frame. Because that frame is allocated below the return address, the
// hardware SP read inside it is lower than GetSP's by at least the frame size.
// It exists to make the effect of a declared frame size on the stack pointer
// directly observable.
func GetFramedSP() uintptr

// Package svc holds the raw supervisor-call primitives.
//
// It is internal because a supervisor call with the wrong parameter list is a
// good way to disturb a system that other people are using. Callers go through
// the service packages, which validate first.
//
// # SVC 35 must be hand-encoded
//
// Go's s390x assembler has no SVC mnemonic. Verified against
// cmd/internal/obj/s390x/anames.go (go1.26.3): 729 mnemonics, and SVC, TR, TRT
// and EX are all absent. SYSCALL exists and assembles to SVC 0, which is Linux's
// trap — not what an MVS service needs.
//
// So SVC 35 is emitted as literal bytes:
//
//	BYTE $0x0A; BYTE $0x23        // SVC 35 (0x23 = 35)
//
// The technique is attested inside the Go distribution itself:
// golang.org/x/sys/unix/asm_zos_s390x.s encodes SVC 08 as
// "BYTE $0x0A; BYTE $0x08" and SVC 09 as "BYTE $0x0A; BYTE $0x09".
//
// This project has already exercised and verified the same technique for TR: the
// hand-written bytes were disassembled with GNU binutils and confirmed to decode
// as the intended instruction, then executed under QEMU with a differential test
// against a reference implementation. The identical verification is MANDATORY
// for SVC 35 before any claim rests on it. See
// docs/evidence/E-L-s390x-port-qemu-2026-07-25.md.
//
// # MVS linkage, which is not the Go ABI
//
// Standard MVS linkage on entry to a service expects:
//
//	R1   address of the parameter list
//	R13  address of a 72-byte save area the callee may use
//	R14  return address
//	R15  entry point address, and the return code on the way back
//
// Two of those collide with Go's s390x conventions. Go reserves R13 for the
// goroutine pointer g and R15 for the stack pointer
// (cmd/internal/obj/s390x/a.out.go: REGG = REG_R13, REGSP = REG_R15). Any Go
// assembly routine that establishes MVS linkage must therefore save and restore
// both, or the runtime loses the current goroutine and the stack.
//
// This collision is the single most dangerous detail in the whole project, and
// it is why the regs exercise makes R13 and R14 directly observable from a Go
// test.
package svc

// Supported reports whether this build includes the supervisor-call primitives.
// It is false on every platform except z/OS on s390x.
const Supported = supported

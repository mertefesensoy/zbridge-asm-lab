//go:build linux && s390x

// s390x implementation of Getpid and Write.
//
// This file replaces the former UNDEF stub. It is the exercise that maps most
// directly onto the endgame, so the differences from amd64 are worth stating
// precisely rather than waving at.
//
// ---------------------------------------------------------------------------
// The Linux s390x system-call ABI
// ---------------------------------------------------------------------------
//
//	R1        system-call number
//	R2..R7    arguments 1 through 6
//	SYSCALL   the trap
//	R2        return value, and the argument register it overwrites
//
// Two things differ structurally from amd64, beyond the register names:
//
//  1. The number goes in R1, which is NOT the first argument register. On amd64
//     the number goes in AX and the first argument in DI, so the number has its
//     own register there too — but on amd64 AX is also the return register. On
//     s390x the return comes back in R2, the FIRST ARGUMENT register. Code that
//     needs its first argument after the call must save it.
//
//  2. Go's SYSCALL mnemonic assembles to SVC 0 on this architecture. The trap
//     instruction of z/Architecture Linux is the same instruction family the
//     z/OS endgame uses; only the operand differs. Linux takes the service
//     selector from R1 and issues SVC 0; MVS encodes the selector in the
//     instruction, so WTO is SVC 35. Seeing SVC appear here, on Linux, is the
//     point of this exercise.
//
// ---------------------------------------------------------------------------
// The errno convention
// ---------------------------------------------------------------------------
//
// A returned value in the range -4095..-1 is -errno. The check below is the
// one Go's own syscall package uses for this target
// (syscall/asm_linux_s390x.s): materialise 0xfffffffffffff001 in a register and
// take an UNSIGNED less-than branch.
//
// Note a one-value difference from the amd64 file in this exercise. amd64 uses
// JLS (unsigned <=), so it treats exactly -4095 as success; s390x uses CMPUBLT
// (unsigned <), so it treats -4095 as an error. Both mirror what the Go
// standard library does on each architecture. No Linux system call returns
// errno 4095, so the two are indistinguishable in practice, but the asymmetry
// is real and is recorded rather than smoothed over.
//
// Registers used, none of which may be R0 (not usable as a base register),
// R10/R11 (REGTMP), R12 (REGCTXT), R13 (g), R14 (link) or R15 (stack pointer).

#include "textflag.h"

#define SYS_write  4
#define SYS_getpid 20

// func Getpid() (pid uintptr, errno uintptr)
//
// Argument layout (FP): pid+0, errno+8.
// getpid takes no parameters, so this is the minimal SVC example.
TEXT ·Getpid(SB), NOSPLIT|NOFRAME, $0-16
	MOVD	$SYS_getpid, R1
	SYSCALL
	MOVD	$0xfffffffffffff001, R8
	CMPUBLT	R2, R8, getpid_ok
	NEG	R2, R2
	MOVD	$-1, pid+0(FP)
	MOVD	R2, errno+8(FP)
	RET

getpid_ok:
	MOVD	R2, pid+0(FP)
	MOVD	$0, errno+8(FP)
	RET

// func Write(fd uintptr, p []byte) (n uintptr, errno uintptr)
//
// Argument layout (FP):
//	fd uintptr: fd+0
//	p []byte:   p_base+8, p_len+16, p_cap+24
//	returns:    n+32, errno+40
//
// Parameter placement:
//	R1 = SYS_write
//	R2 = fd
//	R3 = &p[0]   (the slice data pointer)
//	R4 = len(p)
//
// R2 carries fd in and the result out, which is the asymmetry noted in the
// header. Nothing here needs fd afterwards, so nothing is saved.
TEXT ·Write(SB), NOSPLIT|NOFRAME, $0-48
	MOVD	$SYS_write, R1
	MOVD	fd+0(FP), R2
	MOVD	p_base+8(FP), R3
	MOVD	p_len+16(FP), R4
	SYSCALL
	MOVD	$0xfffffffffffff001, R8
	CMPUBLT	R2, R8, write_ok
	NEG	R2, R2
	MOVD	$-1, n+32(FP)
	MOVD	R2, errno+40(FP)
	RET

write_ok:
	MOVD	R2, n+32(FP)
	MOVD	$0, errno+40(FP)
	RET

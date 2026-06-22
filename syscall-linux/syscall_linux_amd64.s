//go:build linux && amd64

#include "textflag.h"

#define SYS_write 1
#define SYS_getpid 39
#define ERRNO_MAX 0xfffffffffffff001

// func Getpid() (pid uintptr, errno uintptr)
//
// Linux amd64 syscall ABI:
//   AX  syscall number
//   DI  arg1
//   SI  arg2
//   DX  arg3
//   R10 arg4 (not CX; SYSCALL clobbers CX and R11)
//   R8  arg5
//   R9  arg6
//
// Return:
//   AX >= 0             success value
//   AX in -4095..-1     -errno
//
// getpid has no parameters, so this is the minimal SYSCALL example.

TEXT ·Getpid(SB), NOSPLIT, $0-16
	MOVQ	$SYS_getpid, AX
	SYSCALL
	CMPQ	AX, $ERRNO_MAX
	JLS	getpid_ok
	NEGQ	AX
	MOVQ	$-1, pid+0(FP)
	MOVQ	AX, errno+8(FP)
	RET
getpid_ok:
	MOVQ	AX, pid+0(FP)
	MOVQ	$0, errno+8(FP)
	RET

// func Write(fd uintptr, p []byte) (n uintptr, errno uintptr)
//
// Argument layout on the stack (FP):
//   fd uintptr: fd+0(FP)
//   p []byte:   p_base+8(FP), p_len+16(FP), p_cap+24(FP)
//   returns:    n+32(FP), errno+40(FP)
//
// write(2) demonstrates parameter placement:
//   AX = SYS_write
//   DI = fd
//   SI = &p[0] (the slice data pointer)
//   DX = len(p)
//
// Then SYSCALL transfers control to the kernel. The negative-return check is
// the same shape Go's syscall package uses: values in -4095..-1 are errno.

TEXT ·Write(SB), NOSPLIT, $0-48
	MOVQ	$SYS_write, AX
	MOVQ	fd+0(FP), DI
	MOVQ	p_base+8(FP), SI
	MOVQ	p_len+16(FP), DX
	SYSCALL
	CMPQ	AX, $ERRNO_MAX
	JLS	write_ok
	NEGQ	AX
	MOVQ	$-1, n+32(FP)
	MOVQ	AX, errno+40(FP)
	RET
write_ok:
	MOVQ	AX, n+32(FP)
	MOVQ	$0, errno+40(FP)
	RET

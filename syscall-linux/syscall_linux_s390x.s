//go:build linux && s390x

// STUB - not implemented. This file documents the s390x and z/OS port path
// for the syscall-linux exercise.
//
// On Linux, s390x has its own system-call ABI and trap instruction sequence,
// but the conceptual shape is the same as amd64: load a service number and
// parameters into the ABI-defined registers, trap into the operating system,
// and evaluate the returned register value.
//
// The z/OS target is not Linux syscalls; it is MVS services. WTO uses SVC 35:
// parameters are placed in registers such as R0 and R1, the SVC instruction
// transfers control to the supervisor, and the return code comes back in a
// register. This exercise proves the generic "register setup + trap + return
// evaluation" primitive before the project swaps Linux SYSCALL for z/OS SVC.
//
// Implementation is deferred until Phase 1b of the roadmap, when LinuxONE
// Community Cloud access enables real s390x testing.
//
// The stub TEXT directives below cause a "not implemented" build failure if
// this file is accidentally targeted; this is intentional.

#include "textflag.h"

TEXT ·Getpid(SB), NOSPLIT, $0-16
	UNDEF

TEXT ·Write(SB), NOSPLIT, $0-48
	UNDEF

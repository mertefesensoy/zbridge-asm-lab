// Package syscalllinux demonstrates issuing Linux system calls directly from Go
// assembly. It is a rehearsal for the z/OS SVC path: load the service number
// and parameters into the ABI-defined registers, execute the trap instruction,
// and interpret the returned register value.
//
// Two targets are implemented, and the pair is more instructive than either
// alone:
//
//	linux/amd64  number in AX, args in DI/SI/DX/R10/R8/R9, trap is SYSCALL,
//	             result in AX. write is 1, getpid is 39.
//	linux/s390x  number in R1, args in R2..R7, trap is SYSCALL which assembles
//	             to SVC 0, result in R2. write is 4, getpid is 20.
//
// The numbers differ, the registers differ, and on s390x the result comes back
// in the same register that carried the first argument. What does NOT differ is
// the shape: a service selector, parameters in fixed registers, one trap
// instruction, and a return value that must be classified as result or error.
// That shape is what the z/OS endgame reuses, with SVC 35 in place of SYSCALL.
package syscalllinux

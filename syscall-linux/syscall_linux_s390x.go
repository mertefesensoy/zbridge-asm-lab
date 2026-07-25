//go:build linux && s390x

package syscalllinux

// The system-call NUMBERS differ from amd64, and that is not a detail — it is
// the first concrete demonstration in this repo that a system-call interface is
// a per-architecture contract, not a per-operating-system one.
//
// Linux on s390x uses the numbering inherited from the 31-bit s390 port, which
// is close to the original i386 table and nothing like the amd64 table:
//
//	           amd64   s390x
//	write        1        4
//	getpid      39       20
//
// Verified against Go's own runtime for this target,
// runtime/sys_linux_s390x.s (go1.26.3), which defines SYS_write 4 and
// SYS_getpid 20.
//
// The same lesson carries to the endgame with a twist. On z/OS the service
// number is not a table index in a Go constant at all: it is encoded in the SVC
// instruction itself, so WTO is literally the instruction SVC 35. Where Linux
// puts the number in a register and traps, MVS puts it in the opcode.
const (
	// SYSWrite is the Linux s390x write(2) system-call number.
	SYSWrite uintptr = 4

	// SYSGetpid is the Linux s390x getpid(2) system-call number.
	SYSGetpid uintptr = 20
)

// Getpid issues the Linux getpid system call and returns the process ID plus
// an errno value. getpid normally cannot fail, but returning errno keeps the
// return-value handling visible for the exercise.
//
//go:noescape
func Getpid() (pid uintptr, errno uintptr)

// Write issues the Linux write system call for fd and p. It returns the byte
// count and errno. On error the byte count is ^uintptr(0), matching the raw
// Linux syscall convention after converting a negative return into errno.
//
//go:noescape
func Write(fd uintptr, p []byte) (n uintptr, errno uintptr)

//go:build linux && amd64

package syscalllinux

const (
	// SYSWrite is the Linux amd64 write(2) system-call number.
	SYSWrite uintptr = 1

	// SYSGetpid is the Linux amd64 getpid(2) system-call number.
	SYSGetpid uintptr = 39
)

// Getpid issues the Linux getpid system call and returns the process ID plus
// an errno value. getpid normally cannot fail, but returning errno keeps the
// return-value handling visible for the exercise.
//
//go:noescape
func Getpid() (pid uintptr, errno uintptr)

// Write issues the Linux write system call for fd and p. It returns the byte
// count and errno. On error the byte count is ^uintptr(0), matching the raw
// Linux syscall convention after converting a negative RAX return into errno.
//
//go:noescape
func Write(fd uintptr, p []byte) (n uintptr, errno uintptr)

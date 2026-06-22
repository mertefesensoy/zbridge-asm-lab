//go:build !linux || !amd64

package syscalllinux

import "testing"

func TestSyscallLinuxRequiresLinuxAMD64(t *testing.T) {
	t.Skip("syscall-linux executes the Linux amd64 SYSCALL instruction; run on linux/amd64 for runtime tests")
}

//go:build !linux || (!amd64 && !s390x)

package syscalllinux

import "testing"

func TestSyscallLinuxRequiresSupportedTarget(t *testing.T) {
	t.Skip("syscall-linux executes the Linux SYSCALL trap; run on linux/amd64 or linux/s390x for runtime tests")
}

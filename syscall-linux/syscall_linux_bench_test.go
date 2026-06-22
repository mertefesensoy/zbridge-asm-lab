//go:build linux && amd64

package syscalllinux

import "testing"

var pidSink uintptr

// BenchmarkGetpid measures the cost of one tiny Go-to-assembly-to-kernel
// round trip. It is not meant to beat the runtime or standard library; it is
// a visible baseline for the trap path this exercise is about.
func BenchmarkGetpid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pidSink, _ = Getpid()
	}
}

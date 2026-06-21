package regs

import "testing"

// sink prevents the benchmarked call from being optimized away.
var sink uintptr

// BenchmarkGetSP measures the ABI0 call overhead of a minimal assembly
// function. It is here for template consistency with the other exercises;
// reading a register needs no optimizing, but the number is a useful baseline
// for what a single Go-to-assembly call costs.
func BenchmarkGetSP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = GetSP()
	}
}

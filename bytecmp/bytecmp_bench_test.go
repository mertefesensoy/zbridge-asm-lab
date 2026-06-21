package bytecmp

import (
	"bytes"
	"testing"
)

func benchPair(n int) (a, b []byte) {
	a = make([]byte, n)
	b = make([]byte, n)
	for i := range a {
		a[i] = byte('A' + i%26)
		b[i] = a[i]
	}
	b[n-1] ^= 1 // differ in the last byte: worst case, full scan
	return a, b
}

var benchA, benchB = benchPair(256)

func BenchmarkEqualAsm(b *testing.B) {
	b.SetBytes(int64(len(benchA)))
	for i := 0; i < b.N; i++ {
		Equal(benchA, benchB)
	}
}

func BenchmarkEqualStd(b *testing.B) {
	b.SetBytes(int64(len(benchA)))
	for i := 0; i < b.N; i++ {
		bytes.Equal(benchA, benchB)
	}
}

func BenchmarkCompareAsm(b *testing.B) {
	b.SetBytes(int64(len(benchA)))
	for i := 0; i < b.N; i++ {
		Compare(benchA, benchB)
	}
}

func BenchmarkCompareStd(b *testing.B) {
	b.SetBytes(int64(len(benchA)))
	for i := 0; i < b.N; i++ {
		bytes.Compare(benchA, benchB)
	}
}

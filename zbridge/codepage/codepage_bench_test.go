package codepage

import "testing"

// Pure-Go reference implementation for benchmark comparison.
func atoeRef(dst, src []byte) {
	for i, b := range src {
		dst[i] = atoeTable[b]
	}
}

var benchSrc = func() []byte {
	b := make([]byte, 4096)
	for i := range b {
		b[i] = byte('A' + i%26)
	}
	return b
}()

func BenchmarkAtoEAsm(b *testing.B) {
	dst := make([]byte, len(benchSrc))
	b.SetBytes(int64(len(benchSrc)))
	for i := 0; i < b.N; i++ {
		AtoE(dst, benchSrc)
	}
}

func BenchmarkAtoEGo(b *testing.B) {
	dst := make([]byte, len(benchSrc))
	b.SetBytes(int64(len(benchSrc)))
	for i := 0; i < b.N; i++ {
		atoeRef(dst, benchSrc)
	}
}

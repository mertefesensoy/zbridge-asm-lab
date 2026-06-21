package strmanip

import "testing"

var benchMsg = string(func() []byte {
	b := make([]byte, 120) // a realistic single-line operator-message length
	for i := range b {
		b[i] = byte('A' + i%26)
	}
	return b
}())

func BenchmarkWrapAsm(b *testing.B) {
	dst := make([]byte, len(benchMsg)+2)
	b.SetBytes(int64(len(benchMsg) + 2))
	for i := 0; i < b.N; i++ {
		WrapLengthPrefixed(dst, benchMsg)
	}
}

func BenchmarkWrapGo(b *testing.B) {
	dst := make([]byte, len(benchMsg)+2)
	b.SetBytes(int64(len(benchMsg) + 2))
	for i := 0; i < b.N; i++ {
		goReference(dst, benchMsg)
	}
}

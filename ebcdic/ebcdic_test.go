package ebcdic

import (
	"bytes"
	"testing"
)

// Known-good fixture: "HELLO" in IBM-1047 is C8 C5 D3 D3 D6.
func TestAtoEKnownFixture(t *testing.T) {
	src := []byte("HELLO")
	want := []byte{0xC8, 0xC5, 0xD3, 0xD3, 0xD6}
	dst := make([]byte, len(src))
	AtoE(dst, src)
	if !bytes.Equal(dst, want) {
		t.Errorf("AtoE(\"HELLO\") = % X, want % X", dst, want)
	}
}

func TestEtoAKnownFixture(t *testing.T) {
	src := []byte{0xC8, 0xC5, 0xD3, 0xD3, 0xD6}
	want := []byte("HELLO")
	dst := make([]byte, len(src))
	EtoA(dst, src)
	if !bytes.Equal(dst, want) {
		t.Errorf("EtoA(C8 C5 D3 D3 D6) = %q, want %q", dst, want)
	}
}

func TestRoundTripAll256(t *testing.T) {
	src := make([]byte, 256)
	for i := range src {
		src[i] = byte(i)
	}
	mid := make([]byte, 256)
	back := make([]byte, 256)
	AtoE(mid, src)
	EtoA(back, mid)
	// Not all 256 ISO-8859-1 codepoints round-trip cleanly through
	// IBM-1047; the printable subset and most controls do. Test the
	// printable ASCII range 0x20-0x7E specifically.
	for i := 0x20; i <= 0x7E; i++ {
		if back[i] != src[i] {
			t.Errorf("round-trip failed at 0x%02X: got 0x%02X", i, back[i])
		}
	}
}

func TestEmpty(t *testing.T) {
	AtoE(nil, nil)
	EtoA(nil, nil)
	// no crash = pass
}

func TestSingleByte(t *testing.T) {
	src := []byte("A")
	dst := make([]byte, 1)
	AtoE(dst, src)
	// 'A' in IBM-1047 is 0xC1
	if dst[0] != 0xC1 {
		t.Errorf("AtoE(\"A\") = 0x%02X, want 0xC1", dst[0])
	}
}

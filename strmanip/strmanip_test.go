package strmanip

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestStrLen(t *testing.T) {
	cases := map[string]int{
		"":          0,
		"A":         1,
		"HELLO":     5,
		"operator!": 9,
	}
	for s, want := range cases {
		if got := StrLen(s); got != want {
			t.Errorf("StrLen(%q) = %d, want %d", s, got, want)
		}
	}
}

// goReference is the pure-Go equivalent of WrapLengthPrefixed, used as the
// oracle for the fuzz-style and benchmark comparisons.
func goReference(dst []byte, msg string) int {
	binary.BigEndian.PutUint16(dst[0:2], uint16(len(msg)))
	copy(dst[2:], msg)
	return len(msg) + 2
}

func TestWrapKnownFixture(t *testing.T) {
	msg := "HELLO"
	dst := make([]byte, len(msg)+2)
	n := WrapLengthPrefixed(dst, msg)

	want := []byte{0x00, 0x05, 'H', 'E', 'L', 'L', 'O'}
	if n != len(want) {
		t.Fatalf("returned n = %d, want %d", n, len(want))
	}
	if !bytes.Equal(dst, want) {
		t.Errorf("WrapLengthPrefixed(%q) = % X, want % X", msg, dst, want)
	}
}

func TestWrapEmpty(t *testing.T) {
	dst := make([]byte, 2)
	n := WrapLengthPrefixed(dst, "")
	if n != 2 {
		t.Errorf("returned n = %d, want 2", n)
	}
	if dst[0] != 0x00 || dst[1] != 0x00 {
		t.Errorf("empty header = % X, want 00 00", dst)
	}
}

// TestWrapHighByte exercises the big-endian split: a length above 255 puts a
// non-zero value in the HIGH byte, which is exactly what amd64 has to build by
// hand and what s390x will do with a single STH.
func TestWrapHighByte(t *testing.T) {
	msg := string(bytes.Repeat([]byte{'x'}, 300)) // 300 = 0x012C
	dst := make([]byte, len(msg)+2)
	n := WrapLengthPrefixed(dst, msg)

	if n != 302 {
		t.Fatalf("returned n = %d, want 302", n)
	}
	if dst[0] != 0x01 || dst[1] != 0x2C {
		t.Errorf("length header = % X, want 01 2C", dst[0:2])
	}
	if !bytes.Equal(dst[2:], []byte(msg)) {
		t.Errorf("payload mismatch")
	}
}

// TestWrapMatchesReference checks the assembly against the pure-Go oracle
// across many lengths and byte values.
func TestWrapMatchesReference(t *testing.T) {
	for n := 0; n <= 512; n++ {
		buf := make([]byte, n)
		for i := range buf {
			buf[i] = byte((i * 7) & 0xFF)
		}
		msg := string(buf)

		gotDst := make([]byte, n+2)
		wantDst := make([]byte, n+2)
		gotN := WrapLengthPrefixed(gotDst, msg)
		wantN := goReference(wantDst, msg)

		if gotN != wantN {
			t.Fatalf("len %d: returned %d, want %d", n, gotN, wantN)
		}
		if !bytes.Equal(gotDst, wantDst) {
			t.Fatalf("len %d: buffer mismatch\n got % X\nwant % X", n, gotDst, wantDst)
		}
	}
}

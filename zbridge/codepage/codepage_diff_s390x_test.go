//go:build s390x

package codepage

import (
	"bytes"
	"testing"
)

// diffLengths covers the boundaries that matter to the s390x implementation:
// the empty case, the single-byte EXRL case, either side of the 256-byte block
// stride, and a multi-block buffer with a partial tail.
//
// 255/256/257 are the interesting ones. At 255 the whole operation is one
// EXRL-driven MVC plus one EXRL-driven TR. At 256 it is exactly one directly
// assembled block and no tail. At 257 it is one block plus a one-byte tail,
// which is where an off-by-one in the pointer advance or the length decrement
// would show up.
var diffLengths = []int{0, 1, 2, 15, 16, 255, 256, 257, 511, 512, 513, 4096}

// fill produces a buffer whose byte values cycle through all 256 possibilities,
// so every entry of the translation table is exercised on any buffer of length
// 256 or more, and the offset makes successive lengths use different data.
func fill(n, offset int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte((i + offset) % 256)
	}
	return b
}

// TestAtoETRMatchesLoop is the differential test for the hand-encoded TR path.
//
// If these two disagree, the six bytes DC FF 20 00 30 00 (or the EXRL-patched
// variant) do not mean what the header of codepage_s390x.s says they mean, and
// the TR path must not be trusted. Failure here falsifies H002 claim C2 for
// this module.
func TestAtoETRMatchesLoop(t *testing.T) {
	for _, n := range diffLengths {
		src := fill(n, 0)
		viaTR := make([]byte, n)
		viaLoop := make([]byte, n)

		AtoE(viaTR, src)
		AtoELoop(viaLoop, src)

		if !bytes.Equal(viaTR, viaLoop) {
			t.Errorf("AtoE vs AtoELoop disagree at len=%d:\n  TR   = % X\n  loop = % X",
				n, viaTR, viaLoop)
		}
	}
}

func TestEtoATRMatchesLoop(t *testing.T) {
	for _, n := range diffLengths {
		src := fill(n, 7)
		viaTR := make([]byte, n)
		viaLoop := make([]byte, n)

		EtoA(viaTR, src)
		EtoALoop(viaLoop, src)

		if !bytes.Equal(viaTR, viaLoop) {
			t.Errorf("EtoA vs EtoALoop disagree at len=%d:\n  TR   = % X\n  loop = % X",
				n, viaTR, viaLoop)
		}
	}
}

// TestTRCoversEveryTableEntry checks the TR path against the Go table directly,
// for all 256 input values at once. The loop path shares the table lookup with
// the Go reference, so agreeing with the loop is a weaker statement than
// agreeing with the table; this test makes the stronger one.
func TestTRCoversEveryTableEntry(t *testing.T) {
	src := make([]byte, 256)
	for i := range src {
		src[i] = byte(i)
	}

	got := make([]byte, 256)
	AtoE(got, src)
	for i := 0; i < 256; i++ {
		if got[i] != atoeTable[i] {
			t.Errorf("AtoE via TR: input 0x%02X -> 0x%02X, table says 0x%02X",
				i, got[i], atoeTable[i])
		}
	}

	EtoA(got, src)
	for i := 0; i < 256; i++ {
		if got[i] != etoaTable[i] {
			t.Errorf("EtoA via TR: input 0x%02X -> 0x%02X, table says 0x%02X",
				i, got[i], etoaTable[i])
		}
	}
}

// TestTRDoesNotOverrun guards the block/tail arithmetic. The destination is
// allocated with padding on both sides; the padding must be untouched.
func TestTRDoesNotOverrun(t *testing.T) {
	const pad = 32

	for _, n := range diffLengths {
		buf := make([]byte, pad+n+pad)
		for i := range buf {
			buf[i] = 0xAA
		}
		src := fill(n, 3)

		AtoE(buf[pad:pad+n], src)

		for i := 0; i < pad; i++ {
			if buf[i] != 0xAA {
				t.Fatalf("len=%d: wrote %d bytes before dst[0]", n, pad-i)
			}
		}
		for i := pad + n; i < len(buf); i++ {
			if buf[i] != 0xAA {
				t.Fatalf("len=%d: wrote past dst[%d] at offset %d", n, n, i-(pad+n))
			}
		}
	}
}

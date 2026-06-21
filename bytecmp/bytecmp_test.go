package bytecmp

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestEqualFixtures(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"", "", true},
		{"", "a", false},
		{"a", "", false},
		{"abc", "abc", true},
		{"abc", "abd", false},
		{"abc", "ab", false},
		{"HELLO", "HELLO", true},
		{"HELLO", "hello", false},
	}
	for _, c := range cases {
		got := Equal([]byte(c.a), []byte(c.b))
		if got != c.want {
			t.Errorf("Equal(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestCompareFixtures(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"", "a", -1},
		{"a", "", 1},
		{"abc", "abc", 0},
		{"abc", "abd", -1},
		{"abd", "abc", 1},
		{"ab", "abc", -1},
		{"abc", "ab", 1},
		{"\x00", "\xff", -1}, // unsigned byte ordering
		{"\xff", "\x00", 1},
	}
	for _, c := range cases {
		got := Compare([]byte(c.a), []byte(c.b))
		if got != c.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

// TestEqualAgainstStdlib checks Equal against bytes.Equal across many shapes.
func TestEqualAgainstStdlib(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for iter := 0; iter < 5000; iter++ {
		a := randBytes(r, r.Intn(40))
		b := randBytes(r, r.Intn(40))
		if r.Intn(3) == 0 { // sometimes force equal inputs
			b = append([]byte(nil), a...)
		}
		if got, want := Equal(a, b), bytes.Equal(a, b); got != want {
			t.Fatalf("Equal(% X, % X) = %v, want %v", a, b, got, want)
		}
	}
}

// TestCompareAgainstStdlib checks Compare against bytes.Compare across many
// shapes, including shared-prefix pairs that exercise the length tiebreak.
func TestCompareAgainstStdlib(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for iter := 0; iter < 5000; iter++ {
		a := randBytes(r, r.Intn(40))
		b := randBytes(r, r.Intn(40))
		switch r.Intn(4) {
		case 0:
			b = append([]byte(nil), a...) // equal
		case 1:
			if len(a) > 0 {
				b = append([]byte(nil), a[:r.Intn(len(a))]...) // prefix of a
			}
		}
		if got, want := Compare(a, b), bytes.Compare(a, b); got != want {
			t.Fatalf("Compare(% X, % X) = %d, want %d", a, b, got, want)
		}
	}
}

func randBytes(r *rand.Rand, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.Intn(256))
	}
	return b
}

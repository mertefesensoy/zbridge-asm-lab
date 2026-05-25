package add

import "testing"

func TestAdd(t *testing.T) {
	cases := []struct {
		a, b, want int64
	}{
		{2, 3, 5},
		{0, 0, 0},
		{-1, 1, 0},
		{1 << 40, 1 << 40, 1 << 41},
	}
	for _, c := range cases {
		if got := add(c.a, c.b); got != c.want {
			t.Errorf("add(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	var sum int64
	for i := 0; i < b.N; i++ {
		sum = add(int64(i), int64(i+1))
	}
	_ = sum
}

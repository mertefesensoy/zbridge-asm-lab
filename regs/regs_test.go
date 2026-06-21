package regs

import "testing"

func TestGetSPNonZero(t *testing.T) {
	if GetSP() == 0 {
		t.Error("GetSP() returned 0; the hardware stack pointer is never 0 during execution")
	}
}

func TestGetBPNonZero(t *testing.T) {
	if GetBP() == 0 {
		t.Error("GetBP() returned 0; the frame pointer should be set during execution")
	}
}

// TestFrameSizeShiftsSP shows that a declared stack frame lowers the hardware
// stack pointer by at least the frame size. Both functions are called from the
// same test frame, so the only difference between them is GetFramedSP's
// 256-byte frame.
func TestFrameSizeShiftsSP(t *testing.T) {
	shallow := GetSP()
	framed := GetFramedSP()

	if framed >= shallow {
		t.Fatalf("expected framed SP (%#x) below shallow SP (%#x); the stack grows toward lower addresses", framed, shallow)
	}
	if shallow-framed < 256 {
		t.Errorf("frame shift = %d bytes, want >= 256 (the declared frame size)", shallow-framed)
	}
}

//go:noinline
func spAtDepth(d int) uintptr {
	if d == 0 {
		return GetSP()
	}
	return spAtDepth(d - 1)
}

// TestStackGrowsDownward shows that deeper call nesting yields a lower stack
// pointer: each frame is placed below the previous one on amd64.
func TestStackGrowsDownward(t *testing.T) {
	shallow := spAtDepth(0)
	deep := spAtDepth(8)

	if deep >= shallow {
		t.Fatalf("expected deep SP (%#x) below shallow SP (%#x)", deep, shallow)
	}
}

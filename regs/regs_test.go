package regs

import "testing"

func TestGetSPNonZero(t *testing.T) {
	if GetSP() == 0 {
		t.Error("GetSP() returned 0; the hardware stack pointer is never 0 during execution")
	}
}

// The frame-pointer test lives in regs_amd64_test.go: GetBP is amd64-only,
// because s390x has no frame-pointer register. The s390x register tests are in
// regs_s390x_test.go. Everything in this file is architecture-neutral.

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

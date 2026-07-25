package regs

import "testing"

// TestGetBPNonZero is amd64-only, by filename constraint. BP is Go's frame
// pointer on amd64; s390x has no frame-pointer register and tests its own
// registers in regs_s390x_test.go.
func TestGetBPNonZero(t *testing.T) {
	if GetBP() == 0 {
		t.Error("GetBP() returned 0; the frame pointer should be set during execution")
	}
}

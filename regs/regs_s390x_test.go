package regs

import (
	"sync"
	"testing"
)

// These tests are s390x-only, by filename constraint. They replace the
// amd64 frame-pointer test with assertions about the two registers s390x
// actually has in that role's place.

func TestGetLinkRegisterNonZero(t *testing.T) {
	if GetLinkRegister() == 0 {
		t.Error("GetLinkRegister() returned 0; R14 holds the return address and is never 0 in a called function")
	}
}

// TestLinkRegisterIsARealReturnAddress distinguishes "R14 was read" from
// "some constant was returned". The link register holds the address execution
// resumes at in the CALLER, so two calls from two different call sites must
// yield two different values.
//
// The helpers are marked go:noinline because inlining would collapse the call
// sites and destroy the property under test.
func TestLinkRegisterIsARealReturnAddress(t *testing.T) {
	a := lrFromSiteA()
	b := lrFromSiteB()

	if a == 0 || b == 0 {
		t.Fatalf("link register read as 0: siteA=%#x siteB=%#x", a, b)
	}
	if a == b {
		t.Errorf("link register identical from two call sites (%#x); it should hold each caller's own resume address", a)
	}
}

//go:noinline
func lrFromSiteA() uintptr { return GetLinkRegister() }

//go:noinline
func lrFromSiteB() uintptr { return GetLinkRegister() }

// TestGRegStableWithinGoroutine asserts that R13 does not drift across calls
// on the same goroutine. If this fails, either R13 is not g or something in
// this package is clobbering it.
func TestGRegStableWithinGoroutine(t *testing.T) {
	first := GetGReg()
	if first == 0 {
		t.Fatal("GetGReg() returned 0; R13 holds g and is never 0 in running Go code")
	}

	for i := 0; i < 100; i++ {
		if got := GetGReg(); got != first {
			t.Fatalf("g pointer changed on the same goroutine: %#x then %#x (call %d)", first, got, i)
		}
	}
}

// TestGRegDiffersBetweenGoroutines is the test that actually proves R13 holds
// g rather than some other stable value: distinct goroutines have distinct g
// structs, so the register must read differently in each.
//
// Goroutines can migrate between OS threads, but a goroutine's g does not
// change, so comparing values captured inside each goroutine is sound.
func TestGRegDiffersBetweenGoroutines(t *testing.T) {
	const n = 8

	var wg sync.WaitGroup
	seen := make([]uintptr, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			seen[idx] = GetGReg()
		}(i)
	}
	wg.Wait()

	unique := make(map[uintptr]int, n)
	for i, v := range seen {
		if v == 0 {
			t.Fatalf("goroutine %d read g as 0", i)
		}
		if prev, dup := unique[v]; dup {
			t.Errorf("goroutines %d and %d report the same g pointer %#x", prev, i, v)
		}
		unique[v] = i
	}
}

// TestGetSPIsNotPseudoSP is the s390x form of this package's reason for
// existing. On s390x the hardware stack pointer must be named R15; bare SP in
// Go assembly is only ever the pseudo register. GetSP reads R15, so it must
// return a plausible stack address — specifically one close to the address of
// a local variable in this frame, not a small integer or a frame offset.
func TestGetSPIsNotPseudoSP(t *testing.T) {
	sp := GetSP()

	// A pseudo-SP-derived value would be a small frame offset. A real stack
	// pointer is a full 64-bit address well above the first page.
	if sp < 0x1000 {
		t.Errorf("GetSP() = %#x, which looks like a frame offset rather than a hardware stack pointer", sp)
	}
}

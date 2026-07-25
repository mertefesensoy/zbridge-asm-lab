package zbridge

import (
	"errors"
	"strings"
	"testing"
)

// TestHasCodeDistinguishesNoReturnCodeFromZero is the test that encodes this
// module's most important documented finding.
//
// GC28-0683-2 p.210: a single-line WTO issues NO return code. If Error modelled
// the return code as a bare int32, "the service returned nothing" and "the
// service returned 0, meaning success" would be the same value, and a caller
// could not tell them apart. HasCode is what keeps them distinct.
func TestHasCodeDistinguishesNoReturnCodeFromZero(t *testing.T) {
	noCode := &Error{Op: "console.WTO", Service: "SVC 35", Err: ErrServiceFailed}
	zeroCode := &Error{Op: "console.WTO", Service: "SVC 35", Code: 0, HasCode: true, Err: ErrServiceFailed}

	if noCode.HasCode {
		t.Error("an error built without a return code reports HasCode true")
	}
	if !zeroCode.HasCode {
		t.Error("an error built with an explicit zero return code reports HasCode false")
	}
	if noCode.Code != zeroCode.Code {
		t.Fatal("test setup is wrong: both Codes should be 0 so that only HasCode separates them")
	}

	// The rendered messages must differ, or the distinction is invisible in
	// logs — which is where most people will actually meet it.
	if noCode.Error() == zeroCode.Error() {
		t.Errorf("both render as %q; the no-return-code case must be visibly different", noCode.Error())
	}
	if strings.Contains(noCode.Error(), "rc=") {
		t.Errorf("no-return-code error renders a return code: %q", noCode.Error())
	}
	if !strings.Contains(zeroCode.Error(), "rc=0") {
		t.Errorf("explicit zero return code is not rendered: %q", zeroCode.Error())
	}
}

func TestErrorUnwraps(t *testing.T) {
	err := error(&Error{Op: "console.WTO", Err: ErrLayoutUnverified})
	if !errors.Is(err, ErrLayoutUnverified) {
		t.Error("errors.Is failed to match the wrapped sentinel")
	}
	if errors.Is(err, ErrUnsupportedPlatform) {
		t.Error("errors.Is matched the wrong sentinel")
	}
}

func TestErrorNamesTheBlockingUnknown(t *testing.T) {
	err := Unverified("console.WTO", "SVC 35", "U2")

	var zerr *Error
	if !errors.As(err, &zerr) {
		t.Fatalf("Unverified returned %T, want *Error", err)
	}
	if zerr.Unknown != "U2" {
		t.Errorf("Unknown = %q, want %q", zerr.Unknown, "U2")
	}
	if !strings.Contains(err.Error(), "blocked on U2") {
		t.Errorf("error message %q does not name the blocking unknown", err.Error())
	}
}

func TestUnsupportedNamesU3b(t *testing.T) {
	// The toolchain unknown, not the storage one: on a non-z/OS build the
	// reason nothing works is that this is the wrong platform entirely.
	err := Unsupported("storage.Malloc31", "LE __malloc31")

	var zerr *Error
	if !errors.As(err, &zerr) {
		t.Fatalf("Unsupported returned %T, want *Error", err)
	}
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Error("Unsupported does not wrap ErrUnsupportedPlatform")
	}
	if zerr.Unknown != "U3b" {
		t.Errorf("Unknown = %q, want %q", zerr.Unknown, "U3b")
	}
}

func TestSupportedIsFalseOffZOS(t *testing.T) {
	// This test is a tripwire, not a tautology. If it ever fails, the build is
	// targeting zos/s390x — which as of go1.26.3 upstream Go cannot do — and
	// that fact needs to reach a human.
	if Supported() {
		t.Fatalf("Supported() is true on %s; if this is a real z/OS build, update ADR 0003 §7 and evidence F3", Platform())
	}
}

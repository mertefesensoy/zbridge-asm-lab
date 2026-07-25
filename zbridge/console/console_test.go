package console

import (
	"errors"
	"strings"
	"testing"

	"github.com/mertefesensoy/zbridge"
)

// These tests run on every platform. That is the point of the E3 seam: the
// parts of the WTO path that do not need z/OS are tested on a laptop.

func TestEncodeWPLRefusesRatherThanGuesses(t *testing.T) {
	_, err := EncodeWPL("ZBRIDGE TEST HELLO")
	if err == nil {
		t.Fatal("EncodeWPL returned a parameter list, but the WPL layout has no primary citation; returning bytes here would be a guess")
	}
	if !errors.Is(err, zbridge.ErrLayoutUnverified) {
		t.Errorf("EncodeWPL error = %v, want ErrLayoutUnverified", err)
	}

	var zerr *zbridge.Error
	if !errors.As(err, &zerr) {
		t.Fatalf("EncodeWPL error is %T, want *zbridge.Error", err)
	}
	if zerr.Unknown != "U2" {
		t.Errorf("blocked-on unknown = %q, want %q", zerr.Unknown, "U2")
	}
	if zerr.HasCode {
		t.Error("HasCode is true, but no service was invoked and no return code exists")
	}
}

// TestValidationHappensBeforeTheLayoutError is the test that keeps the refusal
// honest. If EncodeWPL simply returned ErrLayoutUnverified as its first
// statement, the validation code would be dead and would rot unnoticed until the
// layout landed. Bad input must produce a validation error, not the layout error.
func TestValidationHappensBeforeTheLayoutError(t *testing.T) {
	cases := []struct {
		name string
		msg  string
	}{
		{"empty", ""},
		{"too long", strings.Repeat("x", MaxTextLen+1)},
		{"control byte", "hello\x00world"},
		{"newline", "hello\nworld"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := EncodeWPL(c.msg)
			if err == nil {
				t.Fatal("expected an error")
			}
			if errors.Is(err, zbridge.ErrLayoutUnverified) {
				t.Errorf("got the layout error for invalid input %q; validation must run first", c.msg)
			}
		})
	}
}

func TestMaxTextLenBoundary(t *testing.T) {
	// Exactly at the limit must pass validation and fail only on the layout.
	_, err := EncodeWPL(strings.Repeat("x", MaxTextLen))
	if !errors.Is(err, zbridge.ErrLayoutUnverified) {
		t.Errorf("a message of exactly MaxTextLen (%d) failed validation: %v", MaxTextLen, err)
	}

	// One over must fail validation.
	_, err = EncodeWPL(strings.Repeat("x", MaxTextLen+1))
	if !errors.Is(err, zbridge.ErrMessageTooLong) {
		t.Errorf("MaxTextLen+1 error = %v, want ErrMessageTooLong", err)
	}
}

func TestWTORefusesBeforeTrapping(t *testing.T) {
	err := WTO("ZBRIDGE TEST HELLO")
	if err == nil {
		t.Fatal("WTO returned nil on a platform where nothing has ever reached a console")
	}
	// The ordering matters: WTO must fail on the unverified layout BEFORE it
	// reaches the platform stub. Trapping with a guessed parameter list is the
	// failure mode this ordering exists to prevent.
	if !errors.Is(err, zbridge.ErrLayoutUnverified) {
		t.Errorf("WTO error = %v, want ErrLayoutUnverified (refuse before trapping)", err)
	}
}

func TestRouteMask(t *testing.T) {
	// Route codes are 1-based with code 1 as the most significant bit of a
	// 16-bit mask. This is pure arithmetic and does not depend on the layout.
	cases := []struct {
		name  string
		codes []RouteCode
		want  uint16
	}{
		{"none", nil, 0x0000},
		{"route 1", []RouteCode{1}, 0x8000},
		{"route 2", []RouteCode{2}, 0x4000},
		{"route 16", []RouteCode{16}, 0x0001},
		{"routes 1 and 2", []RouteCode{1, 2}, 0xC000},
		{"out of range ignored", []RouteCode{0, 17, 200}, 0x0000},
		{"master + programmer", []RouteCode{RouteMasterConsole, RouteProgrammerInfo}, 0x8020},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			o := resolve([]Option{WithRoute(c.codes...)})
			if got := o.routeMask(); got != c.want {
				t.Errorf("routeMask(%v) = %#04x, want %#04x", c.codes, got, c.want)
			}
		})
	}
}

func TestDescriptorMask(t *testing.T) {
	o := resolve([]Option{WithDescriptor(DescEventualAction)})
	// Descriptor code 3 -> bit 16-3 = 13 -> 0x2000
	if got, want := o.descriptorMask(), uint16(0x2000); got != want {
		t.Errorf("descriptorMask(3) = %#04x, want %#04x", got, want)
	}
}

// TestFormatDC checks the other half of the E3 seam. FormatDC output is pasted
// straight into an MVS assembler program, so its shape matters: a label in
// columns 1-8, DC starting in the operation field, and hex wrapped so the line
// fits fixed-form source.
func TestFormatDC(t *testing.T) {
	got := FormatDC("WPL", []byte{0x00, 0x05, 0xC8, 0xC5, 0xD3, 0xD3, 0xD6})
	want := "WPL      DC    X'0005C8C5D3D3D6'\n"
	if got != want {
		t.Errorf("FormatDC =\n%q\nwant\n%q", got, want)
	}
}

func TestFormatDCWraps(t *testing.T) {
	b := make([]byte, 40)
	for i := range b {
		b[i] = byte(i)
	}
	got := FormatDC("BUF", b)

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("40 bytes produced %d lines, want 3 (16+16+8)", len(lines))
	}
	if !strings.HasPrefix(lines[0], "BUF ") {
		t.Errorf("first line %q does not start with the label", lines[0])
	}
	if !strings.HasPrefix(lines[1], "     ") {
		t.Errorf("continuation line %q should have no label", lines[1])
	}
	if !strings.Contains(lines[2], "X'2021222324252627'") {
		t.Errorf("last line %q does not hold the final 8 bytes", lines[2])
	}
}

// TestLayoutStatusIsHonest guards the project's own convention. If someone sets
// LayoutVerified to true without supplying a layout, EncodeWPL panics by design;
// this test makes the inconsistency visible at test time instead.
func TestLayoutStatusIsHonest(t *testing.T) {
	if LayoutVerified {
		t.Fatal("LayoutVerified is true — supply the layout constants, cite the source in LayoutStatus, and update ADR 0003 §4 before flipping this")
	}
	if !strings.Contains(LayoutStatus, "UNVERIFIED") {
		t.Error("LayoutStatus must say UNVERIFIED while LayoutVerified is false")
	}
}

package console

import (
	"errors"
	"strings"
	"testing"

	"github.com/mertefesensoy/zbridge"
)

// These tests run on every platform. That is the point of the E3 seam: the parts
// of the WTO path that do not need z/OS are tested on a laptop.
//
// The byte-exact comparison against IBM's own macro expansion lives in
// wpl_oracle_test.go, which is the stronger test. This file covers the
// surrounding behaviour: validation, option handling, mask arithmetic, and the
// platform boundary.

// TestEncodeWPLSucceedsForTheMinimalForm replaces an earlier test that asserted
// EncodeWPL must NOT return bytes. That assertion was correct while the layout
// had no primary citation, and became obsolete on 2026-07-26 when the layout was
// read out of the IFOX00 expansion of IBM's WTO macro on MVS 3.8j.
func TestEncodeWPLSucceedsForTheMinimalForm(t *testing.T) {
	msg := "ZBRIDGE TEST HELLO"
	buf, err := EncodeWPL(msg)
	if err != nil {
		t.Fatalf("EncodeWPL returned %v", err)
	}
	if len(buf) != len(msg)+4 {
		t.Errorf("buffer is %d bytes, want %d", len(buf), len(msg)+4)
	}
}

// TestValidationRunsBeforeEncoding keeps the validation path honest: bad input
// must produce a specific validation error, not a silently mangled buffer.
func TestValidationRunsBeforeEncoding(t *testing.T) {
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
			buf, err := EncodeWPL(c.msg)
			if err == nil {
				t.Fatalf("EncodeWPL accepted invalid input and returned % X", buf)
			}
			if errors.Is(err, zbridge.ErrLayoutUnverified) {
				t.Errorf("got the layout error for invalid input %q; validation must run first", c.msg)
			}
		})
	}
}

func TestMaxTextLenBoundary(t *testing.T) {
	// Exactly at the limit must be accepted.
	if _, err := EncodeWPL(strings.Repeat("x", MaxTextLen)); err != nil {
		t.Errorf("a message of exactly MaxTextLen (%d) was rejected: %v", MaxTextLen, err)
	}

	// One over must fail validation.
	_, err := EncodeWPL(strings.Repeat("x", MaxTextLen+1))
	if !errors.Is(err, zbridge.ErrMessageTooLong) {
		t.Errorf("MaxTextLen+1 error = %v, want ErrMessageTooLong", err)
	}
}

// TestWTOStopsAtThePlatformBoundary. WTO now builds a real parameter list, so it
// no longer fails on the layout — it fails because this is not z/OS. That is the
// correct and honest outcome on a developer's laptop.
func TestWTOStopsAtThePlatformBoundary(t *testing.T) {
	err := WTO("ZBRIDGE TEST HELLO")
	if err == nil {
		t.Fatal("WTO returned nil on a platform where nothing has ever reached a console")
	}
	if !errors.Is(err, zbridge.ErrUnsupportedPlatform) {
		t.Errorf("WTO error = %v, want ErrUnsupportedPlatform", err)
	}

	var zerr *zbridge.Error
	if !errors.As(err, &zerr) {
		t.Fatalf("WTO error is %T, want *zbridge.Error", err)
	}
	if zerr.HasCode {
		t.Error("HasCode is true, but no service was invoked and no return code exists")
	}
}

// TestWTORejectsBadInputBeforeThePlatformCheck. Ordering matters: a caller with a
// bad message should learn that, not that they are on the wrong operating system.
func TestWTORejectsBadInputBeforeThePlatformCheck(t *testing.T) {
	err := WTO(strings.Repeat("x", MaxTextLen+1))
	if !errors.Is(err, zbridge.ErrMessageTooLong) {
		t.Errorf("WTO with an over-long message gave %v, want ErrMessageTooLong", err)
	}
}

func TestRouteMask(t *testing.T) {
	// Route codes are 1-based with code 1 as the most significant bit of a
	// 16-bit mask. IBM's macro expansion confirmed this convention directly:
	// ROUTCDE=(11) generated B'0000000000100000' = 0x0020, which is what this
	// implementation produces.
	cases := []struct {
		name  string
		codes []RouteCode
		want  uint16
	}{
		{"none", nil, 0x0000},
		{"route 1", []RouteCode{1}, 0x8000},
		{"route 2", []RouteCode{2}, 0x4000},
		{"route 11 (macro-confirmed)", []RouteCode{11}, 0x0020},
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

// TestDescriptorMask. IBM's macro expansion confirmed this too: DESC=(7)
// generated B'0000001000000000' = 0x0200.
func TestDescriptorMask(t *testing.T) {
	o := resolve([]Option{WithDescriptor(7)})
	if got, want := o.descriptorMask(), uint16(0x0200); got != want {
		t.Errorf("descriptorMask(7) = %#04x, want %#04x (as IBM's macro generated)", got, want)
	}
}

// TestFormatDC checks the other half of the E3 seam. FormatDC output is pasted
// straight into an MVS assembler program, so its shape matters: a label in
// columns 1-8, DC in the operation field, and hex wrapped to fit fixed-form
// source.
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

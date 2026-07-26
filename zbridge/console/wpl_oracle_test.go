package console

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/mertefesensoy/zbridge"
)

// This file is the strongest test in the module, and it is worth saying why.
//
// The expected bytes below were not derived from a specification, a manual, or
// reasoning. They were produced by IBM's own WTO macro, assembled by IFOX00 on
// MVS 3.8j, and read out of the assembler listing:
//
//	000000 0016                DC AL2(22)   TEXT LENGTH
//	000002 0000                DC B'0...0'  MCS FLAGS
//	000004 E9C2D9C9C4C7C540    DC C'ZBRIDGE TEST HELLO'
//	00000C E3C5E2E340C8C5D3
//	000014 D3D6
//
// So this is not a unit test asserting that the code does what the code says.
// It asserts that our Go implementation produces byte-for-byte what a real IBM
// macro produced on a real system — which is the entire Go side of rung E3,
// executable on a laptop.
//
// Evidence: docs/evidence/E1-wto-macro-expansion-2026-07-26.md

// macroOracle is the exact parameter list IBM's macro generated for
// WTO 'ZBRIDGE TEST HELLO',MF=L — 18 characters of text, length field 0x0016.
const macroOracle = "0016" + "0000" +
	"E9C2D9C9C4C7C540" + "E3C5E2E340C8C5D3" + "D3D6"

// macroOracleLong is the same for the 38-character message
// 'ZBRIDGE TEST A LONGER OPERATOR MESSAGE', length field 0x002A.
const macroOracleLong = "002A" + "0000" +
	"E9C2D9C9C4C7C540" + "E3C5E2E340C140D3" + "D6D5C7C5D940D6D7" +
	"C5D9C1E3D6D940D4" + "C5E2E2C1C7C5"

func mustDecode(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad oracle constant: %v", err)
	}
	return b
}

// TestEncodeWPLMatchesIBMMacro is the central correctness claim of this package.
func TestEncodeWPLMatchesIBMMacro(t *testing.T) {
	want := mustDecode(t, macroOracle)

	got, err := EncodeWPL("ZBRIDGE TEST HELLO")
	if err != nil {
		t.Fatalf("EncodeWPL returned %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("EncodeWPL disagrees with IBM's macro expansion:\n got % X\nwant % X", got, want)
	}
}

func TestEncodeWPLMatchesIBMMacroLongText(t *testing.T) {
	want := mustDecode(t, macroOracleLong)

	got, err := EncodeWPL("ZBRIDGE TEST A LONGER OPERATOR MESSAGE")
	if err != nil {
		t.Fatalf("EncodeWPL returned %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("EncodeWPL disagrees with IBM's macro expansion at 38 chars:\n got % X\nwant % X", got, want)
	}
}

// TestLengthFieldIsTextPlusFour checks the rule rather than the two fixtures, so
// a regression in the arithmetic is caught at every length rather than only at
// the two the macro happened to be run with.
func TestLengthFieldIsTextPlusFour(t *testing.T) {
	for n := 1; n <= MaxTextLen; n++ {
		msg := strings.Repeat("A", n)
		buf, err := EncodeWPL(msg)
		if err != nil {
			t.Fatalf("len %d: %v", n, err)
		}
		if len(buf) != n+4 {
			t.Fatalf("len %d: buffer is %d bytes, want %d", n, len(buf), n+4)
		}
		got := int(buf[0])<<8 | int(buf[1])
		if got != n+4 {
			t.Errorf("len %d: length field = %d, want %d", n, got, n+4)
		}
	}
}

// TestMCSFlagsAreZeroForMinimalForm pins the flags field. The macro generated
// B'0000000000000000' for the minimal form; anything else here would be an
// invention.
func TestMCSFlagsAreZeroForMinimalForm(t *testing.T) {
	buf, err := EncodeWPL("ZBRIDGE TEST HELLO")
	if err != nil {
		t.Fatal(err)
	}
	if buf[2] != 0x00 || buf[3] != 0x00 {
		t.Errorf("MCS flags = % X, want 00 00 (as IBM's macro generated)", buf[2:4])
	}
}

// TestRoutingIsRefusedNotDropped guards the decision recorded in ADR 0003 §4.
//
// The macro expansion showed the routed form exists — MCS flag 0x8000 with
// descriptor and routing halfwords appended after the text — but brief 003 Q3
// found no citable MCS bit table, so z/OS bit assignments are unconfirmed.
// Silently producing a list without the caller's routing would send an operator
// message somewhere they did not ask for.
func TestRoutingIsRefusedNotDropped(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
	}{
		{"route", WithRoute(RouteProgrammerInfo)},
		{"descriptor", WithDescriptor(DescApplicationStatus)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf, err := EncodeWPL("ZBRIDGE TEST HELLO", c.opt)
			if err == nil {
				t.Fatalf("EncodeWPL accepted %s and returned % X; it must refuse until the MCS bits are cited", c.name, buf)
			}
			if !errors.Is(err, zbridge.ErrLayoutUnverified) {
				t.Errorf("error = %v, want ErrLayoutUnverified", err)
			}
		})
	}
}

// TestLengthFieldExcludesRoutingHalfwords documents the trap the macro revealed,
// as an executable note rather than a comment.
//
// In the routed form the length field STILL read 22 even though four more bytes
// followed the text. So "length = whole parameter list" is wrong. This test
// asserts the property we rely on for the minimal form — buffer length is exactly
// the length field — and will need revisiting if the routed form is ever
// implemented.
func TestLengthFieldEqualsBufferLengthInMinimalForm(t *testing.T) {
	buf, err := EncodeWPL("ZBRIDGE TEST HELLO")
	if err != nil {
		t.Fatal(err)
	}
	field := int(buf[0])<<8 | int(buf[1])
	if field != len(buf) {
		t.Errorf("length field %d != buffer length %d; in the MINIMAL form these coincide", field, len(buf))
	}
}

// TestFormatDCProducesTheMacroBytes closes the E3 loop: the hex that goes into an
// assembler DC statement must be the hex IBM's macro produced.
func TestFormatDCProducesTheMacroBytes(t *testing.T) {
	buf, err := EncodeWPL("ZBRIDGE TEST HELLO")
	if err != nil {
		t.Fatal(err)
	}
	dc := FormatDC("WPL", buf)

	// FormatDC wraps at 16 bytes per DC statement, so the parameter list splits
	// as 16 + 6. Both halves must be exactly the macro's bytes.
	if !strings.Contains(dc, "X'00160000E9C2D9C9C4C7C540E3C5E2E3'") {
		t.Errorf("FormatDC first line does not carry the macro's leading 16 bytes:\n%s", dc)
	}
	if !strings.Contains(dc, "X'40C8C5D3D3D6'") {
		t.Errorf("FormatDC continuation does not carry the macro's trailing 6 bytes:\n%s", dc)
	}

	// And the concatenation of all the hex must equal the oracle exactly.
	var all strings.Builder
	for _, line := range strings.Split(strings.TrimRight(dc, "\n"), "\n") {
		if i := strings.Index(line, "X'"); i >= 0 {
			all.WriteString(strings.TrimSuffix(line[i+2:], "'"))
		}
	}
	if got, want := all.String(), strings.ToUpper(macroOracle); got != want {
		t.Errorf("FormatDC hex = %s, want %s", got, want)
	}
}

func TestLayoutStatusReflectsVerification(t *testing.T) {
	if !LayoutVerified {
		t.Fatal("LayoutVerified is false but the layout was verified from the macro expansion on 2026-07-26")
	}
	if !strings.Contains(LayoutStatus, "VERIFIED") {
		t.Error("LayoutStatus must say VERIFIED while LayoutVerified is true")
	}
	// The status must keep naming the limit of what was verified.
	if !strings.Contains(LayoutStatus, "minimal single-line form") {
		t.Error("LayoutStatus must state that only the minimal single-line form is verified")
	}
}

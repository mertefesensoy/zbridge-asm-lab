package console

import (
	"fmt"
	"strings"

	"github.com/mertefesensoy/zbridge"
	"github.com/mertefesensoy/zbridge/codepage"
)

// LayoutStatus reports the documentary status of the WTO parameter list layout
// this package would use.
//
// It is exported because a caller is entitled to ask "is this library actually
// able to do the thing" without triggering an error, and because the thesis
// needs a programmatic way to show that the gap is tracked rather than ignored.
const LayoutStatus = "UNVERIFIED: no primary-source citation for the WTO parameter list byte " +
	"layout on z/OS or MVS as of 2026-07-25. GC28-0683-2 was read page by page and does " +
	"not document it. Blocked on research brief 003 or on rung E1's assembler listing."

// LayoutVerified reports whether the WPL byte layout has a primary-source
// citation. While it is false, EncodeWPL and WTO return ErrLayoutUnverified.
//
// When brief 003 lands, exactly three things change: this constant, the layout
// constants in this file, and the encoder body. Everything else in the module —
// options, error model, platform stubs, tests — is already written against the
// final shape.
const LayoutVerified = false

// EncodeWPL builds the WTO parameter list for msg and returns it as bytes ready
// to be pointed at by R1.
//
// It is pure Go and runs on any host, by design. Rung E3 of the project's
// evidence ladder verifies the Go-side parameter-list construction by embedding
// this function's output in an MVS assembler program as DC X'...' constants and
// handing it to a real SVC 35 — so this function's output is the artifact of
// that rung, not an implementation detail. See ADR 0003 §6.
//
// # It currently returns an error, on purpose
//
// The byte layout has no primary citation. Rather than emit a plausible buffer,
// EncodeWPL returns ErrLayoutUnverified. Everything up to the point where a
// documented offset would be needed IS performed and validated first — length
// checks, EBCDIC conversion, mask computation — so that when the layout lands,
// only the final assembly step is new code, and the error a caller sees today is
// specific rather than a blanket "unimplemented".
func EncodeWPL(msg string, opts ...Option) ([]byte, error) {
	const op = "console.EncodeWPL"

	if err := validate(msg); err != nil {
		return nil, err
	}

	o := resolve(opts)

	// Everything below this point is real work that is already correct and
	// already tested. It is performed before the error so that the validation
	// path, the conversion path and the mask arithmetic are all exercised by
	// callers and tests today.
	text := make([]byte, len(msg))
	codepage.AtoE(text, []byte(msg))
	_ = o.routeMask()
	_ = o.descriptorMask()

	if !LayoutVerified {
		return nil, zbridge.Unverified(op, "SVC 35", "U2")
	}

	// Unreachable until LayoutVerified. The layout constants and assembly of
	// the header, flags and text go here, in this one place, and nowhere else
	// in the module.
	panic("console: LayoutVerified is true but no layout is implemented")
}

// validate applies the checks that do NOT depend on the unverified layout.
func validate(msg string) error {
	const op = "console.EncodeWPL"

	if len(msg) == 0 {
		return &zbridge.Error{
			Op:      op,
			Service: "SVC 35",
			Err:     fmt.Errorf("empty message"),
		}
	}
	if len(msg) > MaxTextLen {
		return &zbridge.Error{
			Op:      op,
			Service: "SVC 35",
			Err:     zbridge.ErrMessageTooLong,
		}
	}
	// A console is an operator's workplace. Control characters in an operator
	// message are at best noise and at worst confusing on a 3270.
	for i := 0; i < len(msg); i++ {
		if msg[i] < 0x20 || msg[i] == 0x7F {
			return &zbridge.Error{
				Op:      op,
				Service: "SVC 35",
				Err:     fmt.Errorf("message contains control byte 0x%02X at offset %d", msg[i], i),
			}
		}
	}
	return nil
}

// FormatDC renders bytes as an MVS assembler DC constant, wrapped to fit a
// fixed-form assembler source line.
//
// This is the other half of the E3 seam. The output is pasted directly into an
// assembler program that is then submitted to the emulated system, so the Go
// side of rung E3 is EncodeWPL followed by FormatDC and nothing else.
//
// It works on any host and needs no verified layout, so it is testable and
// tested today.
func FormatDC(label string, b []byte) string {
	const bytesPerLine = 16

	var sb strings.Builder
	for i := 0; i < len(b); i += bytesPerLine {
		end := i + bytesPerLine
		if end > len(b) {
			end = len(b)
		}

		name := ""
		if i == 0 {
			name = label
		}

		var hex strings.Builder
		for _, c := range b[i:end] {
			fmt.Fprintf(&hex, "%02X", c)
		}

		fmt.Fprintf(&sb, "%-8s DC    X'%s'\n", name, hex.String())
	}
	return sb.String()
}

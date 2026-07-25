// Package console writes messages to the z/OS operator console via WTO
// (Write To Operator, SVC 35) and its reply-taking variant WTOR.
//
// This is the endgame of the zbridge project: WTO has no public Go-assembly
// implementation, and closing that gap is the contribution.
//
// # Nothing in this package has reached a console
//
// Read this before using anything here. The package is complete in structure
// and deliberately incomplete in one specific place: the WTO parameter list
// (WPL) byte layout.
//
// As of 2026-07-25 the WPL layout has NO primary-source citation on either z/OS
// or its ancestor. Research brief 001 returned a layout attributed to
// GC28-0683-2; the manual was then retrieved and read page by page, and it does
// not document the byte layout at all — the quotation the return attributed to
// it is not in it (docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md).
// Research brief 003 is the outstanding gap-closer.
//
// So EncodeWPL returns ErrLayoutUnverified rather than a plausible-looking
// buffer. That is a deliberate design decision recorded in ADR 0003 §4: a
// library that returns a guess is worse than one that returns an error, because
// the guess would appear to work right up until an operator console somewhere
// showed the wrong thing.
//
// # The seam, and why the encoder is separate from the call
//
// EncodeWPL is pure Go and runs on any host. That is not an accident of
// factoring — it is the shape the project's evidence ladder requires.
//
// Rung E3 (ADR 0001 §6) verifies the Go-side parameter-list construction
// WITHOUT Go ever running on a mainframe: a program on the laptop emits the
// bytes it intends to hand to SVC 35, those bytes are embedded in an MVS
// assembler program as DC X'...' constants, and a real SVC 35 is given them. If
// the console displays the message, the Go-side construction is verified against
// a genuine supervisor call.
//
// EncodeWPL plus FormatDC is the entire Go side of that rung. Splitting
// construction from invocation is what makes an unknown testable before the
// platform exists.
package console

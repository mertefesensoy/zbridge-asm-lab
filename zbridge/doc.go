// Package zbridge is a pure-Go bridge from Go programs to IBM z/OS system
// services, implemented in Go (Plan 9) assembly for s390x with no cgo.
//
// # What this is
//
// z/OS exposes its services through supervisor calls and macro-generated
// parameter lists, which is an assembly-language interface. Reaching it from Go
// has historically meant cgo and a C compiler, which drags in the Language
// Environment and gives up Go's static-binary story. ibmruntimes/go-recordio
// proved the alternative is viable — Go assembly can call an MVS service
// directly — but it covers only IEFSSREQ. Operator-facing services such as WTO
// have no public Go-assembly implementation. That gap is what this module
// exists to close.
//
// # Status: nothing here has run on z/OS
//
// This is honest scaffolding, not a shipped library. Read
// docs/decisions/0003-production-bridge-module-architecture.md before depending
// on anything. The rule that governs every exported function is:
//
//	Every exported function whose behaviour depends on an unretired unknown
//	returns a typed error naming that unknown — on every platform, including
//	z/OS — until the corresponding evidence rung passes.
//
// So the API is complete and the errors are specific. What is NOT here is a
// parameter-list byte layout: the WTO parameter list has no primary citation on
// any system yet (research brief 003 is outstanding), and inventing one that
// looked plausible would be worse than returning an error, because it would
// produce a library that appears to work.
//
// What does work today, on any host:
//
//	codepage  EBCDIC IBM-1047 conversion, verified on linux/s390x under QEMU
//	          against every one of the 256 table entries
//
// # Layout
//
//	codepage/          EBCDIC ⇄ ASCII, usable standalone
//	console/           WTO and WTOR — the endgame
//	dataset/           record I/O — declared, not scoped
//	subsys/            IEFSSREQ subsystem requests — declared, not scoped
//	internal/svc/      raw supervisor-call primitives
//	internal/linkage/  MVS register and save-area conventions
//	internal/storage/  Malloc31 / below-the-bar allocation
//
// # No cgo, anywhere
//
// The no-cgo constraint is not a preference. It is the point: a cgo bridge to
// z/OS already exists in every shop that wants one, and it costs the Language
// Environment dependency and the static binary. This module is only interesting
// if it never needs a C compiler.
package zbridge

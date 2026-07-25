//go:build zos && s390x

package console

import (
	"github.com/mertefesensoy/zbridge"
	"github.com/mertefesensoy/zbridge/internal/svc"
)

// issueWTO hands the parameter list to SVC 35.
//
// # This file is compiled by nothing today
//
// Upstream Go cannot target z/OS: `go tool dist list` (go1.26.3) offers
// linux/s390x and no zos/s390x. The string "zos" is in internal/syslist, so this
// file's build constraint parses, but internal/platform does not list the
// target. IBM's Go fork is a hard dependency and it is not yet in hand. See
// docs/evidence/E-L-s390x-port-qemu-2026-07-25.md finding F3.
//
// The file exists anyway because the shape of the call is knowable now and the
// unknowns are better named in code than left implicit.
//
// # What is still missing beyond the toolchain
//
//  1. The WPL byte layout (U2). EncodeWPL refuses before this function is ever
//     reached, so wpl is never a guess.
//  2. Below-the-bar storage (U3a). The parameter list must be addressable by the
//     service, which on z/OS means AMODE 31 and storage below the 2 GB bar —
//     internal/storage.Malloc31, the module's one Language Environment
//     touchpoint, following go-recordio's pattern.
//  3. AMODE switching (U3a). go-recordio's utils.s performs SAM31/SAM64 around
//     the service call. Annotating that file is Phase 2 of the roadmap and it
//     gates rung T3.
//
// The return-code handling below reflects the H001 C4 finding: a single-line WTO
// on the ancestor system issues NO return code, and R1 returns the message
// identification number instead. Until research brief 003 Q4 settles what z/OS
// does, this function reports HasCode=false rather than inventing a contract.
func issueWTO(wpl []byte) error {
	if len(wpl) == 0 {
		return zbridge.Unverified("console.WTO", "SVC 35", "U2")
	}

	// TODO(U3a): copy wpl into storage.Malloc31'd memory below the bar and
	// switch to AMODE 31 before the call. Passing a Go heap pointer directly is
	// wrong on z/OS and would be wrong silently, which is why this returns
	// rather than trying.
	return &zbridge.Error{
		Op:      "console.WTO",
		Service: "SVC 35",
		Unknown: "U3a",
		Err:     zbridge.ErrUnsupportedPlatform,
	}

	// Unreachable until U3a is retired. The shape of the real path:
	//
	//	below, free, err := storage.Malloc31(len(wpl))
	//	if err != nil { return err }
	//	defer free()
	//	copy(below, wpl)
	//	rc, hasCode := svc.Call35(&below[0])
	//	if !hasCode { return nil }
	//	if rc != 0 { return &zbridge.Error{..., Code: rc, HasCode: true} }
	//	return nil
	_ = svc.Supported
}

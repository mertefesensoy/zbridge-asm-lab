// Package subsys issues subsystem requests via IEFSSREQ.
//
// # Status: declared, not scoped
//
// ADR 0003 commits to declaring the three service families an enterprise bridge
// implies. It does not commit to their operations, and this one in particular
// should not be invented: IEFSSREQ is the surface ibmruntimes/go-recordio
// already implements, and go-recordio is this project's architectural blueprint
// (BSD-3-Clause).
//
// The right next step for this package is Phase 2 of the roadmap — annotating
// go-recordio's utils/utils.s — which is reading work, needs no hardware, and
// gates rung T3 anyway. Designing an API before that annotation exists would
// mean designing against a guess about a codebase that is sitting there readable.
//
// The Request type below fixes only what is structurally certain: a subsystem
// request carries a function code and a parameter block, and returns a return
// code plus a reason code.
package subsys

import "github.com/mertefesensoy/zbridge"

// Request is a subsystem request.
type Request struct {
	// Function is the SSOB function code identifying the request.
	Function uint16

	// Parms is the function-dependent parameter block.
	Parms []byte
}

// Response is the result of a subsystem request.
type Response struct {
	Code    int32
	Reason  uint32
	HasCode bool
}

// Do issues req via IEFSSREQ.
func Do(req Request) (Response, error) {
	return Response{}, zbridge.Unsupported("subsys.Do", "IEFSSREQ")
}

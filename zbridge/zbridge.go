package zbridge

import "runtime"

// Version is the module version. Pre-1.0 and unstable by intent: no rung above
// U1 has passed, so the API may still move.
const Version = "0.0.1-dev"

// Supported reports whether the current build targets a platform on which this
// module's system-service paths can execute at all.
//
// It is a build-target question, not a runtime capability question. Supported
// returning true means the z/OS code paths were compiled in; it does not mean
// any of them has ever been executed successfully, and as of the date of this
// file none of them has.
func Supported() bool {
	return runtime.GOOS == "zos" && runtime.GOARCH == "s390x"
}

// Platform returns a human-readable description of the current target, for
// diagnostics and for evidence-file headers.
func Platform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

// Unsupported builds the standard error for an operation that requires z/OS on
// s390x. Sub-packages use it so every platform stub in the module fails the
// same way, with the same sentinel and the same named unknown.
func Unsupported(op, service string) error {
	return unsupported(op, service)
}

// Unverified builds the standard error for an operation that is fully
// implemented but depends on a parameter-list byte layout with no primary
// citation.
//
// The distinction from Unsupported matters to a caller: Unsupported means "wrong
// machine", Unverified means "right machine, but this project will not guess at
// a byte layout it cannot cite". See ADR 0003 §4.
func Unverified(op, service, unknown string) error {
	return &Error{
		Op:      op,
		Service: service,
		Unknown: unknown,
		Err:     ErrLayoutUnverified,
	}
}

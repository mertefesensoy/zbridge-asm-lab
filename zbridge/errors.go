package zbridge

import (
	"errors"
	"fmt"
)

// Sentinel errors. Callers match with errors.Is.
var (
	// ErrUnsupportedPlatform means the operation requires z/OS on s390x and the
	// program is not running there. Every non-z/OS build of this module returns
	// this rather than panicking: a library must be importable and testable on a
	// developer's laptop.
	ErrUnsupportedPlatform = errors.New("zbridge: operation requires z/OS on s390x")

	// ErrLayoutUnverified means the operation depends on a parameter-list byte
	// layout that has no primary-source citation yet, so this module refuses to
	// guess at it.
	//
	// This is deliberately not a "not implemented" error. The code path exists
	// and is complete; what is missing is documentary evidence for a byte
	// layout, and shipping a plausible-looking guess would produce a library
	// that appears to work. See research brief 003.
	ErrLayoutUnverified = errors.New("zbridge: parameter-list layout has no primary citation (research brief 003 outstanding)")

	// ErrNotAuthorized means the caller's authorization state is insufficient
	// for the requested service or option.
	ErrNotAuthorized = errors.New("zbridge: caller not authorized for this service")

	// ErrMessageTooLong means the message exceeds what the service accepts.
	ErrMessageTooLong = errors.New("zbridge: message too long for service")

	// ErrServiceFailed means the system service was invoked and reported
	// failure. Inspect the Error's Code and Reason, but check HasCode first.
	ErrServiceFailed = errors.New("zbridge: system service reported failure")
)

// Error is the structured error every operation in this module returns.
//
// # Why HasCode exists
//
// It would be natural to model an MVS service error as "a return code from
// R15", and natural would be wrong. Reading GC28-0683-2 p.210 directly
// established that on the ancestor system a single-line WTO issues NO return
// code at all: R1 comes back holding the 24-bit message identification number,
// and the documented 00-14 return-code table applies only to MLWTO. Whether
// z/OS behaves the same way is exactly what is in doubt, and it is one of the
// open questions in research brief 003.
//
// So "the service returned nothing to map" must be a representable state, not a
// zero value indistinguishable from success. HasCode reports whether Code means
// anything. A caller that reads Code without checking HasCode is reading a
// number the platform may never have produced.
type Error struct {
	// Op is the operation, e.g. "console.WTO".
	Op string

	// Service names the system service, e.g. "SVC 35".
	Service string

	// Code is the service return code. Valid only if HasCode is true.
	Code int32

	// HasCode reports whether the service supplied a return code at all.
	HasCode bool

	// Reason is a service reason code where one applies.
	Reason uint32

	// Unknown names the project unknown that blocks this operation, if any:
	// "U1", "U2", "U3a" or "U3b". Empty when the failure is a genuine runtime
	// failure rather than an unretired unknown.
	Unknown string

	// Err is the wrapped sentinel, for errors.Is.
	Err error
}

func (e *Error) Error() string {
	msg := e.Op
	if e.Service != "" {
		msg += " (" + e.Service + ")"
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	if e.HasCode {
		msg += fmt.Sprintf(" [rc=%d", e.Code)
		if e.Reason != 0 {
			msg += fmt.Sprintf(" reason=0x%X", e.Reason)
		}
		msg += "]"
	}
	if e.Unknown != "" {
		msg += " [blocked on " + e.Unknown + "]"
	}
	return msg
}

func (e *Error) Unwrap() error { return e.Err }

// unsupported builds the error every non-z/OS platform stub returns.
func unsupported(op, service string) error {
	return &Error{
		Op:      op,
		Service: service,
		Unknown: "U3b",
		Err:     ErrUnsupportedPlatform,
	}
}

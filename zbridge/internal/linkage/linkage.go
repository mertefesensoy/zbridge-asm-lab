// Package linkage documents and, eventually, implements standard MVS linkage
// for Go assembly routines that call system services.
//
// # The register collision, which is the whole reason this package exists
//
// MVS linkage and Go's s390x ABI both assign fixed meanings to registers, and
// they disagree on two of them:
//
//	Register  MVS linkage                     Go on s390x
//	R1        parameter list address          general use
//	R13       caller-provided 72-byte save    the goroutine pointer, g
//	          area address
//	R14       return address                  the link register — agrees
//	R15       entry point in, return code     the stack pointer
//	          out
//
// R13 and R15 are the collisions, and they are not cosmetic. Clobber R13 and the
// Go runtime cannot find the current goroutine. Clobber R15 and there is no
// stack. A routine that establishes MVS linkage must save both, establish the
// MVS values, call the service, and restore both before returning to Go.
//
// Go's own register assignment is not folklore here — it is
// cmd/internal/obj/s390x/a.out.go: REGG = REG_R13, REGSP = REG_R15.
//
// # The save area
//
// Standard MVS linkage passes a 72-byte (18-word) save area in R13, chained to
// the caller's. The convention is old and rigid: word 2 points back to the
// caller's save area, word 3 forward to the callee's, and words 4..17 hold R14,
// R15, R0..R12 in that order.
//
// SaveAreaSize is exported because the size is a fact worth stating once rather
// than as a magic number in assembly.
package linkage

// SaveAreaSize is the size in bytes of a standard MVS save area: 18 fullwords.
const SaveAreaSize = 72

// Register roles under standard MVS linkage, named so that assembly comments and
// Go code can refer to the same constants.
const (
	// RegParmList is R1: the address of the parameter list on entry to a
	// service.
	RegParmList = 1

	// RegSaveArea is R13: the address of the caller-provided save area. This
	// collides with Go's goroutine pointer.
	RegSaveArea = 13

	// RegReturnAddr is R14: the return address. Go uses R14 as the link
	// register, so this one convention agrees.
	RegReturnAddr = 14

	// RegEntryPoint is R15: the entry point address on the way in and the
	// return code on the way out. This collides with Go's stack pointer.
	RegEntryPoint = 15
)

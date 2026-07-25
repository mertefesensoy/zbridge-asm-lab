package regs

// This file is the s390x half of the split described in the package doc.
// There is no GetBP here, because s390x has no frame-pointer register. What it
// has instead are two registers with fixed, load-bearing roles that amd64
// spells differently or not at all.

// GetLinkRegister returns the value of R14, the link register, read inside
// GetLinkRegister's own frame. A call performed with BASR or BRASL leaves the
// return address in R14 rather than pushing it onto the stack the way amd64's
// CALL does, so the value returned here is the address in the CALLER that
// execution resumes at.
//
// This is the register the WTO path depends on. Standard MVS linkage returns
// through R14, and any Go assembly routine that issues a supervisor call must
// preserve it or arrange to restore it, or control never comes back.
func GetLinkRegister() uintptr

// GetGReg returns the value of R13, which Go reserves on s390x for the
// goroutine pointer g (cmd/internal/obj/s390x/a.out.go: REGG = REG_R13).
//
// It has no amd64 counterpart in this package because Go addresses g through a
// thread-local slot on amd64 rather than a dedicated register. On s390x it is a
// register, and it is the one that most often breaks hand-written assembly:
// clobber R13 and the runtime loses the current goroutine.
//
// The returned value is stable for the lifetime of a goroutine and differs
// between goroutines, which is what regs_s390x_test.go asserts.
func GetGReg() uintptr

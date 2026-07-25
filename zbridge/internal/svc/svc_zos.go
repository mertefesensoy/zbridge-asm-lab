//go:build zos && s390x

package svc

const supported = true

// Call35 issues SVC 35 with R1 pointing at the parameter list at wpl.
//
// It returns the service return code and whether the service supplied one at
// all. The second return value is not defensive style — on the ancestor system a
// single-line WTO issues no return code whatsoever, and whether z/OS differs is
// an open question (research brief 003 Q4). "Nothing was returned" has to be
// representable.
//
// The assembly body does not exist yet. Writing it before U3a (below-the-bar
// storage and AMODE switching) is settled would produce a routine that traps
// with a Go heap pointer in R1, which is wrong in a way that could disturb a
// live system rather than simply failing.
//
//go:noescape
func Call35(wpl *byte) (rc int32, hasCode bool)

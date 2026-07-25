// Package storage provides allocation that z/OS system services can address.
//
// # Why a Go slice is not good enough
//
// A parameter list handed to an MVS service must live where the service can
// reach it. On z/OS that means storage below the 2 GB bar, addressable in
// AMODE 31. Go's heap makes no such promise, so passing &slice[0] to a service
// is wrong — and wrong silently, which is worse than wrong loudly.
//
// The allocation therefore goes through Malloc31, which is the pattern
// ibmruntimes/go-recordio established. It is this module's single Language
// Environment touchpoint, and it is the one place where "no LE dependency"
// becomes "no LE dependency beyond this".
//
// # This is U3a and nothing emulates it
//
// Below-the-bar allocation, AMODE 31↔64 switching, and the LE entry point are
// z/OS behaviours. QEMU does not have them; MVS 3.8j is a 24-bit system where
// the equivalent constraint is the 16 MB line, which is the same category of
// problem with different numbers and no AMODE switching at all (ADR 0001,
// "what this decision does not claim").
//
// Annotating go-recordio's utils.s — the Malloc31 and SAM31/SAM64 sections in
// particular — is Phase 2 of the roadmap. It needs no hardware and it gates
// rung T3.
package storage

import "github.com/mertefesensoy/zbridge"

// Malloc31 allocates n bytes of storage addressable in AMODE 31, below the 2 GB
// bar, and returns the buffer together with a function that frees it.
//
// The returned free function must be called exactly once. It is not safe to
// retain the buffer after freeing it, and the buffer is not Go-managed memory:
// the garbage collector neither scans nor moves it, which is precisely why it is
// usable as a service parameter list.
func Malloc31(n int) (buf []byte, free func(), err error) {
	return nil, nil, zbridge.Unsupported("storage.Malloc31", "LE __malloc31")
}

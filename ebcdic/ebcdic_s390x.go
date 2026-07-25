//go:build s390x

package ebcdic

// AtoELoop and EtoALoop are byte-at-a-time reference implementations of AtoE
// and EtoA, written in s390x assembly as a direct transliteration of the amd64
// loop. They exist for one reason: AtoE and EtoA are driven by a hand-encoded
// TR instruction (Go's s390x assembler has no TR mnemonic — see the header of
// ebcdic_s390x.s), so no assembler validated those six bytes.
//
// ebcdic_diff_s390x_test.go asserts that the TR path and the loop path produce
// byte-identical output. That differential test is the substitute for the
// assembler check the TR path does not get, and it is the direct instrument for
// hypothesis H002's claim C2 (instruction-collapse fidelity).
//
// These are deliberately not exported on amd64: on amd64 the byte loop IS the
// implementation, so there would be nothing to compare against.

//go:noescape
func AtoELoop(dst, src []byte)

//go:noescape
func EtoALoop(dst, src []byte)

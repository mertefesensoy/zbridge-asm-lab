// s390x implementation of AtoE and EtoA.
//
// This file replaces the former UNDEF stub. Read the header before changing
// anything: it contains a hand-encoded instruction, and the reason it is
// hand-encoded is a finding, not an accident.
//
// ---------------------------------------------------------------------------
// 1. Go's s390x assembler cannot emit TR.
// ---------------------------------------------------------------------------
//
// The stub this file replaces said the amd64 lookup loop "collapses into a
// single hardware instruction: TR (Translate), opcode DC". The instruction is
// real, but Go's assembler does not know it: cmd/internal/obj/s390x/anames.go
// (go1.26.3) lists 729 mnemonics and none of them is TR, TRT, TRE, TROO or EX.
// MVC, CLC, XC, MVCIN, EXRL and SYSCALL are all present; the translate family
// is simply absent.
//
// So TR is emitted here as literal bytes. That is not a hack invented for this
// repo — it is attested practice inside the Go distribution itself:
//
//   golang.org/x/sys/unix/asm_zos_s390x.s  encodes SVC 08 as BYTE $0x0A; BYTE $0x08
//   internal/bytealg/indexbyte_s390x.s     encodes SRST  as WORD $0xB25E0082
//
// TR is SS-a format, six bytes:
//
//        byte 0   byte 1     bytes 2-3        bytes 4-5
//       +--------+--------+----------------+----------------+
//       |  0xDC  |  L     | B1 (4) D1 (12) | B2 (4) D2 (12) |
//       +--------+--------+----------------+----------------+
//
//   L  is the operand length MINUS ONE, so 0xFF means 256 bytes.
//   B1/D1 is the first operand: the buffer to translate, IN PLACE.
//   B2/D2 is the second operand: the base of the 256-byte table.
//
// With R2 holding the buffer and R3 holding the table, both at displacement 0:
//
//   TR 0(256,R2),0(R3)   ->   DC FF 20 00 30 00
//   TR 0(1,R2),0(R3)     ->   DC 00 20 00 30 00
//
// TR does not set the condition code.
//
// ---------------------------------------------------------------------------
// 2. TR translates in place, so a copy comes first.
// ---------------------------------------------------------------------------
//
// The Go signature is AtoE(dst, src []byte) — two distinct buffers. TR has one
// storage operand that is both source and destination. The idiomatic S/390
// answer, and the one an assembler programmer would write, is MVC then TR:
//
//   MVC   dst <- src      (move up to 256 bytes)
//   TR    dst in place    (translate through the table)
//
// This is a real cost the roadmap's "one instruction replaces the whole loop"
// framing hides: it is two SS-format instructions per 256-byte block, not one.
// Both are still block operations, so the per-byte loop is genuinely gone.
//
// ---------------------------------------------------------------------------
// 3. Runtime lengths need EXRL.
// ---------------------------------------------------------------------------
//
// MVC and TR encode their length in the instruction text, so a length known
// only at run time cannot be assembled directly. EXRL (Execute Relative Long)
// solves this: it executes one instruction located elsewhere, after ORing the
// low 8 bits of a register into byte 1 of that instruction — which for SS-a
// format is exactly the length field. The register therefore holds len-1.
//
// This is the pattern the Go standard library uses for precisely the same
// problem, and H002's threat #1 says to prefer stdlib-attested encodings:
//
//   runtime/memmove_s390x.s          EXRL $memmove_exrl_mvc<>(SB), R5
//   internal/bytealg/equal_s390x.s   EXRL $memeqbodyclc<>(SB), R8
//   internal/bytealg/compare_s390x.s EXRL $cmpbodyclc<>(SB), R7
//
// The EXRL target blocks below are marked NOFRAME deliberately. Without it the
// assembler would emit a prologue and EXRL would execute the prologue rather
// than the intended instruction.
//
// ---------------------------------------------------------------------------
// 4. There is a second, deliberately naive implementation in this file.
// ---------------------------------------------------------------------------
//
// AtoELoop and EtoALoop are byte-at-a-time reference implementations, declared
// in ebcdic_s390x.go and exercised by ebcdic_diff_s390x_test.go, which asserts
// that the TR path and the loop path produce byte-identical output. They exist
// because the TR path contains hand-encoded machine code that no assembler
// checked: the differential test is what stands in for the assembler.
//
// Register allocation (avoiding R0, which means "no base" in an address;
// R10/R11 REGTMP; R12 REGCTXT; R13 g; R14 link; R15 stack pointer):
//
//   R2  destination pointer  (TR first operand, MVC destination)
//   R3  translation table    (TR second operand)
//   R4  source pointer       (MVC source)
//   R5  bytes remaining
//   R6  scratch: len-1 for EXRL, and the source byte in the loop path
//   R7  scratch: the translated byte in the loop path

#include "textflag.h"

// func AtoE(dst, src []byte)
//
// Argument layout (FP): dst_base+0, dst_len+8, dst_cap+16,
//                       src_base+24, src_len+32, src_cap+40.
TEXT ·AtoE(SB), NOSPLIT|NOFRAME, $0-48
	MOVD	dst_base+0(FP), R2
	MOVD	src_base+24(FP), R4
	MOVD	src_len+32(FP), R5
	MOVD	$·atoeTable(SB), R3
	BR	trbody<>(SB)

// func EtoA(dst, src []byte)
TEXT ·EtoA(SB), NOSPLIT|NOFRAME, $0-48
	MOVD	dst_base+0(FP), R2
	MOVD	src_base+24(FP), R4
	MOVD	src_len+32(FP), R5
	MOVD	$·etoaTable(SB), R3
	BR	trbody<>(SB)

// trbody translates R5 bytes at R4 into R2 using the table at R3.
//
// Full 256-byte blocks are handled by directly assembled MVC and hand-encoded
// TR with the length field maxed out. The remainder, 1..255 bytes, is handled
// by one EXRL-driven MVC and one EXRL-driven TR.
//
// Not a Go function: entered by BR from AtoE/EtoA, so R14 still holds their
// caller's return address and the RET here returns all the way out.
TEXT trbody<>(SB), NOSPLIT|NOFRAME, $0-0
	CMPBEQ	R5, $0, done		// empty slice: nothing to do

blocks:
	CMP	R5, $256
	BLT	tail			// fewer than 256 bytes left

	MVC	$256, 0(R4), 0(R2)	// dst[0:256] = src[0:256]

	// TR 0(256,R2),0(R3) -- translate the block just copied, in place.
	BYTE	$0xDC
	BYTE	$0xFF			// length-1 = 255, i.e. 256 bytes
	BYTE	$0x20			// B1 = R2, D1 high nibble = 0
	BYTE	$0x00			// D1 low byte = 0
	BYTE	$0x30			// B2 = R3, D2 high nibble = 0
	BYTE	$0x00			// D2 low byte = 0

	LA	256(R2), R2
	LA	256(R4), R4
	SUB	$256, R5
	CMPBNE	R5, $0, blocks
	RET

tail:
	SUB	$1, R5, R6		// EXRL ORs R6's low byte into the
					// length field, which encodes len-1
	EXRL	$ebcdic_exrl_mvc<>(SB), R6
	EXRL	$ebcdic_exrl_tr<>(SB), R6

done:
	RET

// DO NOT CALL - target for EXRL. NOFRAME is required: a prologue here would
// be the instruction EXRL executes.
TEXT ebcdic_exrl_mvc<>(SB), NOSPLIT|NOFRAME, $0-0
	MVC	$1, 0(R4), 0(R2)	// length byte is patched by EXRL
	MOVD	R0, 0(R0)		// guard: fault rather than fall through
	RET

// DO NOT CALL - target for EXRL. The hand-encoded TR with its length field
// zeroed, ready for EXRL to OR len-1 into it.
TEXT ebcdic_exrl_tr<>(SB), NOSPLIT|NOFRAME, $0-0
	// TR 0(1,R2),0(R3)
	BYTE	$0xDC
	BYTE	$0x00			// length-1 = 0; EXRL overwrites this
	BYTE	$0x20			// B1 = R2, D1 = 0
	BYTE	$0x00
	BYTE	$0x30			// B2 = R3, D2 = 0
	BYTE	$0x00
	MOVD	R0, 0(R0)		// guard: fault rather than fall through
	RET

// ---------------------------------------------------------------------------
// Reference implementations: the same translation, byte at a time.
// ---------------------------------------------------------------------------

// func AtoELoop(dst, src []byte)
TEXT ·AtoELoop(SB), NOSPLIT|NOFRAME, $0-48
	MOVD	dst_base+0(FP), R2
	MOVD	src_base+24(FP), R4
	MOVD	src_len+32(FP), R5
	MOVD	$·atoeTable(SB), R3
	BR	loopbody<>(SB)

// func EtoALoop(dst, src []byte)
TEXT ·EtoALoop(SB), NOSPLIT|NOFRAME, $0-48
	MOVD	dst_base+0(FP), R2
	MOVD	src_base+24(FP), R4
	MOVD	src_len+32(FP), R5
	MOVD	$·etoaTable(SB), R3
	BR	loopbody<>(SB)

// loopbody is the direct transliteration of the amd64 loop: load a byte, index
// the table with it, store the result. Same registers as trbody.
TEXT loopbody<>(SB), NOSPLIT|NOFRAME, $0-0
	CMPBEQ	R5, $0, loopdone

loop:
	MOVBZ	0(R4), R6		// R6 = src[i]
	MOVBZ	0(R3)(R6*1), R7		// R7 = table[src[i]]
	MOVB	R7, 0(R2)		// dst[i] = R7
	LA	1(R4), R4
	LA	1(R2), R2
	SUB	$1, R5
	CMPBNE	R5, $0, loop

loopdone:
	RET

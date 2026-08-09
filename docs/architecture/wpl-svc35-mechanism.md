# The WPL / `SVC 35` mechanism, end to end

**Read [`README.md`](README.md) and [`zbridge-module.md`](zbridge-module.md) first** —
this document assumes you know what WTO and `SVC 35` are and have seen the `console`
package's shape. This document is the one place that goes all the way from a Go string
to a line on an operator's screen, one concrete step at a time, with real bytes.

A companion, diagram-first version of some of this material already exists at
[`../diagrams/wto-call-path.md`](../diagrams/wto-call-path.md) (Mermaid diagrams). This
document is the prose walkthrough; read whichever form suits you, or both.

---

## 0. The whole trip, in one table

| # | Step | Where in the code | Proven by |
|---|---|---|---|
| 1 | Validate the message | `console.validate` | pure logic, tested directly |
| 2 | Translate UTF-8 → EBCDIC IBM-1047 | `codepage.AtoE` | [rung E-L](evidence-ladder.md#6-the-other-track-e-l-and-how-u1-got-retired) (U1) |
| 3 | Build the WTO Parameter List (WPL) | `console.EncodeWPL` | [rungs E1 + E3](evidence-ladder.md) (U2) |
| 4 | Put the WPL somewhere `SVC 35` can address | `internal/storage.Malloc31` (not yet implemented) | nothing — U3, open |
| 5 | Point R1 at it, per MVS linkage | `internal/linkage`, future assembly | [rungs E2 + E3](evidence-ladder.md) |
| 6 | Issue `SVC 35` | `internal/svc.Call35` (declared, no body yet) | [rungs E2 + E3](evidence-ladder.md) |
| 7 | Read whatever comes back, map to a Go error | `zbridge.Error`, `HasCode` | **not retirable off-mainframe** — see §6 |

Steps 3, 5, and 6 are this document's real subject — they're where the actual byte
layout lives and where the project's most-cited evidence came from. Steps 1, 2, 4 are
covered briefly here and in depth in [`zbridge-module.md`](zbridge-module.md).

---

## 1. Step 1 — validation

Before anything else, `console.validate` rejects an empty message, a message longer
than `MaxTextLen` (124 bytes — see the note in §3.1 on where that number comes from),
and any message containing a control byte (`< 0x20` or `0x7F`). This is ordinary input
validation, and it's worth stating plainly *why* it happens first: an operator console
is a shared, human-facing workplace, not a log file nobody reads, so garbage input is
refused rather than translated and sent.

## 2. Step 2 — EBCDIC translation

The validated UTF-8/ASCII text is translated to **EBCDIC IBM-1047** by
`codepage.AtoE` — see [`zbridge-module.md`](zbridge-module.md) §2 for how that
translation itself is implemented (table-driven, `MVC`+`TR` on s390x). By the time
step 3 sees the text, every byte in it is already what the mainframe expects to read,
character-for-character.

## 3. Step 3 — building the WTO Parameter List (the WPL)

This is the heart of the whole project, and it's worth being explicit about why it was
ever in doubt: **the byte layout of the WPL is not written down in prose anywhere IBM
publishes.** Mainframe parameter lists like this one are built by an assembler *macro*
(conventionally `IEZWPL`/the `WTO` macro), and IBM documents what keywords the macro
accepts — not the raw bytes the macro expands into. The project's own roadmap
originally described the layout as "per IBM docs"; that phrase turned out to describe a
document that doesn't exist. `GC28-0683-2`, the natural place to look, was read page by
page and **does not contain a byte layout for this structure at all**
(`docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md`).

### 3.1 Where the real layout came from

An assembler can be asked to print the fully macro-expanded source it actually
assembled, not just the macro invocation you typed — a `PRINT ON,GEN,DATA` directive on
TK5's IFOX00 assembler. Assembling a single, minimal WTO macro call —
`WTO 'ZBRIDGE TEST HELLO',MF=L` — with that directive turned on produces this listing
(`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` §1):

```
  LOC  OBJECT CODE      STMT   SOURCE STATEMENT
000000                     7 WPLMIN   WTO   'ZBRIDGE TEST HELLO',MF=L
000000                     8+WPLMIN   DC    0F'0'
000000 0016                9+         DC    AL2(22)                 TEXT LENGTH
000002 0000               10+         DC    B'0000000000000000'     MCS FLAGS
000004 E9C2D9C9C4C7C540   11+         DC    C'ZBRIDGE TEST HELLO'   MESSAGE TEXT
00000C E3C5E2E340C8C5D3
000014 D3D6
```

Nobody had to guess. **The macro processor's own literal output is the primary
source** — this project's version of "read the source, not a summary of it."

### 3.2 The layout, byte by byte

```
 byte:  0     1     2     3     4                              len+4
       +-----------+-----------+--------------------------------+
       |  length   | MCS flags |     message text (EBCDIC)      |
       +-----------+-----------+--------------------------------+
        big-endian halfword,
        value = len(text) + 4        the whole thing is fullword-aligned
```

- **Bytes 0–1:** a big-endian 16-bit length field.
- **Bytes 2–3:** "MCS flags" (MCS = Multiple Console Support) — for the minimal,
  single-line, unrouted form used here, this is simply zero.
- **Byte 4 onward:** the message text, in EBCDIC, exactly as long as the original
  string.

**The correction that mattered most in this whole project** is right there in bytes
2–3. An earlier version of the project's own roadmap described this structure as *"a
2-byte length header followed by EBCDIC message text"* — no flags field at all. Had
that been implemented as written, **every byte from offset 2 onward would have been
wrong**: the flags field's two zero bytes would have been read as the first two
characters of the message, and everything after that would have been shifted. The
project would have discovered this on borrowed z/OS time, watching a garbled or
truncated message appear on a real operator's console. Finding it here, for free,
against a 1981 ancestor system, is the entire return on the emulation strategy
described in [`evidence-ladder.md`](evidence-ladder.md). See
[ADR 0004](../decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §2.1 for the
formal correction record.

### 3.3 Why the length is "text plus 4," not just "text"

The length field's value is **`len(text) + 4`**, and it's worth being able to defend
that arithmetic rather than just memorize it: **the field counts its own two bytes,
plus the two flag bytes, plus the text** — it's a length of *the whole header-plus-text
structure*, not a length of the text alone. This was confirmed at two different message
lengths in the same listing session, which is what turns "this could be a coincidence"
into "this is a rule": an 18-character message produced length field `0x0016` (22 =
18+4), and a 38-character message produced `0x002A` (42 = 38+4).

### 3.4 Alignment

The listing's very first line — `WPLMIN DC 0F'0'` — is a zero-length fullword-alignment
directive. It has no bytes of its own; it just forces whatever comes immediately after
it to start on a 4-byte boundary. That's a real constraint the encoder's caller has to
honor eventually (see §4 below and
[`zbridge-module.md`](zbridge-module.md) §5) — it just isn't something `EncodeWPL`
itself can enforce, because a plain `[]byte` carries no alignment guarantee in Go.

### 3.5 The routed form, and the trap hiding in it

The same assembler listing also showed what happens when you ask the `WTO` macro for
routing and descriptor codes (which console groups receive the message, and what kind
of message it is — see [`zbridge-module.md`](zbridge-module.md) §3.5):

```
000018 0016   DC AL2(22)                TEXT LENGTH      ← still 22!
00001A 8000   DC B'1000000000000000'    MCS FLAGS        ← bit 0 now set
00001C ...    DC C'ZBRIDGE TEST HELLO'  MESSAGE TEXT
00002E 0200   DC B'0000001000000000'    DESCRIPTOR CODES
000030 0020   DC B'0000000000100000'    ROUTING CODES
```

Two findings, and the second one is a genuine trap:

1. Setting the top bit of the MCS flags field (`0x8000`) signals that two more
   halfwords — descriptor codes, then routing codes — follow the message text.
2. **The length field still reads 22, even though the list is now 26 bytes long.** It
   never grows to cover the appended halfwords. An implementation that assumed "the
   length field always covers the whole parameter list" would work perfectly in testing
   with the minimal form and then be subtly, silently wrong the moment a caller asked
   for routing — which is exactly the scenario `zbridge/console` refuses outright
   rather than risk (see [`zbridge-module.md`](zbridge-module.md) §3.5): the remaining
   fifteen MCS flag bits have no citable specification (research brief 003 Q3), so
   `WithRoute`/`WithDescriptor` return `ErrLayoutUnverified` instead of emitting a list
   that might be subtly wrong in a way nobody could currently check.

### 3.6 The Go code, matched to the listing line by line

```go
const (
    wplLengthOff = 0  // the 2-byte big-endian length field
    wplFlagsOff  = 2  // the 2-byte MCS flags field
    wplTextOff   = 4  // where the EBCDIC text starts

    wplMinFlags = 0x0000 // minimal single-line WTO: all flag bits zero
)

func EncodeWPL(msg string, opts ...Option) ([]byte, error) {
    // ... validation, and refusal if routing/descriptor codes were requested ...

    buf := make([]byte, wplTextOff+len(msg))
    binary.BigEndian.PutUint16(buf[wplLengthOff:], uint16(len(msg)+wplTextOff)) // +4
    binary.BigEndian.PutUint16(buf[wplFlagsOff:], wplMinFlags)                  // 0x0000
    codepage.AtoE(buf[wplTextOff:], []byte(msg))                                // EBCDIC
    return buf, nil
}
```

Every constant here traces back to a specific line in the §3.1 listing — there is
deliberately no arithmetic in this function that isn't directly justified by something
the macro itself produced.

### 3.7 The proof that this Go function is right, not just plausible

`console/wpl_oracle_test.go` hard-codes the *exact* bytes the macro produced in §3.1 as
a constant (`macroOracle`) and asserts, byte-for-byte, that `EncodeWPL("ZBRIDGE TEST
HELLO")` produces exactly that sequence:

```go
const macroOracle = "0016" + "0000" +
    "E9C2D9C9C4C7C540" + "E3C5E2E340C8C5D3" + "D3D6"

func TestEncodeWPLMatchesIBMMacro(t *testing.T) {
    want := mustDecode(t, macroOracle)
    got, err := EncodeWPL("ZBRIDGE TEST HELLO")
    // ... assert got == want, byte for byte ...
}
```

This is worth pausing on, because it's a different *kind* of test than most codebases
have: **the expected value isn't a specification restated as a fixture — it's the
literal output of a real IBM macro, running on a real (if ancestor) system.** A second
test (`TestLengthFieldIsTextPlusFour`) checks the arithmetic rule itself across every
valid message length, not just the two lengths that happened to get assembled, so a
regression at, say, length 50 would be caught even though the macro was only ever run
at 18 and 38 characters.

---

## 4. Step 4 — storage below the bar

The encoded WPL is, at this point, an ordinary Go `[]byte` — living wherever Go's
garbage collector put it, with no guarantee about which address range that is. z/OS
services expect the parameter list to live in memory the service can actually reach,
which for `SVC 35` means **below the 2 GB address boundary** ("below the bar"), in
31-bit addressing mode. Getting the bytes from a Go slice into that kind of memory is
`internal/storage.Malloc31`'s job — and it's flatly not implemented yet, because
below-the-bar allocation is a pure z/OS behavior with no analogue QEMU or MVS 3.8j can
stand in for. Full detail: [`zbridge-module.md`](zbridge-module.md) §5. This is **U3**,
unretired, waiting on real z/OS access.

## 5. Step 5 — the register linkage

Standard MVS calling convention says: before issuing a supervisor call, put the address
of your parameter list in register **R1**. That sounds trivial and isn't, because Go's
own s390x calling convention has *already* assigned meanings to two of the registers
this handshake needs — R13 (Go: the goroutine pointer; MVS: a save-area address) and
R15 (Go: the stack pointer; MVS: entry point in, return code out). Any assembly routine
that sets up this call has to save and restore both, or Go's runtime loses track of
either the current goroutine or the stack. This collision — arguably the single most
dangerous fact in the whole codebase — is documented in full, with the exact register
table, in [`zbridge-module.md`](zbridge-module.md) §5.

**Rungs E2 and E3** (see [`evidence-ladder.md`](evidence-ladder.md)) both exercised this
handshake directly, in real MVS assembler — `LA 1,WPL` immediately before `SVC 35` — and
both worked. That verifies the *linkage convention itself* is understood correctly; it
does not verify the Go-assembly side of saving/restoring R13/R15 around it, because that
code hasn't been written yet (it waits on U3, same as step 4).

## 6. Step 6 — issuing `SVC 35`

This is where a fact that first showed up in an entirely different package
(`codepage`'s `TR` instruction — see [`zbridge-module.md`](zbridge-module.md) §2)
becomes unavoidable: **Go's s390x assembler has no `SVC` mnemonic.** This was verified
directly against the toolchain source, not assumed —
`cmd/internal/obj/s390x/anames.go` in go1.26.3 lists 729 recognized instruction names,
and `SVC`, along with `TR`, `TRT`, and `EX`, is simply not among them. (Go's `SYSCALL`
pseudo-instruction *does* assemble to `SVC 0` — but that's the Linux system-call trap,
architecturally unrelated to what an MVS/z/OS service needs.)

So `SVC 35` will be emitted the same way `TR` already was: as **literal instruction
bytes**, not a mnemonic:

```asm
BYTE $0x0A; BYTE $0x23     // SVC 35 — 0x23 is 35 in hex
```

This isn't an improvised workaround — it's the exact, attested technique other
IBM-authored Go code already uses for the same problem. `golang.org/x/sys/unix/asm_zos_s390x.s`,
part of the official Go distribution's vendored dependencies, encodes `SVC 08` as
`BYTE $0x0A; BYTE $0x08` for precisely the same reason. Two independent IBM-authored
sources (that file, and `ibmruntimes/go-recordio`, which hand-encodes `SAM31`/`SAM64`
the same way — see `docs/roadmap-errata.md` entry E4) now corroborate that this is the
accepted way to reach an un-mnemonic'd s390x instruction from Go assembly, not a
project-specific improvisation.

**Rungs E2 and E3 both issued a bare `SVC 35` from real MVS assembler**, and both
reached the console — proving the *instruction and its calling convention* are
understood, on the one system this project can actually test against. What hasn't
happened yet is writing and running the **Go-assembly** encoding of those same two
bytes; that step is unwritten pending U3 for exactly the same alignment/addressing
reasons as steps 4 and 5 (see `internal/svc.Call35`'s doc comment,
[`zbridge-module.md`](zbridge-module.md) §4).

## 7. Step 7 — reading whatever comes back

Here the project ran into a genuine, documented **divergence** between the ancestor
system and z/OS itself, rather than a simple unknown:

| | MVS 3.8j (tested) | z/OS (documented, not yet tested) |
|---|---|---|
| Return code for a single-line, non-MLWTO WTO | **None issued** | Yes, in R15 |
| R1 on return | 24-bit message identification number | message identification number |

The MVS 3.8j side comes from reading the primary manual directly: `GC28-0683-2` p.210
states plainly that no return code is issued for this exact call shape, and the
documented 00–14 return-code table only applies to a different, multi-line form
(MLWTO) this project doesn't use. The z/OS side is recorded in
[ADR 0004](../decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §2.2.

**The consequence for the evidence ladder:** rung E3, however convincing about steps 1
through 6, **cannot** verify step 7 — the ancestor system has nothing in its return
register for E3 to check. This is exactly why `zbridge.Error` carries a `HasCode bool`
field (see [`zbridge-module.md`](zbridge-module.md) §6) instead of assuming every call
produces a status the module can read: "the service returned nothing to map" had to be
a representable state, because on the one system this project could actually test
against, that's exactly what happens.

---

## 8. Summary: what's proven, and what's still a promise

| Step | Proven how far |
|---|---|
| 1. Validate | Fully — pure logic, no unknowns involved |
| 2. EBCDIC translate | Fully, on real s390x machine code (E-L, QEMU) |
| 3. Build the WPL | Fully, for the minimal form, against a real `SVC 35` (E1, E3) — routed form deliberately refused |
| 4. Below-the-bar storage | Not started — pure U3 |
| 5. Register linkage (concept) | Verified in real MVS assembler (E2, E3); the Go-assembly side is unwritten |
| 6. Issue `SVC 35` (concept) | Verified in real MVS assembler (E2, E3); the Go-assembly encoding is unwritten |
| 7. Read the result | **Cannot be verified off-mainframe** — documented divergence, not an open unknown |

Next: **[`emulation-harnesses.md`](emulation-harnesses.md)** — exactly what QEMU and
Hercules are actually emulating, and precisely where each one's evidentiary authority
runs out.

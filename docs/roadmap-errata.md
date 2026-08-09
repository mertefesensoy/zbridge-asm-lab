# Roadmap errata

**Read this alongside [`zbridge-asm-roadmap.pdf`](../zbridge-asm-roadmap.pdf).**

The roadmap is the project's mandate and is deliberately **never edited**. It is the
record of what was believed when it was written, and that record is part of what makes the
project's claims checkable. Where later evidence contradicts it, the contradiction is
recorded here and superseded by an ADR — never by rewriting the PDF.

This page exists for one practical reason: **nothing otherwise stops a fresh reader
implementing Phase 3b step 3 from the uncorrected text and building a parameter list that
is wrong from byte 2 onward.**

> If you are about to implement anything from the roadmap, check this page first.

---

## Corrections

### E1 · Phase 3b step 3 — the parameter list is missing a field

| | |
|---|---|
| **Roadmap** | p.6, Phase 3b step 3 |
| **As written** | "Construct the WTO parameter list (2-byte length header followed by EBCDIC message text, per IBM docs)." |
| **Superseded by** | [ADR 0004](decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §2.1 |
| **Evidence** | [`E1-E3-wto-layout-and-svc35-2026-07-26.md`](evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md) |
| **Severity** | **High — implementing as written produces a wrong parameter list** |

There is a **2-byte MCS flags field between the length and the text.** The layout is:

```
 byte:  0     1     2     3     4                              len+4
       +-----------+-----------+--------------------------------+
       |  length   | MCS flags |     message text (EBCDIC)      |
       +-----------+-----------+--------------------------------+
        big-endian, = len(text) + 4        fullword aligned (DC 0F'0')
```

The length field is `len(text) + 4` because it counts its own two bytes and the two flag
bytes as well as the text. Confirmed at two lengths: 18 chars → `0x0016` (22), 38 chars →
`0x002A` (42).

**Also:** "per IBM docs" is not available. GC28-0683-2 was read page by page and does not
document the layout at all ([`DOC-001-…`](evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md)).
The layout is carried only by the `IEZWPL` mapping macro, and no IBM publication prints the
generated expansion — `PRINT GEN` on the macro is the authoritative instrument. Corrected
wording: *"per the assembler expansion of the WTO macro."*

**Routed form, for completeness:** with `ROUTCDE`/`DESC`, MCS flag bit 0 (`0x8000`) is set
and two more halfwords are appended **after** the text — and the length field **still reads
22**. It never covers them.

---

### E2 · Phase 3b step 6 — the R15 return code is system-dependent

| | |
|---|---|
| **Roadmap** | p.6, Phase 3b step 6; and p.2, primitive table ("return code read from R15") |
| **As written** | "Read R15 for the return code and translate to a Go error." |
| **Superseded by** | [ADR 0004](decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §2.2 |
| **Evidence** | GC28-0683-2 p.210, read directly; research brief 003 Q4 |
| **Severity** | Medium — the step is valid on z/OS, invalid on the ancestor |

| | MVS 3.8j | z/OS |
|---|---|---|
| Return code for a single-line, non-MLWTO WTO | **None issued** | Yes, in R15 |
| R1 on return | 24-bit message identification number | message identification number |

**Consequence:** rung E3 **cannot** retire Phase 3b step 6 — the ancestor has nothing to
read. ADR 0001 §6 claimed it could, and that claim is corrected here.

Corrected wording: *"read the service's result — a return code in R15 on z/OS, nothing on
MVS 3.8j — and translate to a Go error where one exists."* This is why `zbridge.Error`
carries `HasCode bool`.

⚠️ The z/OS side of this table rests on brief 003 Q4, which cited the *Authorized Assembler
Services Reference* **without a form number and was not independently verified.** Confirming
it is an open question for the mentor.

---

### E3 · The unauthorized-WTO console prefix is `+`, not `@`

| | |
|---|---|
| **Roadmap** | p.1, "Why WTO is the right endgame target" |
| **As written** | "Unauthorized WTO messages carry an at-sign prefix on the console." |
| **Superseded by** | [ADR 0004](decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §2.3 |
| **Evidence** | [`E1-E3-…`](evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md) |
| **Severity** | Low — cosmetic detail; the roadmap's substance is confirmed |

```
FFFF 15.41.36 JOB    6  +ZBRIDGE TEST E2 RAW SVC 35 NO MACRO
```

A **plus sign**, on MVS 3.8j. The roadmap's actual point — that the message is
console-*prefixed*, not blocked, so the architectural payoff is full — **is confirmed by
this observation.** Only the character differs. z/OS's prefix is unverified.

---

### E4 · Phase 2's target file path does not exist

| | |
|---|---|
| **Roadmap** | p.2 ("From go-recordio's utils/utils.s") and p.5 (Phase 2) |
| **As written** | "reading `ibmruntimes/go-recordio/utils/utils.s` line by line" |
| **Verified** | 2026-07-27, against the repository tree via the GitHub API |
| **Severity** | Low — a path, but it would waste the first hour of Phase 2 |

**There is no `utils/utils.s`.** The v1 `utils/` directory contains only `utils.go`
(2,617 bytes) and no assembly at all. The file Phase 2 is about is:

```
v2/utils/utils.s          6,934 bytes, 154 lines
```

Also in `v2/utils/`, and relevant to Phase 2's deliverables:
`utils.go` (513 lines — `Malloc31` is here, at line 47, in Go rather than assembly),
`libvec_zos.go` (149,799 bytes — the Language Environment library vector),
`textflag.h` (vendored), and `README.md`.

**Two things this already confirms**, before annotation starts:

- **`SAM31` and `SAM64` are hand-encoded** — `BYTE $0x01; BYTE $0x0D` and
  `BYTE $0x01; BYTE $0x0E`. Go's s390x assembler has no mnemonic for them either.
- **`SVC 8` and `SVC 9` are hand-encoded as `BYTE $0x0A; …`** — so IBM's own module uses
  exactly the technique this project derived independently for `TR`, and which
  `SVC 35` (`BYTE $0x0A; BYTE $0x23`) will need. That is now attested in two independent
  IBM-authored sources (`x/sys/unix/asm_zos_s390x.s` and go-recordio).

The nine routines to annotate: `IefssreqX`, `Bpxcall`, `Svc8`, `Svc9`, `Call24`,
`Call31`, `Call64`, `Deref`, `Pc31`.

---

## Scope changes (not errors — decisions)

These are not corrections to mistakes; they are deliberate amendments recorded so the
roadmap and the plan do not silently diverge.

### S1 · Roadmap question 1 is closed — no cgo fallback

| | |
|---|---|
| **Roadmap** | p.8, question 1 |
| **As written** | "should the target be downscaled (WTO via `__console2()` through cgo first, then the SVC-direct version as a stretch)?" |
| **Resolved by** | [ADR 0004](decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §3, owner ruling 2026-07-27 |

**Closed. The direct `SVC 35` route is the deliverable; the cgo fallback is removed.** cgo
survives only as the Phase 3c `__console2()` **measurement baseline**, never in the shipped
module.

Reasoning: the cgo route's dependencies are a **strict superset** of the direct route's —
it needs IBM's Go fork and z/OS access just the same, *plus* a C compiler and the full
Language Environment. If either open risk materialises, the cgo version cannot be built or
run either. It insures against nothing that can actually happen, while costing the thesis
its central claim. The one risk it did cover — not being able to work out the parameter
list — is exactly what rungs E1–E3 retired.

### S2 · The Phase 1b benchmark table moves to Phase 3c

| | |
|---|---|
| **Roadmap** | p.5, Phase 1b deliverable |
| **As written** | "the ebcdic exercise gets a side-by-side comparison … with benchmark numbers attached" |
| **Resolved by** | [ADR 0004](decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §4 |

The timing table **moves to Phase 3c**, where hardware benchmarking is already planned. It
cannot be produced honestly under emulation: QEMU and Hercules implement `TR` as a software
loop, so timing it measures the emulator. This was pre-registered as an explicit non-claim
in [H002](hypotheses/002-s390x-port-equivalence.md) *before* any code ran.

**Shipping in its place now**, and needing no hardware — the instruction-count and encoding
comparison, which is the architectural claim the roadmap actually wanted:

| | amd64 | s390x |
|---|---|---|
| Per unit of work | 7 instructions **per byte** | 2 instructions **per 256 bytes** |
| Real encoding | — | `d2 ff 20 00 40 00` (`mvc`) · `dc ff 20 00 30 00` (`tr`) |

**Phase 1b is complete** on that basis, with the timing deliverable relocated rather than
dropped.

### S3 · Phase 1b's blocker no longer applies

| | |
|---|---|
| **Roadmap** | p.5, Phase 1b status and open question |
| **As written** | "Status: blocked on access." — "is LinuxONE Community Cloud sign-up the right path…?" |

**Not blocked.** `GOOS=linux GOARCH=s390x` cross-compiles on the laptop and `qemu-s390x`
runs the binaries; 29 tests pass on real big-endian s390x
([`E-L-…`](evidence/E-L-s390x-port-qemu-2026-07-25.md)). LinuxONE access is still *wanted*
— it is the only way to produce the S2 timing table — but it no longer blocks anything.

### S4 · The phase windows and the solo time budget are superseded

| | |
|---|---|
| **Roadmap** | p.3, the phase table's *Window* column; p.7, "Sustained time budget" |
| **As written** | Phase 2 "mid July – early August", Phase 3 "mid August – late September", Phase 4 "Q4 2026 and beyond"; and "5-8 hours per week" for one person |
| **Resolved by** | [ADR 0005](decisions/0005-team-scale-up-and-academic-year-recalibration.md) §5, owner ruling 2026-07-30 |
| **Replaced by** | [`docs/roadmap-2026-27.md`](roadmap-2026-27.md) |

The project scaled from one person to **five** (lead + four contributors; Jürgen Holtz
continues as mentor), and the plan re-anchors to the **2026–27 academic year**: Phase 2 in the
pre-semester block, Phases 3a/3b in the fall, Phase 3c and Phase 4 in the spring.

**Nothing else about the phases changes.** Definitions, deliverables and the endgame are
untouched — ADR 0005 §2 says so explicitly. Only *who*, *in what order*, and *by when*.

The roadmap licenses this itself (p.8): *"the phase definitions are stable, the timeline is
provisional."*

**Not changed by ADR 0005, and worth stating because it is tempting:** the U1/U2/U3 unknown
decomposition in [ADR 0001](decisions/0001-emulation-strategy-hercules-two-track.md) §3 is
**not** re-split. The toolchain dependency and the access dependency are visibly different
things, but separating them formally would amend the structure every evidence file's
`speaks_to:` header points at, and that needs its own decision. ADR 0005 §6 records the
ordering fact — *you cannot run a binary on a system you have no compiler for* — without
amending the decomposition.

---

## Structures the roadmap does not contain

Not errata; listed so a reader is not confused by finding them in the repo but not the PDF.

- **The U1 / U2 / U3 unknown decomposition** — [ADR 0001](decisions/0001-emulation-strategy-hercules-two-track.md) §3.
- **The E-ladder (E0–E3)** — [ADR 0001](decisions/0001-emulation-strategy-hercules-two-track.md) §6. An
  off-mainframe ladder that runs before the roadmap's z/OS work and pre-validated four of
  the six Phase 3b steps.
- **The production module `zbridge/`** — [ADR 0003](decisions/0003-production-bridge-module-architecture.md).
  The roadmap names a deliverable module in Phase 3c; ADR 0003 fixes its architecture.
- **The five-seat team structure and the PR gate** — [`docs/team/charter.md`](team/charter.md),
  shape fixed by [ADR 0005](decisions/0005-team-scale-up-and-academic-year-recalibration.md) §3–§4.
  The roadmap was written for one person and has no notion of contributors or review.

---

## How to use this page

1. Before implementing from the roadmap, check the affected page number above.
2. Where an entry exists, the **ADR is authoritative** and the roadmap line is superseded.
3. Where no entry exists, **the roadmap governs** — including on scope and intent
   (`docs/goal-prompt.md` §1).
4. New contradictions get raised as an ADR and added here. They are never fixed by editing
   the PDF.

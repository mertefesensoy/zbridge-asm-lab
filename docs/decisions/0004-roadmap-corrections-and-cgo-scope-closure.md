# ADR 0004 · Roadmap corrections from E1–E3 evidence, and closure of the cgo-fallback question

- Status: **Accepted**
- Date: 2026-07-27
- Author: Mert Efe Şensoy
- Decided by: **Mert Efe Şensoy (owner)**, 2026-07-27, in session
- Supersedes: three specific statements in `zbridge-asm-roadmap.pdf` (§2 below), and
  closes the roadmap's open question 1 (§3). Amends the Phase 1b and Phase 2 deliverable
  definitions (§4, §5).
- Evidence: `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`,
  `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`,
  `docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md`

---

## 1. Context

`docs/goal-prompt.md` §1 states that where this project's doctrine and the roadmap appear
to conflict, **the roadmap wins on scope and intent, and the conflict must be raised
rather than resolved silently.** Rungs E1–E3 produced evidence that conflicts with three
roadmap statements and that answers one of its open questions.

This ADR raises all four, records the owner's rulings, and supersedes the affected lines.
**The roadmap PDF is deliberately left unmodified.** It is the mentor-facing mandate and
the record of what was believed when it was written; rewriting it would destroy the
evidence trail this project's credibility depends on. A reader-facing pointer lives in
`docs/roadmap-errata.md` (§6).

## 2. Three corrections

### 2.1 Phase 3b step 3 omits the MCS flags field — superseded

**The roadmap says** (p.6, Phase 3b step 3):

> Construct the WTO parameter list (2-byte length header followed by EBCDIC message text,
> per IBM docs).

**Corrected:** there is a **2-byte MCS flags field between the length and the text.** From
the IFOX00 expansion of `WTO 'ZBRIDGE TEST HELLO',MF=L` on MVS 3.8j:

```
000000                     8+WPLMIN   DC    0F'0'
000000 0016                9+         DC    AL2(22)                 TEXT LENGTH
000002 0000               10+         DC    B'0000000000000000'     MCS FLAGS
000004 E9C2D9C9C4C7C540   11+         DC    C'ZBRIDGE TEST HELLO'   MESSAGE TEXT
```

The layout is **length (2) · MCS flags (2) · EBCDIC text**, fullword-aligned, and the
length field is `len(text) + 4` because it counts its own two bytes and the two flag bytes
as well as the text. Confirmed at two message lengths (18 → `0x0016`, 38 → `0x002A`).

**Why this correction matters more than the others:** an implementation built from the
roadmap's wording would be wrong from byte 2 onward and would have been discovered on
borrowed z/OS time.

**A second defect in the same sentence:** *"per IBM docs"* is not available. GC28-0683-2
was retrieved and read page by page and **does not document the byte layout at all**
(`DOC-001-…`). Research brief 003 Q1/Q6 established that the layout is carried only by the
`IEZWPL` mapping macro, and that no IBM publication prints the generated expansion —
`PRINT GEN` is the authoritative instrument. The corrected wording is therefore *"per the
assembler expansion of the WTO macro"*, not *"per IBM docs"*.

**Also recorded:** the routed form appends descriptor and routing halfwords **after** the
text with MCS flag bit 0 (`0x8000`) set, and **the length field does not cover them** — it
still reads 22. `zbridge/console` refuses `WithRoute`/`WithDescriptor` because brief 003
Q3 found no citable table for the remaining fifteen MCS bits.

### 2.2 Phase 3b step 6 and the R15 return contract — superseded

**The roadmap says** (p.6, step 6): *"Read R15 for the return code and translate to a Go
error."* And (p.2, primitive table): SVC dispatch has *"return code read from R15."*

**Corrected:** this is **system-dependent, and it is a documented divergence.**

| | MVS 3.8j | z/OS |
|---|---|---|
| Return code for a single-line, non-MLWTO WTO | **None issued** | Yes, in R15 |
| R1 on return | 24-bit message identification number | message identification number |

Source for the MVS side: GC28-0683-2 p.210, read directly. For the z/OS side: research
brief 003 Q4, cited to *z/OS MVS Programming: Authorized Assembler Services Reference*
(SET-WTO) — **without a form number, and not independently verified.**

**Consequences:**

1. **Rung E3 cannot retire Phase 3b step 6.** ADR 0001 §6 claimed it could; the doubt
   recorded against that claim on 2026-07-25 is confirmed correct. Step 6 is the one
   implementation step the off-mainframe programme provably cannot validate.
2. Step 6's corrected wording is: *"read the service's result — a return code in R15 on
   z/OS, nothing on MVS 3.8j — and translate to a Go error where one exists."*
3. `zbridge.Error` carries `HasCode bool` so that "the service returned nothing to map" is
   representable rather than indistinguishable from "returned zero". That field predates
   brief 003 and is retained.

### 2.3 The unauthorized-WTO console prefix is `+`, not `@` — superseded

**The roadmap says** (p.1): *"Unauthorized WTO messages carry an at-sign prefix on the
console, which is a cosmetic limitation, not a functional block."*

**Corrected on the ancestor:** the observed prefix is a **plus sign**.

```
FFFF 15.41.36 JOB    6  +ZBRIDGE TEST E2 RAW SVC 35 NO MACRO
```

**The roadmap's substance is confirmed and is the important part:** the message is
console-*prefixed*, not blocked, so the architectural payoff is full. Only the character
differs, and only on MVS 3.8j — **z/OS's prefix for an unauthorized WTO is unverified.**

## 3. Decision: roadmap question 1 is closed — cgo is a baseline, not a fallback

**The roadmap asks** (p.8, Q1): *"Is this the right scope for a student-level project on
this timeline, or should the target be downscaled (WTO via `__console2()` through cgo
first, then the SVC-direct version as a stretch)?"*

**Ruling: closed. The cgo fallback is removed from the plan. The direct `SVC 35` route is
the deliverable.** cgo survives in exactly one role: the `__console2()` latency comparison
Phase 3c already specifies, as a **measurement baseline only**.

### The reasoning, because it is not the obvious one

A fallback is insurance only if it survives the failure it covers. It does not:

| Requirement | Direct route | cgo route |
|---|---|---|
| IBM's Go fork for z/OS | required | required |
| Access to a z/OS system | required | required |
| A C compiler on z/OS | not needed | **required** |
| Language Environment | `Malloc31` touchpoint only | **full dependency** |
| Parameter-list knowledge | required — **and obtained** | not needed |

**The cgo route's dependencies are a strict superset of the direct route's.** The two risks
still open are z/OS access and the Go fork — and if either materialises, the cgo version
cannot be built or run either. It insures against nothing that can actually happen, while
costing the thesis its central claim (roadmap p.1: *"No cgo. No Language Environment
dependency."*).

The one risk the hedge genuinely covered — *"we may not be able to work out the parameter
list"* — is precisely the risk E1–E3 retired.

### What is retained

Phase 3c's comparison against `__console2()` via cgo stands, reframed as a **benchmark
baseline**. "As fast as the C path, with no C, no LE, and a static binary" is a stronger
result than either half alone, and it requires cgo to exist only in a test harness — never
in the shipped module. The no-cgo constraint on `zbridge` itself is unchanged and absolute.

## 4. Decision: the Phase 1b benchmark table moves to Phase 3c

**The roadmap says** (p.5, Phase 1b deliverable): *"the ebcdic exercise gets a side-by-side
comparison: the amd64 lookup loop versus the s390x `TR` instruction, with benchmark
numbers attached."*

**Ruling: the timing table moves to Phase 3c**, where hardware benchmarking is already
planned. It cannot be produced honestly under emulation — QEMU and Hercules implement `TR`
as a software loop, so timing it measures the emulator (ADR 0001; H002's explicit
non-claim, pre-registered before any code ran).

**Replacing it now, and shippable without hardware: an instruction-count and encoding
comparison.** That is the architectural claim the roadmap actually wanted:

| | amd64 | s390x |
|---|---|---|
| Per unit of work | 7 instructions **per byte** | 2 instructions **per 256 bytes** |
| The instructions | `MOVBQZX` / `MOVB` / `MOVB` / 2× pointer bump / `DECQ` / `JNZ` | `MVC` then `TR` |
| Real encoding | — | `d2 ff 20 00 40 00` · `dc ff 20 00 30 00` |

**Phase 1b is therefore complete**, with the timing deliverable relocated rather than
dropped. Two honest caveats travel with it, both already recorded: `TR` translates **in
place**, so a two-buffer API costs `MVC` *then* `TR` — "one instruction replaces the loop"
is not literally true; and Go's s390x assembler **has no `TR` mnemonic**, so those six
bytes are hand-encoded and disassembly-verified.

## 5. Decision: Phase 2 runs as the roadmap specifies

**Ruling: the full publication-grade walkthrough first**, in the roadmap's order, not
narrowed to the two primitives that block Phase 3b.

Considered and rejected: annotating `Malloc31` and SAM31/SAM64 first to unblock Phase 3b
steps 1 and 4 sooner. The owner chose the roadmap's scope, and the reasoning holds — the
roadmap calls Phase 2 *"the most important pre-z/OS phase"*, its four deliverables form one
coherent artifact, and roadmap question 3 asks the mentor to **review** it. A document
assembled for review is worth more whole than in fragments, and the reviewer's time is
scarce.

All four Phase 2 deliverables stand as written: the instruction-by-instruction annotation,
the CVT → JESCT → SSREQ navigation diagram with byte offsets, the SAM31/SAM64 switching
documentation, and the identification of which patterns are SSREQ-specific versus
transferable to WTO.

## 6. Decision: an errata page, not an edited PDF

`docs/roadmap-errata.md` lists every superseded roadmap statement with its correction and
the ADR and evidence that supersede it, and is linked from the project's standing
rules and doctrine notes.

Considered and rejected: regenerating the roadmap PDF with corrections marked inline. The
PDF is the mentor-facing mandate and the historical record of what was believed in June;
editing it would blur that and is the owner's document to change, not this project's
automation. The errata is additive, cannot silently disagree with the ADR chain, and
answers the only real objection to leaving the PDF alone — that nothing otherwise stops a
fresh reader implementing step 3 without the MCS flags field.

## 7. What this decision does not claim

- **It does not claim the corrections apply to z/OS.** Every correction in §2 was observed
  on MVS 3.8j under emulation. The MCS flags field's presence is a structural fact
  corroborated for z/OS by brief 003 Q2's `+4` citation; the `+` prefix and the
  return-code behaviour are **ancestor observations** and z/OS may differ. H001 still owns
  that gap.
- **It does not claim the roadmap was careless.** All three corrected statements are
  reasonable readings of what was available in June, and one of them ("per IBM docs") was
  wrong only because the docs turn out not to exist.
- **It does not reduce the thesis scope.** Closing question 1 *removes a hedge*; it does
  not narrow the deliverable. The endgame is unchanged.
- **It does not claim the direct route will succeed.** Access and toolchain remain open,
  and §3 argues only that a cgo fallback cannot help with either.
- **It does not change Phase 3's dependency on Phase 2** or on z/OS access.

## 8. What would reopen or reverse this decision

1. **Jürgen rules the other way on question 1.** His position governs scope; bring it here.
   A specific institutional reason for a cgo deliverable — a demo that must run before
   access lands, a comparison IBM wants — would reopen §3 immediately.
2. **A z/OS publication that documents the WPL byte layout is produced.** Then §2.1's
   corrected citation changes from the macro expansion to that source, and the layout
   should be re-checked against it rather than assumed to match.
3. **z/OS is observed to issue no return code for a single-line WTO.** Then §2.2's
   divergence collapses, step 6 becomes a non-step on both systems, and `HasCode` becomes
   the normal case rather than the defensive one.
4. **The Phase 2 walkthrough proves too large to complete in one pass.** Then §5 is
   revisited in favour of the narrowed ordering that was rejected here — on schedule
   evidence, not preference.

## 9. Links

- `zbridge-asm-roadmap.pdf` — the mandate; pp.1, 2, 5, 6, 8 affected
- `docs/roadmap-errata.md` — the reader-facing pointer created by §6
- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — §6's step-6 claim is
  corrected by §2.2 here
- `docs/decisions/0003-production-bridge-module-architecture.md` — §4's layout status
- `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` — the source of §2
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — owns the ancestor-versus-descendant gap
- `docs/hypotheses/002-s390x-port-equivalence.md` — pre-registered the §4 non-claim
- `docs/mentor-briefings/2026-07-27-phase-checkpoint.md` — how §2 and §3 are reported

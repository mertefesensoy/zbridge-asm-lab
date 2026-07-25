# Synthesis of research returns 001 and 002, and the primary-source audit that followed

**Date:** 2026-07-25 against `docs/goal-prompt.md`
**Status:** shipped

---

## 1. Problem / Motivation

Two Gemini returns landed in `research/` on 2026-07-25, answering briefs 001 (the WTO
parameter list, byte for byte) and 002 (the prior-art census behind the thesis novelty
claim). Per the action-selection rule in `memory/MEMORY.md`, an unsynthesised research
return is the highest-priority work in the project — ahead of rungs, ahead of code.

Brief 001 was the one that mattered. `docs/research-briefs/README.md` records rung E2
as undesignable without it, and H001 registers the length-field semantics as the single
detail most likely to produce an S0C4 or a silently truncated console message on
borrowed machine time.

Both returns were fluent, fully tagged with `FOUND`/`INFERRED`, and carried confidence
ratings. Return 001 answered all seven questions and concluded "No divergence found for
the simple form" — the comfortable answer, and the one the project wanted.

Goal-prompt §4 names the exact failure mode this describes: *"producing something
plausible, fluent, and unverified, and presenting it with the same confidence as
something checked."* So the return's central citation was checked before anything was
built on it.

## 2. What Changed

| File | Change |
|---|---|
| `docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md` | **New.** Page-by-page read of GC28-0683-2 recording what the manual actually says about WTO, and the two places the return contradicts it. |
| `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` | Status `PRE-REGISTERED` → `IN PROGRESS`; new "Evidence log — Line 1" section with per-sub-claim status. Pre-registration text below it untouched. |
| `docs/hypotheses/README.md` | H001 index row updated to reflect the C3/C4 contradiction. |
| `docs/research-briefs/001-…-authoritative-layout.md` | Status → `ANSWERED (PARTIAL)`; acceptance-criteria audit added (3 of 5 criteria unmet). |
| `docs/research-briefs/002-prior-art-census-…md` | Status → `ANSWERED`; acceptance-criteria audit added (verdict accepted, two non-blocking gaps). |
| `docs/research-briefs/003-wto-wpl-layout-source-and-return-code-contract.md` | **New.** The gap-closer: finds the publication that actually documents the layout, the MCSFLAG bit map, and the return-code contract. |
| `docs/research-briefs/README.md` | Status index updated; brief 003 added. |
| `research/README.md` | Index table filled in, each row carrying its audit verdict so a reader meets the caveat before the content. |
| `docs/evidence/README.md` | Documents the provisional `DOC` prefix and flags it for owner review. |
| `memory/MEMORY.md` | Current state, open claims, and next unblocked action updated. |

No code was touched. No `.s` file, no `go.mod`, no exercise.

## 3. Implementation Approach

### The audit method

The return cited *OS/VS2 MVS Supervisor Services and Macro Instructions*, GC28-0683,
for its Q2 answer — the load-bearing one. Verification ran in four steps:

1. **Confirm the form number is real.** Search established GC28-0683 as correctly
   titled, with editions ‑1 (Rel 3.0), ‑2 (Rel 3.7, Apr 78) and ‑3 (Rel 3.7, Sep 83),
   held by bitsavers and the Internet Archive. The form number was genuine.
2. **Retrieve the document.** GC28-0683-2 fetched from bitsavers (9.7 MB). It is a
   scanned image with no text layer, so no text extraction was possible — the pages
   were read visually instead.
3. **Locate the section.** The table of contents gave `WTO — Write to Operator` at
   manual p. 208, `WTO (List Form)` at p. 211, and `Communicating with the System
   Operator` at p. 75. Manual-page → PDF-page offset is **not constant**: +24 in the
   macro reference, +12 in Part I, because TNL GN28-2914 inserted fractional pages
   (138.1, 174.1–174.3). The offset was solved locally at each target.
4. **Read every page where the claim could live** — pp. 75–77, 208–213.

The contract of this method: it can prove a sentence **is** in the manual, and it can
prove a sentence is **not in the sections where it would belong**. It cannot prove
absence from the whole 240-page book. The evidence file states that boundary.

### What the audit found

- **The manual does not document the WTO parameter list byte layout anywhere.** The
  `WTO (List Form)` section — the one that generates the parameter list — gives macro
  syntax only and refers the reader back to the standard form. The quoted sentence the
  return attributed to it does not appear.
- **Maximum message text is 124 characters** (pp. 208, 211), not the 126 the return
  claimed under a `[FOUND]` tag.
- **A single-line WTO issues no return code.** p. 210: *"No return codes are issued by
  the WTO service routine if the MLWTO feature is not used."* On return, R1 holds the
  24-bit message identification number (corroborated p. 77). The documented 00–14
  return-code table applies only to MLWTO — which H001 explicitly scopes out.
- **Confirmed, not everything failed:** R1 carries the parameter list address in
  (p. 212, Execute Form Example 2); text is EBCDIC with non-printables blanked
  (p. 208); the blank/`@`/`*`/`+` console prefixing in ADR 0001 evidence item 7
  (p. 75).

### Handling the returns themselves

`research/README.md` requires returns to stay verbatim, and that anything wrong with
one be noted **in the consuming document**. Both return files were left byte-identical.
Every correction lives in the consumers: the briefs, H001, the evidence file, and the
research index.

## 4. Mathematical / Statistical Details

Not applicable — this change is documentary. The one quantitative claim in play is the
parameter-list length arithmetic, and it is deliberately **not** settled here.

For the record, the working (uncited) assumption is header-inclusive:

```
length = 2 (length halfword) + 2 (MCS flags halfword) + n (EBCDIC text bytes)
       = n + 4
```

Self-consistency notes that make this *plausible but still uncited*: the return's own
abend threshold ("less than 5") is exactly what `n + 4` predicts for a 1-byte minimum
text, and its worked example `DC AL2(8)` for `C'TEST'` is 4 + 4. Independent vendor
documentation describes the same construction. **None of that is a primary citation**,
which is the entire point — brief 003 Q2 exists to obtain one, and the E1 assembler
listing remains the tie-break for the MVS side under H001's honesty clause.

## 5. Design Decisions

**Audit the citation rather than accept a well-formed return.** The return met the
formatting protocol completely — tags, confidence ratings, a side-by-side table, an
explicit "no divergence found". Every surface signal said *usable*. Goal-prompt §4.1
allows only cited, evidenced, or registered claims, and "cited" has to mean the source
supports the sentence, or the tagging protocol is theatre. Checking one form number
cost roughly fifteen minutes and changed two sub-claims.

**Extend the rung namespace with `DOC` rather than distort an existing category.**
Options considered: (a) file as an `E`-rung — rejected, it executes nothing and the
provenance header exists precisely to stop that kind of blurring; (b) bury it in H001 —
rejected, H001 must preserve pre-registration text and a 25-page source check needs to
be citable on its own; (c) put it in `research/` — rejected, that directory is Gemini's
verbatim returns and mixing Claude-authored analysis in erodes the role split that
makes the returns auditable. `DOC` is marked provisional and flagged for owner review.

**Do not touch ADR 0001.** The C4 finding puts a specific ADR 0001 §6 table entry in
doubt (Phase 3b step 6, *"Read R15, map to a Go error"*, currently marked ✅ retired by
E3). Goal-prompt §5 boundary 4 reserves ADR reversal to the owner. The evidence is
brought and cross-linked from H001, the evidence file, and brief 003; the ADR is
unedited.

**Record the decision-rule defect instead of repairing it.** H001's pre-registered rule
names outcomes for C2 failing and for C1-or-C3 failing, but not for **C4 alone**. The
gap is real and was found only because evidence arrived. Editing a decision rule after
seeing evidence is what pre-registration exists to prevent, so the gap is documented,
a resolution is *proposed*, and adoption is left to a dated amendment made before E1
runs.

**Do not write rung E2's assembler yet.** Brief 001 was commissioned specifically to
unblock E2's parameter list, and it is tempting to treat the corroborated `n + 4`
reading as good enough to start drafting. It is not: the layout currently has no
primary citation on either system, and E2's whole purpose is to prove the WPL is
understood rather than copied. Building it on a vendor page would produce a rung that
looks passed and proves nothing. E2 waits for brief 003 Q1/Q2 or for the E1 listing.

## 6. Verification

What was actually run, versus what was reasoned about:

**Run:**
- Form-number verification for GC28-0683 (web search, 2026-07-25) — confirmed genuine.
- Retrieval of GC28-0683-2 from bitsavers — 9.7 MB, no text layer.
- Visual read of PDF pages 3–10, 86–90, 99–102, 217–226, 232–237, covering manual pages
  iii–x, 74–78, 87–90, 193–202 and **208–213**. Findings above are from those pages.

**Not run, and not claimed:**
- No search of the full 240-page manual. Absence is asserted only for the WTO sections.
- No Release 3.8 edition retrieved. All findings are **Release 3.7**; TK5 runs 3.8j.
- No z/OS publication consulted. Every z/OS statement in the returns remains uncited.
- No machine executed anything. No rung advanced; E0 remains unstarted.

**To reproduce the audit:**

1. Fetch `https://bitsavers.trailing-edge.com/pdf/ibm/370/OS_VS2/Release_3.7_1976/GC28-0683-2_OS_VS2_MVS_Supervisor_Services_and_Macros_Rel_3.7_Apr78.pdf`
2. Open manual pages 208–213 (PDF 232–237) and 75–77 (PDF 87–89).
3. Confirm: `msg: Up to 124 characters` (p. 208); the MLWTO return-code note (p. 210);
   `WTO MF=(E,(1))` with a pre-built list in register 1 (p. 212).
4. Confirm the absence: no byte offsets or length-field description in `WTO (List
   Form)` (p. 211).

## 7. Related Docs

- `docs/goal-prompt.md` — §4.1 (no plausible facts), §4.2 (verification is run), §5
  boundary 4 (ADR reversal is the owner's)
- `docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md` — the audit
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — Line 1 evidence log
- `docs/research-briefs/003-wto-wpl-layout-source-and-return-code-contract.md` — gap-closer
- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — §6, the table entry
  now in doubt
- `research/001-…`, `research/002-…` — the returns, verbatim

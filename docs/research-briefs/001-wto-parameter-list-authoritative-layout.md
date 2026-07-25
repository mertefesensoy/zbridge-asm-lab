# Research Brief 001 · The WTO parameter list, byte for byte: MVS 3.8 versus current z/OS

- Status: **ANSWERED (PARTIAL)** — returned 2026-07-25, audited 2026-07-25
- Date: 2026-07-25
- Requested by: Claude (architecture role)
- Consumed by: `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — **Line 1
  (documentary)**
- Priority: **highest open item.** Rung E2 cannot be designed without it.
- Return: `research/001-wto-parameter-list-authoritative-layout.md`
- Audit: `docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md`
- Follow-on: `docs/research-briefs/003-wto-wpl-layout-source-and-return-code-contract.md`

> ## Acceptance-criteria audit (2026-07-25)
>
> Scored against §"Acceptance criteria" below, which was written before the return
> arrived. **The brief is not closed.**
>
> | # | Criterion | Met? |
> |---|---|---|
> | 1 | Q1–Q4 answered for both systems, each with a form-numbered IBM citation | **No.** Q1, Q3, Q4 carry no citation at all. Q2's z/OS side names a manual without form number or release. Only Q2's MVS side carries a form number — and see 2. |
> | 2 | Q2 answered with a verbatim quotation of the manual's own words, per system | **No.** GC28-0683-2 was retrieved and read directly (pp. 75–77, 208–213). **The manual does not document the parameter list layout at all, and the quoted sentence does not appear in it.** |
> | 3 | Every claim tagged `FOUND`/`INFERRED` with confidence | **Yes.** Consistently done. |
> | 4 | Side-by-side table, divergence called out explicitly | **Yes.** Table present; "no divergence found" stated explicitly. |
> | 5 | Unfound answers stated plainly rather than filled in | **No.** Q3's bit map, Q4's full return-code semantics, and Q7's out-of-bounds behaviour are absent but not declared absent. Q3 and Q5 are `INFERRED` reconstructions presented in place of the requested documented material. |
>
> **Two `[FOUND]` claims are contradicted by the primary source:** maximum message
> length is **124**, not 126; and on MVS Release 3.7 a **single-line WTO issues no
> return code at all** — R1 returns the message identification number, and the
> documented 00–14 return-code table applies only to MLWTO.
>
> The return is not discarded — its Q1 shape and Q6 "no divergence found" remain
> useful, and per `research/README.md` the return file itself stays verbatim. But
> **nothing in it may be cited as primary** until brief 003 supplies real citations.

## Why this brief exists

The project's endgame is `WTO(message string) error` implemented in pure Go assembly,
issuing `SVC 35` with a hand-built parameter list. Before that list can be built, its
exact byte layout must be known — and known *authoritatively*, not approximately.

A first research pass on 2026-07-25 established the **shape** of the list from
multiple secondary sources: a 2-byte length, a 2-byte MCS flag field, then EBCDIC
message text. That much is consistent everywhere.

**What could not be established from any secondary source is what the length field
actually counts.** Sources are silent, or vague, or mutually inconsistent on whether
it includes the 4-byte header and whether it includes trailing routing/descriptor
code areas. That single ambiguity is the difference between a working call, a
truncated console message, and an S0C4.

The project cannot afford to discover this on borrowed mainframe time, and it cannot
publish a walkthrough built on a guess.

## The questions

Answer for **both** systems, separately and explicitly:

- **System A — MVS 3.8** (the TK5 guest; *OS/VS2 MVS* era documentation)
- **System B — current z/OS** (the deployment target)

### Q1 — Simple-form layout

For the simplest single-line WTO with no routing codes, no descriptor codes, and no
extended options: what is the complete parameter list, field by field, with **byte
offsets and field widths**?

### Q2 — The length field (the critical one)

At offset 0 there is a halfword length. **Exactly which bytes does it count?**

- Does it include its own 2 bytes?
- Does it include the MCS flags halfword?
- Does it include trailing routing/descriptor code areas when present?
- Is there a documented maximum, and what happens on overflow?

Quote the manual's own wording verbatim. This field is the whole reason for the brief.

### Q3 — The MCS flags halfword

What is each bit at offset 2? Which bits are *required* to be set for a plain
problem-state message, and which must be zero? Specifically: is there a bit whose
zero value on MVS 3.8 has acquired a *defined meaning* on z/OS — for example the bit
indicating an extended WPL (WPX) exists? A bit that is "reserved, set to zero" on one
system and load-bearing on the other is exactly the kind of divergence that would
break the H001 oracle relationship.

### Q4 — Invocation contract

On each system: which register carries the parameter list address into `SVC 35`?
Which register carries the return code out? What is the documented meaning of each
return code value — most importantly, is 0 the only success value?

### Q5 — The macro expansion

What assembler source does the `WTO` macro generate for the simplest case on each
system? If the *OS/VS2 MVS* documentation prints an example expansion, reproduce it.
This is directly comparable to the listing rung E1 will capture from IFOX00 on TK5,
which makes it the highest-value single item after Q2.

### Q6 — Documented divergences

Does any IBM migration or compatibility documentation state changes to the WTO
parameter list between MVS 3.8 and current z/OS? Compatibility statements, migration
guides, and "changes to this macro" sections are the target. **An explicit IBM
statement of compatibility, or of a specific incompatibility, is worth more than
every other answer in this brief combined.**

### Q7 — Addressing-mode constraints

Where must the parameter list reside on each system? MVS 3.8 is 24-bit (below the
16 MB line); z/OS is expected to want below the 2 GB bar in AMODE 31. Confirm both
from documentation, and state what the documented behaviour is if the list is above
the applicable boundary.

## Required sources

**Primary — these are what count:**

- *OS/VS2 MVS Supervisor Services and Macro Instructions* (System A). Locate the
  correct form number and cite it.
- *z/OS MVS Programming: Assembler Services Reference* (System B), `WTO` entry —
  current release. Note the release the citation is from.
- *z/OS MVS Programming: Authorized Assembler Services Reference* if the unauthorized
  volume is incomplete on the parameter list.
- IBM's own macro-changes or migration documentation for Q6.

Bitsavers and the CBT Tape archive are legitimate homes for the historical manuals;
cite the manual by form number, and note the archive as the access route.

**Corroboration only — never the primary citation:** Jay Moseley's Hercules
materials, SimoTime, the H390-MVS list archives, IBM Mainframe Forum, blogs. These
are genuinely useful for locating a manual or sanity-checking a reading, and several
are written by people who know this material deeply. They still do not settle a
byte-layout question for a thesis.

## Explicit non-goals

Do not spend effort on: WTOR (roadmap Phase 4, separate brief later), multi-line WTO,
extended WPL/WPX details beyond the single flag bit in Q3, console routing
configuration, or the `__console2()` C runtime path.

## Acceptance criteria

The brief is answerable and closed when:

1. Q1, Q2, Q3, Q4 are answered for **both** systems, each with a form-numbered IBM
   manual citation.
2. Q2 in particular is answered with a **verbatim quotation** of the manual's own
   words, for each system.
3. Every claim is tagged `FOUND` or `INFERRED` and carries a confidence rating.
4. A **side-by-side comparison table** of the two systems' simple-form layouts is
   included, with any divergence called out explicitly — and with "no divergence
   found" stated explicitly if that is the result.
5. Where an answer could not be found, that is stated plainly rather than filled in.

## How the result will be used

The return lands at `research/001-wto-parameter-list-authoritative-layout.md`.
Claude then folds it into H001 as Line 1 evidence and designs rung E2's hand-built
parameter list against it. **If Line 1 and the E1 assembler listing disagree, the
listing wins for the MVS side** (H001's tie-break clause) — and that disagreement is
itself recorded as a finding, because it would mean a published manual is wrong or
ambiguous.

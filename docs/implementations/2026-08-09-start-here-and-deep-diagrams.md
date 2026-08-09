# `START_HERE.md` and deep sequence/state/object diagrams

**Date:** 2026-08-09
**Author:** Claude Sonnet 5, in session with Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

After issue #1's five phases shipped (architecture docs, C4 diagrams, testing docs,
`RUN.md`, v1.0.0), the owner reported that the documentation, while thorough, was not
enough for *them personally* to present the project confidently to their mentor — a
CTO-level reviewer with far more mainframe experience than the owner. The stated need
was explicit: a single guided entry point assuming zero prior mainframe knowledge, with
every IBM/mainframe term defined ("no guessing"), plus additional diagram types beyond
C4's four levels — sequence, state, and object diagrams — with deep, jargon-defining
explanations. This was requested under real time pressure ahead of a mentor
presentation.

## 2. What Changed

| File | Change |
|---|---|
| `START_HERE.md` | New, at the repository root. Zero-assumed-knowledge master guide: the whole project story in plain English, a complete grouped glossary of every mainframe/IBM/Go term used anywhere in this repository, a guided reading path through every other document with a one-line takeaway for each, a detailed walkthrough of the WPL byte-layout story (the likely deepest technical question), a demo-narration guide, and a one-page cheat sheet. |
| `docs/architecture/c4/sequence-state-object-diagrams.md` | New. Three diagrams beyond the four C4 levels: a sequence diagram of the live demo across all five systems it touches (Mermaid), a state diagram of every condition the emulated mainframe can be in including the failure mode that loses work (Mermaid), and an object diagram showing the actual byte/field values from one real captured run (hand-built SVG). Each has extensive prose explaining every participant/state/field in plain language. |
| `docs/architecture/c4/object-diagram-e3-run.svg` | New. The object diagram as a standalone SVG asset, built via the visualization widget and geometry-verified against the rendered DOM. |
| `docs/architecture/c4/README.md` | Added a pointer to the new diagram document, placed right after the level table so a reader sees it before diving into Level 1. |
| `README.md` | Added a one-line banner at the top pointing to `START_HERE.md` and `RUN.md`, so a first-time visitor to the repository (including the mentor) finds the guided entry point immediately. |

No source or test files changed.

## 3. Implementation Approach

**Calibrated the assumed audience precisely, rather than over- or under-explaining.**
The owner is a computer engineering student who already understands general software
engineering concepts (sequence/state diagrams, Go, assembly) — their stated gap was
specifically IBM/mainframe domain vocabulary, not diagramming notation or programming
fundamentals. `START_HERE.md` and the new diagram document therefore explain every
mainframe-specific term (SVC, WTO, JCL, JES2, EBCDIC, macro, IPL, DASD, AMODE, and
about twenty more) but do not re-explain what a sequence diagram or a Go function is.

**Picked diagram content that is genuinely different from the existing four C4
levels, not a restatement of them.** The sequence diagram covers the *cross-system
operational flow* of the live demo (laptop → harness script → mainframe job queue →
the `SVC 35` instruction → console) — a different axis from C4 Level 4's *intra-Go
call chain*, which stops at the platform boundary. The state diagram covers the
*emulated mainframe's own lifecycle*, including the specific failure mode
(`docs/evidence/E0-tk5-boot-2026-07-26.md`'s three unclean-shutdown incidents)
that no existing document renders visually. The object diagram is the only place in
the repository that shows one concrete, fully-instantiated example with real captured
values (length=37, flags=0x0000, the actual message text, the actual job name and
return codes) rather than the general byte-layout *shape* already covered in
`wpl-svc35-mechanism.md`.

**Used Mermaid for sequence/state, raw SVG for object — matching each diagram type to
the tool that renders it correctly**, rather than forcing one mechanism for all three.
Mermaid has first-class `sequenceDiagram` and `stateDiagram-v2` syntax that handles
lifeline/state layout automatically and correctly; hand-coding that in raw SVG would
have meant manually replicating what Mermaid already does well. This also matches
existing project precedent — `docs/diagrams/wto-call-path.md` already uses Mermaid
fenced blocks for diagrams-as-code. An object diagram (boxes of instance/attribute
text with association arrows) has no equivalent first-class Mermaid diagram type, so
it was hand-built as SVG using the same structural-diagram technique and verification
method as the four C4 levels (`mcp__visualize__show_widget`, then a standalone file
with hardcoded colors, then geometry checked via `getBBox()` against the actual
rendered DOM — no text/box overlaps, nothing outside the safe viewBox area).

**Every number in the object diagram and in `START_HERE.md` §4 traces to a specific
already-captured evidence file, not to a fresh claim.** The message text, length field
(37), flags (0x0000), job name (`ZBE3GO`), return codes (all `0000`), and console
prefix (`+`) are copied from `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`'s
rung E3 section, read again during this session rather than recalled.

## 4. Mathematical / Statistical Details

One arithmetic fact is restated for the third time across this repository's
documentation (also in `wpl-svc35-mechanism.md` and `testing.md`): the WPL length
field equals the message's character count plus 4. `START_HERE.md` §4.5 calls this out
explicitly as "the one number to have ready," because it is concrete, checkable, and
was judged the single most likely follow-up a technically sophisticated reviewer would
probe (e.g., "why plus 4, specifically").

## 5. Design Decisions

- **`START_HERE.md` at the repository root, ahead of even `README.md` in practice**
  (linked from the top of `README.md`), rather than folded into
  `docs/architecture/README.md`. The existing architecture `README.md` already
  assumes the reader is comfortable navigating a technical documentation folder; this
  owner's stated need was closer to "prep material to internalize before a live,
  high-stakes conversation," which called for a single, self-contained, narrative
  document rather than another entry in a folder of peer documents.
- **Grouped glossary, not alphabetical.** Considered an A-to-Z glossary. Rejected
  because terms build on each other conceptually (you need "SVC" before "SVC 35"
  makes sense, "macro" before "WTO macro expansion" does) — grouping by topic lets a
  reader absorb the vocabulary in the order the rest of the documentation will use it,
  rather than encountering "WPL" before "WTO" alphabetically.
- **An explicit "if asked, say this plainly" instruction for the one open technical
  question (the return-code behavior on the real target system).** Considered writing
  around the gap or minimizing it. Rejected — this project's own doctrine
  (`docs/goal-prompt.md` §4.1, "no plausible facts") and its precedent (ADR 0002's
  withdrawal) both treat precise honesty about what's unproven as a strength in front
  of an expert reviewer, not a weakness to be smoothed over, and `START_HERE.md` says
  so directly rather than leaving the owner to discover that stance under pressure
  mid-conversation.
- **No content from the mentor's actual private email reply was used anywhere in
  this document**, consistent with the standing, still-open decision from earlier in
  this session (issue #1, Phase 0) that filing his correspondence into this public
  repository is a separate, not-yet-made owner decision. The "if your mentor asks"
  section is grounded entirely in this project's own already-public ADRs, errata, and
  hypotheses — which happen to cover the same technical ground, without attributing
  anything to a private message.

## 6. Verification

- The object diagram SVG was opened directly (`file://`) and checked via
  `getBBox()` on every `<rect>` and `<text>`, the same method used for the four C4
  diagrams: zero pairwise text overlaps, nothing outside the safe viewBox area
  (max rendered extent 540×670 inside a 680×720 canvas).
- Every glossary entry and every fact in `START_HERE.md` §4 (the WPL story) was
  checked against the specific source file it summarizes — `docs/architecture/README.md`,
  `wpl-svc35-mechanism.md`, `zbridge-module.md`, and the E0/E1-E3 evidence files — read
  again in this session rather than recalled from earlier in the conversation, given
  the owner's explicit "no guessing" requirement.
- The reading-path table in `START_HERE.md` §3 was checked link-by-link against the
  actual files present in the repository at the time of writing.

## 7. Related Docs

- [`START_HERE.md`](../../START_HERE.md) — the new master guide.
- [`docs/architecture/c4/sequence-state-object-diagrams.md`](../architecture/c4/sequence-state-object-diagrams.md) —
  the new diagram document.
- [`docs/architecture/c4/README.md`](../architecture/c4/README.md) — updated to point
  at it.
- [`docs/architecture/README.md`](../architecture/README.md),
  [`zbridge-module.md`](../architecture/zbridge-module.md),
  [`wpl-svc35-mechanism.md`](../architecture/wpl-svc35-mechanism.md),
  [`testing.md`](../architecture/testing.md) — the documents `START_HERE.md` sequences
  and summarizes.
- [`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`](../evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md) —
  source of every real value used in the object diagram.

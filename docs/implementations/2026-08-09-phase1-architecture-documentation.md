# Phase 1 architecture documentation: `docs/architecture/`

**Date:** 2026-08-09
**Author:** Claude Sonnet 5, in session with Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

The project's existing documentation (ADRs, evidence files, the roadmap) is precise and
well-disciplined, but every individual document assumes the reader already has context
— what WTO and `SVC 35` mean, what U1/U2/U3 are, what an operator console is — because
each one was written to record a specific decision or result, not to onboard a new
reader from zero. There was no single place that explained, from first principles, how
the system actually works.

That gap became concrete with two developments: the project scaled to a five-person
team (ADR 0005), whose four new contributors need to get oriented without an
onboarding phase (`docs/team/charter.md` explicitly rules one out), and the owner asked
for a documentation pass detailed enough to explain the system "like explaining to a
freshman in college." This is Phase 1 of the day-plan recorded in
[GitHub issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1).

## 2. What Changed

| File | Change |
|---|---|
| `docs/architecture/README.md` | New. Orientation doc: background on mainframes/z/OS/WTO for readers with zero context, why no cgo, the U1/U2/U3 framework, a map of the rest of the folder. |
| `docs/architecture/zbridge-module.md` | New. Package-by-package tour of the `zbridge/` production module — contract, invariants, and why several packages are deliberately incomplete. |
| `docs/architecture/evidence-ladder.md` | New. Walkthrough of rungs E-L and E0→E3: what was actually run, on what, and exactly what each result does and does not prove. |
| `docs/architecture/wpl-svc35-mechanism.md` | New. End-to-end, byte-level walkthrough of the WTO Parameter List and `SVC 35`, from a Go string to a console line. |
| `docs/architecture/emulation-harnesses.md` | New. What QEMU and Hercules/TK5 actually emulate, what each can and cannot prove, and why two different tools are used. |

No source or test files changed.

## 3. Implementation Approach

**Synthesis, not new research.** Every fact in these five documents traces to an
existing ADR, evidence file, or the actual source code in `zbridge/` — all read
directly during this session (not recalled from a prior summary) to keep citations
accurate. No new claims are introduced; this is organization and explanation of
claims the repository already makes and has already verified.

**Structured as progressive disclosure, not one large document.** `README.md` assumes
zero context and ends with a reading order; each subsequent document assumes the
previous ones and goes narrower and deeper. A reader who only needs the module tour
isn't required to first read the byte-level mechanism, but the byte-level document can
assume the module tour's vocabulary. Documents cross-link rather than duplicate —
`wpl-svc35-mechanism.md` doesn't restate the register-collision table from
`zbridge-module.md`, for instance; it points at it.

## 4. Mathematical / Statistical Details

One small arithmetic rule appears and is explained in plain English in
`wpl-svc35-mechanism.md` §3.3: the WTO Parameter List's length field equals
`len(message text) + 4` — the field counts its own two bytes, the two MCS-flags bytes,
and the text. This is stated as a rule (confirmed at two independent message lengths
in the underlying evidence, and checked by `TestLengthFieldIsTextPlusFour` across every
valid length in `console/wpl_oracle_test.go`), not derived from first principles —
there's no independent mathematical justification for "+4" beyond "that's what the two
header fields cost in bytes."

## 5. Design Decisions

- **Five focused documents, not one long one.** Considered a single
  `docs/architecture/README.md` covering everything. Rejected because the "detailed and
  self-explanatory" requirement benefits from letting a reader stop once they have what
  they need — someone who only wants the module tour shouldn't have to scroll past the
  byte-level mechanism to find it.
- **Cite and synthesize, never restate as a second source of truth.** Considered
  copying key facts (e.g. the WPL byte layout) directly into these documents as
  standalone assertions. Rejected in favor of always pointing at the ADR/evidence file
  that actually established the fact, matching the project's existing discipline
  (`docs/roadmap-errata.md`'s own pattern of correcting via pointer, never via a second,
  independently-drifting copy) — so a future correction to, say, the WPL layout only
  ever has to happen in one place.
- **Mentor correspondence deliberately excluded.** Per the Phase 0 decision recorded in
  issue #1 ("fix stale status only"), nothing from the mentor's actual email reply is
  filed anywhere in the repository yet. These five documents draw exclusively from
  already-committed repository content and directly-read source code — not from the
  mentor thread discussed in this session.

## 6. Verification

- Every code excerpt quoted in these documents was copied verbatim from the actual
  source files in `zbridge/` during this session (not reconstructed from memory) —
  `zbridge.go`, `errors.go`, `console/*.go`, `codepage/codepage.go`,
  `internal/svc/svc.go`, `internal/storage/storage.go`, `internal/linkage/linkage.go`,
  `subsys/subsys.go`, `dataset/dataset.go`.
- Every numeric/factual claim (byte offsets, instruction encodings, test counts,
  toolchain version numbers, evidence dates) traces to a specific cited ADR, evidence
  file, or source file — checked against those files directly, not against a prior
  summary of them.
- Cross-links between the five documents were checked manually: each document's
  "Next:" pointer matches an actual, existing file at the path given.
- `docs/architecture/c4/README.md`, referenced from `README.md` §7 and
  `emulation-harnesses.md` §6 as the next document in the reading order, does **not**
  exist yet — it's Phase 2 of issue #1, not yet started. That forward reference is
  intentional (it states the reading order this folder is converging on) but is a
  currently-broken link until Phase 2 lands.

## 7. Related Docs

- [GitHub issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1) — the
  day-plan this is Phase 1 of.
- [`docs/architecture/README.md`](../architecture/README.md),
  [`zbridge-module.md`](../architecture/zbridge-module.md),
  [`evidence-ladder.md`](../architecture/evidence-ladder.md),
  [`wpl-svc35-mechanism.md`](../architecture/wpl-svc35-mechanism.md),
  [`emulation-harnesses.md`](../architecture/emulation-harnesses.md) — the five new
  documents.
- [ADR 0001](../decisions/0001-emulation-strategy-hercules-two-track.md),
  [ADR 0003](../decisions/0003-production-bridge-module-architecture.md),
  [ADR 0004](../decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) — the
  decisions these documents explain.
- `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`,
  `docs/evidence/E0-tk5-boot-2026-07-26.md`,
  `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` — the primary evidence these
  documents synthesize.
- [`docs/diagrams/wto-call-path.md`](../diagrams/wto-call-path.md) — the existing
  Mermaid-diagram companion to `wpl-svc35-mechanism.md`.

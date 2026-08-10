# `RUN.md` §7 expanded: pre-flight, fallback plan, and landing the talk

**Date:** 2026-08-10
**Author:** Claude Sonnet 5, in session with Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

Following `START_HERE.md` and the sequence/state/object diagrams, the owner asked for
a dedicated demo-day file (pre-flight checklist, fallback plan, closing talking
points) beyond what `RUN.md` §7 already had. On review, the owner correctly pointed
out that a separate file would duplicate `RUN.md` rather than add anything new — the
content belongs with the rest of the demo material. This change folds the missing
pieces directly into `RUN.md` §7 instead of creating a new file.

## 2. What Changed

| File | Change |
|---|---|
| `RUN.md` | §7 restructured into §7.1 (pre-flight checklist), §7.2 (timing budget), §7.3 (the existing 6-step script, unchanged in substance), §7.4 (new — what to do if something breaks live, with a specific fallback to the E1-E3 evidence file), §7.5 (new — a closing line and a bridge into `START_HERE.md` §4 for Q&A). |

No source or test files changed.

## 3. Implementation Approach

**Added exactly the three things `RUN.md` didn't already cover**, rather than
restating what §3–§6 already establish: a pre-flight checklist (nothing to run live,
just confirms the environment is ready), a live-failure fallback plan (point at
already-captured evidence instead of debugging in front of an audience), and a closing
line. The existing 6-step script (now §7.3) was left substantively unchanged — it
already worked, so it was preserved and just given a fixed home in the new
subsection structure.

## 4. Mathematical / Statistical Details

None.

## 5. Design Decisions

- **No separate `DEMO.md`.** Considered and reversed during this session, following
  the owner's own observation that a second file would just duplicate `RUN.md`. One
  document, one place to look, is a better fit for something meant to be glanced at
  under presentation pressure.
- **The fallback plan points at a specific, already-existing evidence file** rather
  than suggesting a screenshot or recording be prepared separately, because
  `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` already contains the exact
  console line, byte values, and return codes from a real captured run — nothing new
  needed to be created for this to work as a fallback.

## 6. Verification

- Re-read the edited section after writing to confirm the existing 6-step script's
  content and cross-references (to §3.1, §6.1–§6.4) were preserved unchanged inside
  the new §7.3.
- Section numbering checked for consistency with the rest of the document (§8
  "Troubleshooting" and §9 "Where to go for more detail" still follow correctly after
  the expanded §7).

## 7. Related Docs

- [`RUN.md`](../../RUN.md) §7 — the changed section.
- [`docs/implementations/2026-08-09-phase4-run-md.md`](2026-08-09-phase4-run-md.md) —
  the original `RUN.md` implementation doc this extends.
- [`START_HERE.md`](../../START_HERE.md) §4 — where §7.5 bridges to for Q&A depth.

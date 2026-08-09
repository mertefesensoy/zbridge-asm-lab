# Phase 4: `RUN.md` — the demo-ready run guide

**Date:** 2026-08-09
**Author:** Claude Sonnet 5, in session with Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

Nothing in the repository previously collected, in one place, everything needed to go
from a clean machine to running the project's central demo (a message built by Go
accepted by a real, emulated `SVC 35`). The setup detail existed but was scattered
across `docs/runbooks/tk5-hercules-setup.md` (narrated, historical, written as things
were being discovered) and several evidence files. The owner asked for a root-level
`RUN.md`, explicit about every dependency and external binary, written so the demo can
be run live without improvising. This is Phase 4 of the day-plan in
[GitHub issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1) —
reordered ahead of the v1.0.0 release (originally Phase 4) specifically so the tagged
release contains it.

## 2. What Changed

| File | Change |
|---|---|
| `RUN.md` | New, at the repository root. Two-track run guide (laptop-only vs. full emulated demo), a complete dependency table with exact versions and a verifiable SHA-256, step-by-step commands for every gate in the repo, a suggested live demo script, and a troubleshooting table. |

No source or test files changed.

## 3. Implementation Approach

**Extracted and re-sequenced, rather than re-derived.** Every command and expected
output in `RUN.md` traces to a source already in the repository: the vet/test/cross-compile
commands from `CLAUDE.md`'s own "Build and test" section, the WSL2/QEMU procedure from
`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`'s reproduction section, and the TK5
setup and `mvsjob.sh` usage from `docs/runbooks/tk5-hercules-setup.md` §12 and the
`mvsjob.sh` script's own header comment (read in full during this session, not
paraphrased from memory). Nothing in `RUN.md` is a new claim — it is existing,
evidenced procedure, resequenced for a reader executing it live rather than a reader
reconstructing history.

**Structured around the demo, not around the repository's own module layout.** The
document opens with "the two tracks" (laptop-only vs. full emulated demo) and closes
with a numbered live-demo script (§7) that names exactly what to say at each step,
because the owner's stated use case was presenting this live — a reader optimizing for
"reproduce every gate in isolation" would have organized the document differently.

**The `gen-e3` command line was exercised, not assumed.** `zbridge/cmd/gen-e3/main.go`
was read in full to confirm its actual usage (`go run ./cmd/gen-e3 [message]`, message
optional, diagnostics on stderr so stdout stays pure JCL) before writing §6.2 — the demo
section explicitly invites picking a live message from the audience, which only works
if the command's argument handling is stated correctly.

## 4. Mathematical / Statistical Details

None — this document contains no algorithm or statistical claim of its own; it
transcribes commands and expected outputs already established elsewhere.

## 5. Design Decisions

- **Root-level `RUN.md`, not under `docs/`.** Matches the convention the owner named
  explicitly (their other project, T-MAC) and the general convention of `README.md` /
  `RUN.md` living at a repository's root as the first thing a visitor finds.
- **A dependency table with a verifiable hash, not just a download link.** Considered
  linking to the TK5 download page only. Rejected — the page can change, and a demo
  that silently downloads a different `mvs-tk5.zip` than the one every evidence file in
  this repository was captured against is a demo that might not reproduce. The SHA-256
  already recorded in `docs/evidence/E0-tk5-boot-2026-07-26.md` is repeated here as a
  runnable check (§2), not just a citation.
- **Both PowerShell and bash shown for every Windows-side step.** Matches this
  project's existing convention (`CLAUDE.md`'s own build/test section already does
  this) and reflects that the actual demo crosses the Windows/WSL2 boundary at least
  twice (cross-compiling on Windows, running under WSL2/Hercules) — a reader following
  along needs to know which shell they're in at each step, so every command block says
  so.
- **A troubleshooting table scoped to what a live demo can hit**, not a full
  reproduction of `docs/runbooks/tk5-hercules-setup.md`'s narrated incident history.
  That runbook remains the authoritative deep-dive (linked from `RUN.md` §9); this
  document's troubleshooting section is deliberately the shorter, symptom-first version
  suited to "something just broke mid-demo, what do I do."

## 6. Verification

- Every command in `RUN.md` was cross-checked against a source read during this
  session: `CLAUDE.md`'s build/test block, `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`'s
  reproduction section, `docs/runbooks/tk5-hercules-setup.md` §12, `docs/runbooks/mvsjob.sh`
  (read in full — the `up`/`run`/`down` usage, environment variables, and exact expected
  log lines in `RUN.md` §6 match the script's actual behavior), and
  `zbridge/cmd/gen-e3/main.go` (read in full for its actual CLI usage).
  Track A (§3) was **re-executed live** during this session (not merely transcribed):
  `go vet`/`go test` across all seven modules and the `linux/s390x` cross-build across
  all six eligible modules, both confirmed clean immediately before this document was
  written.
- Track B (§4–§7) could not be re-executed in this session (no WSL2/Hercules access
  from this environment) — its commands and expected outputs are transcribed from the
  cited evidence files' actual captured runs, not newly verified. This limitation is
  not hidden: anyone running Track B for the first time is the actual verification of
  this section, and `docs/evidence/E0-tk5-boot-2026-07-26.md` /
  `E1-E3-wto-layout-and-svc35-2026-07-26.md` remain the primary evidence it was
  transcribed from.
- The dependency table's version numbers (§2) are copied verbatim from the provenance
  headers of `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` and
  `docs/evidence/E0-tk5-boot-2026-07-26.md`, not estimated.

## 7. Related Docs

- [GitHub issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1) — the
  day-plan this is Phase 4 of (reordered ahead of the v1.0.0 release).
- [`RUN.md`](../../RUN.md) — the new document.
- [`docs/runbooks/tk5-hercules-setup.md`](../runbooks/tk5-hercules-setup.md),
  [`docs/runbooks/mvsjob.sh`](../runbooks/mvsjob.sh) — the narrated setup and the
  harness script `RUN.md` wraps.
- [`docs/evidence/E0-tk5-boot-2026-07-26.md`](../evidence/E0-tk5-boot-2026-07-26.md),
  [`E1-E3-wto-layout-and-svc35-2026-07-26.md`](../evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md),
  [`E-L-s390x-port-qemu-2026-07-25.md`](../evidence/E-L-s390x-port-qemu-2026-07-25.md) —
  the evidence every command and expected output in `RUN.md` traces back to.
- [`docs/architecture/testing.md`](../architecture/testing.md) — referenced throughout
  `RUN.md` for *why* each gate exists, not just how to run it.

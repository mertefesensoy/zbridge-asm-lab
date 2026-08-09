# Phase 3: testing documentation

**Date:** 2026-08-09
**Author:** Claude Sonnet 5, in session with Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

Phases 1 and 2 explain the system's architecture in prose and diagrams. Neither
explains how the test suite actually proves any of it — and this repository uses the
word "test" for several genuinely different things (an ordinary unit test, a byte-exact
comparison against a real IBM macro's output, an assembly-vs-Go differential check, and
a manually-run mainframe evidence rung are all called "tests" somewhere in this
project's own writing) without anywhere stating that they're different or why. The
owner explicitly asked for this distinction to be documented. This is Phase 3 of the
day-plan in [GitHub issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1).

## 2. What Changed

| File | Change |
|---|---|
| `docs/architecture/testing.md` | New. Explains all seven categories of "test" in this repository — unit, oracle, differential/parity, integration, the cross-compile gate, benchmarks, and evidence rungs — with a real code example for each and current verified pass counts. |
| `docs/architecture/README.md` | Added `testing.md` to the §7 reading-order list (previously said "forthcoming"). |

No source or test files changed.

## 3. Implementation Approach

**Taxonomy first, then one worked example per category**, rather than an exhaustive
file-by-file listing. Skimmed every `*_test.go` file in the repository to find the
category boundaries, then picked the single clearest real example of each (the oracle
test in `console/wpl_oracle_test.go`, the differential test pattern shared by
`ebcdic_diff_s390x_test.go` and `codepage_diff_s390x_test.go`, the real-syscall
integration tests in `syscall-linux`, the honest instruction-count benchmark in
`ebcdic/README.md`) so the document teaches the pattern rather than restating every
test function.

**Verified the test counts directly rather than citing an older document.** Ran
`go test ./... -v` and counted only top-level `--- PASS`/`--- FAIL` lines (excluding
subtests) across `zbridge` and the five Windows-buildable lab modules, closing the
"reconcile exact counts" item flagged during this session's Phase 0 validation. The
`zbridge` count (27) matches `docs/roadmap-2026-27.md` exactly; the lab-module total
(18) differs by one from that document's recorded 19 — recorded as a minor, expected
drift over roughly ten days of development rather than investigated further, since both
counts agree on the fact that actually matters (zero failures).

## 4. Mathematical / Statistical Details

None — no algorithm or statistical method is introduced. The one numeric table (§8 of
the new document) is a directly-observed count from a test run, not a computed or
estimated figure.

## 5. Design Decisions

- **Seven categories, not the conventional two ("unit" and "integration").** Considered
  using standard testing vocabulary throughout. Rejected because it would flatten real,
  load-bearing distinctions this project's own doctrine depends on — an "oracle test"
  (bytes checked against what a real machine produced) and a "differential test"
  (two independent implementations checked against each other) are both loosely
  "unit tests" by conventional definitions, but they answer different questions and
  trace to different hypotheses (H001 vs. H002 respectively); collapsing them would
  have made the document less useful to a reader trying to understand *why* a given
  test exists.
- **"Evidence rungs" documented as a test category at all, despite not running through
  `go test`.** Considered scoping this document to `go test`-driven tests only, with
  evidence rungs left entirely to `evidence-ladder.md`. Rejected because the owner's
  request was framed as "how do the tests work" broadly, and a reader who doesn't
  understand that E0–E3 are evidentiary in a different, heavier way than a `go test`
  run would be missing the single most distinctive thing about how this project
  establishes truth.
- **Benchmarks kept as their own section, separate from correctness tests.** They test
  nothing about correctness and are governed by a completely different rule (never
  quote a number from an emulator) — folding them into "unit tests" would have buried
  that rule.

## 6. Verification

- `go test ./... -v` re-run directly during this session for `zbridge` and all five
  Windows-buildable lab modules; counts in the new document's §8 are taken from that
  run's actual output, not copied from an earlier memory or doc.
- Every code excerpt in the new document was copied verbatim from the actual test files
  read during this session — `zbridge/errors_test.go`, `zbridge/console/console_test.go`,
  `zbridge/console/wpl_oracle_test.go`, `ebcdic/ebcdic_diff_s390x_test.go`,
  `zbridge/codepage/codepage_diff_s390x_test.go`, `syscall-linux/syscall_linux_test.go`,
  `syscall-linux/syscall_nonlinux_test.go`, `ebcdic/ebcdic_bench_test.go`.
- Cross-references to `H001`/`H002` and `evidence-ladder.md` were checked against the
  actual hypothesis files (read in full during this session), not against a summary of
  them.

## 7. Related Docs

- [GitHub issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1) — the
  day-plan this is Phase 3 of.
- [`docs/architecture/testing.md`](../architecture/testing.md) — the new document.
- [`docs/hypotheses/001-mvs38j-svc35-wto-oracle.md`](../hypotheses/001-mvs38j-svc35-wto-oracle.md),
  [`002-s390x-port-equivalence.md`](../hypotheses/002-s390x-port-equivalence.md) — the
  pre-registered claims the oracle and differential test categories make executable.
- [`docs/architecture/evidence-ladder.md`](../architecture/evidence-ladder.md) — full
  detail on the evidence-rung category.

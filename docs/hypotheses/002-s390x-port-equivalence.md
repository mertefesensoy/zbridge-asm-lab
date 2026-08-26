# Hypothesis 002 · The five `_s390x.s` bodies are behaviourally equivalent to their amd64 counterparts, and emulation can prove that equivalence — but emulation cannot measure the performance payoff the roadmap promises

- Status: **PARTIALLY RESOLVED — E-PORT-CLEAN on C1, C2, C4; C3 UNTESTED** (2026-07-25).
  See "Resolution" at the end of this file. Everything between here and that section
  is the original pre-registration text, preserved unchanged.
- Date: 2026-07-25
- Author: Mert Efe Şensoy
- Ladder rung: E-L (Track L / QEMU), feeding T2 on real hardware
- Builds on: ADR 0001 §5 (Track L time-boxed, QEMU as the Phase 1b inner loop);
  the roadmap's Phase 1b deliverable; `docs/mainframe-baseline-strategy.md` §4 (the
  T2 port order and its rationale); `docs/codex-handover.md` §4.5 (the honest-
  benchmark convention)

## Why this, and why now

Phase 1b — porting the five exercises to s390x — has been blocked on LinuxONE
Community Cloud access since late June. ADR 0001 unblocks it: `GOOS=linux
GOARCH=s390x` is a first-class Go port, and emulated s390x execution is available on
the laptop today. The port stops being blocked and becomes ordinary work.

But there is a trap in the roadmap's Phase 1b deliverable, and it needs to be named
before anyone runs a benchmark and believes it.

The roadmap promises, for `ebcdic`: *"the amd64 lookup loop versus the s390x `TR`
instruction, with benchmark numbers attached."* That is the headline result of Phase
1b and it is genuinely compelling — one instruction replacing an entire loop. **An
emulator cannot produce that number honestly.** QEMU's TCG and Hercules both
*implement* `TR` in software, as a loop. Timing `TR` under emulation measures the
emulator's translation loop, not the hardware operation that makes `TR` interesting.
Depending on the emulator's internals, the measured result could show no speedup, or
a slowdown, or an arbitrary speedup — and none of the three would say anything about
an IBM Z machine.

So this hypothesis splits what Phase 1b was treating as one deliverable into two,
with different evidence requirements and different completion dates. Correctness is
provable now. Performance is not provable until real hardware, and the repo's
existing honest-benchmark convention makes this a convention violation to fudge.

## What is being claimed, precisely

- **C1 · Correctness parity.** Each of the five `_s390x.s` implementations passes the
  identical test suite its `_amd64.s` counterpart passes, including the stdlib oracle
  comparisons (`bytes.Equal`, `bytes.Compare`, `binary.BigEndian`, `len`).
- **C2 · Instruction-collapse fidelity.** The z/Architecture single-instruction forms
  produce byte-identical output to the amd64 loops they replace: `TR` for `ebcdic`,
  `STH` + `MVC` for `strmanip`, `CLC` (via `EX`/`EXRL` or a strided loop) for
  `bytecmp`, and the R13/R14/R15 reads for `regs`.
- **C3 · Emulator independence.** Where both are available, QEMU s390x and Hercules-
  hosted Linux s390x agree on every test outcome. Disagreement means an emulator bug
  is in play and neither result is trustworthy until resolved.
- **C4 · ABI contract.** The Go frame-size/argument-size contract (`$frame-args`)
  holds on a big-endian 64-bit target, and `go vet` passes clean for
  `GOOS=linux GOARCH=s390x` across all five modules.

## What is explicitly NOT being claimed

- **No performance claim from emulated execution.** Any benchmark run under QEMU or
  Hercules measures the emulator. Such numbers may be recorded in `docs/evidence/`
  for regression-detection purposes only, and **must** carry the provenance header
  from ADR 0001 §7 plus an explicit "emulated — not a hardware measurement" line.
  They may not appear in the README benchmark tables, the thesis, or any published
  artifact as s390x performance figures.
- **The roadmap's `ebcdic` side-by-side benchmark table is therefore deferred to real
  hardware** (LinuxONE, or z/OS USS at T2). This is a genuine reduction in what Phase
  1b can deliver off-mainframe, and it is recorded here rather than quietly dropped.
  Everything else in the Phase 1b deliverable — five ported exercises, tests passing,
  the x86-to-s390x mapping notes in each README — completes under emulation.

## The instrument

1. **Cross-compile gate (laptop, no emulator).** `GOOS=linux GOARCH=s390x go vet` and
   `go build` for all five modules. Catches syntax and frame-contract errors before
   any emulator is involved. This is already named in
   `docs/mainframe-baseline-strategy.md` §5.1 as pre-access preparation.
2. **QEMU execution (inner loop).** Build test binaries with `go test -c` for
   `linux/s390x` and run them under QEMU s390x emulation. Seconds per cycle.
3. **Hercules Linux s390x (time-boxed, C3 only).** Within ADR 0001's 8-hour box, run
   the same test binaries to check emulator agreement. If the box expires, C3 is
   recorded as untested and C1/C2/C4 stand on QEMU alone.

Port order follows `docs/mainframe-baseline-strategy.md` §4 unchanged: `ebcdic` →
`strmanip` → `regs` → `bytecmp` → `syscall`. That order was chosen so the strongest
test oracle comes first, and nothing about emulation changes the reasoning.

## Pre-registered decision rule

**E-PORT-CLEAN** — C1 ∧ C2 ∧ C4 hold (C3 holds or is untested within the box).
→ Phase 1b's correctness deliverable is complete. The `_s390x.s` bodies replace the
`UNDEF` stubs. Rung T2 on real hardware is re-scoped from "write and debug s390x
assembly on borrowed time" to "confirm on hardware and capture the benchmark table
that emulation could not produce." T2's machine-time cost drops substantially.

**E-PORT-PARTIAL** — some modules pass, others fail.
→ Passing modules ship; failing ones keep their `UNDEF` stubs and a documented
failure note. The stub convention (`docs/codex-handover.md` §4.3) exists exactly for
this: a stub that fails loudly is correct behaviour for unfinished work. No module
ships a silently-wrong implementation to make the table look complete.

**E-PORT-EMULATOR-SUSPECT** — C3 fails; QEMU and Hercules disagree.
→ Both results are quarantined for the affected module. Diagnose which emulator is
wrong (the Go standard library's own s390x tests are the tiebreaker: if they fail
under one emulator, that emulator is the problem, not our assembly). No correctness
claim for that module until resolved or until real hardware settles it.

## Threats to validity

1. **Emulator-tolerated undefined behaviour.** An emulator may accept an instruction
   encoding or alignment that real hardware rejects. `EX`/`EXRL` in `bytecmp` is the
   most likely site. Mitigation: prefer the encodings the Go standard library's own
   s390x assembly uses (reading those files is already scheduled as Phase 2 work);
   flag any construct not attested there for hardware confirmation at T2.
2. **Big-endian bugs that the tests do not reach.** `strmanip`'s big-endian length
   header is the one place where amd64 testing could have hidden an
   endianness assumption, because on amd64 the byte order was constructed
   explicitly. Mitigation: `strmanip`'s tests must assert on exact byte sequences,
   not on round-trip behaviour alone — a round-trip is endianness-blind.
3. **QEMU version drift.** Results are only reproducible against a pinned emulator
   version. Provenance headers record the exact QEMU/Hercules version.
4. **The `syscall-linux` exercise is Linux-specific by construction.** Its s390x form
   validates the `SVC` *trap shape* on Linux, not the z/OS `SVC 35` path. It sits in
   U1, not U2; H001 owns U2. Do not let a passing `syscall-linux` on s390x be read as
   evidence about WTO.

## What would falsify

- Any module's s390x tests produce different results from its amd64 tests on inputs
  where both are defined → C1 fails for that module.
- `TR`, `MVC`/`STH`, or `CLC` output differs byte-for-byte from the amd64 loop → C2
  fails.
- `go vet` cannot verify the `$frame-args` contract on s390x, or the declared frames
  are wrong → C4 fails.
- QEMU and Hercules disagree on any test → C3 fails, quarantine applies.

## Links

- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` §5 (Track L
  time-box, QEMU inner loop), §7 (provenance headers)
- `docs/mainframe-baseline-strategy.md` §4 (port order), §5.1 (cross-compile prep)
- `docs/codex-handover.md` §4.3 (stub convention), §4.5 (honest benchmarks)
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — the U2 sibling; this hypothesis
  is U1 and the two must not be conflated

---

# Resolution — 2026-07-25

Everything above this line is the pre-registration text as written before any
evidence existed, and is unchanged.

**Evidence:** `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`
(QEMU 10.2.1 user-mode, Go 1.26.3, `GOOS=linux GOARCH=s390x`, all five modules).

**Outcome: E-PORT-CLEAN** — the decision rule's condition is "C1 ∧ C2 ∧ C4 hold
(C3 holds or is untested within the box)", and that is what happened. C3 is untested:
Hercules 3.13 is installed but no Hercules-hosted Linux s390x guest exists yet, so
QEMU has no counterparty to be compared against.

| Claim | Verdict | Basis |
|---|---|---|
| **C1** correctness parity | **HOLDS**, with one recorded deviation for `regs` — see below | All five modules' full test suites pass under QEMU, including the stdlib oracle comparisons in `bytecmp` and `strmanip` |
| **C2** instruction-collapse fidelity | **HOLDS** | `TR`, `MVC`+`MOVH`, and `CLC` produce byte-identical output to the amd64 loops. For `ebcdic` this is checked three ways: differential test against a byte-loop implementation, direct comparison against the Go table for all 256 inputs, and GNU objdump disassembly of the hand-encoded bytes |
| **C3** emulator independence | **UNTESTED** | QEMU only. Permitted by the decision rule; C3 remains open |
| **C4** ABI contract | **HOLDS** | `GOOS=linux GOARCH=s390x go vet` and `go build` clean across all five modules; `go vet` is what mechanically checks `$frame-args` |

## The C1 deviation, and why it is not a post-hoc amendment

C1 as written says each s390x implementation "passes the **identical** test suite its
`_amd64.s` counterpart passes." For `regs` that is not literally true, and the reason
is architectural rather than a shortfall: **s390x has no frame-pointer register**, so
`GetBP` cannot exist there. Go assigns R13 to `g`, R14 to the link register and R15
to the stack pointer, and nothing to a frame pointer.

The API was therefore split: `GetBP` is amd64-only; s390x declares `GetLinkRegister`
(R14) and `GetGReg` (R13), each with its own tests. `GetSP`, `GetFramedSP` and every
architecture-neutral test are shared and pass identically on both.

**The timing matters and is recorded deliberately.** This split was chosen by the
owner at the start of the session, *before* any s390x code was written and before any
test was run — not after seeing a failure. It is a scope decision made in advance,
not a rule loosened to reach green (goal-prompt §4.3). The alternative considered and
rejected was returning R13 from a function named `GetBP`, which would have preserved
the letter of C1 by mislabelling a register in a teaching repository.

The pre-registration defect this exposes is worth naming, because it is the same
class as H001's: **C1's wording assumed API-identical ports.** A future hypothesis of
this shape should say "passes an equivalent test suite, with any per-architecture API
divergence declared in advance" rather than "identical".

## Threats to validity, revisited against what actually happened

1. **Emulator-tolerated undefined behaviour** — this threat was correctly aimed. The
   mitigation was followed: `MVC`/`CLC`/`EXRL` are used in exactly the forms attested
   in `runtime/memmove_s390x.s`, `internal/bytealg/equal_s390x.s` and
   `compare_s390x.s`. **The `TR` path has no stdlib precedent** because Go's assembler
   has no `TR` mnemonic at all, so it is hand-encoded. It is flagged here for hardware
   confirmation at T2, exactly as the threat instructed. Disassembly agreement is
   strong evidence but is not execution on real hardware.
2. **Big-endian bugs the tests do not reach** — mitigated as specified.
   `strmanip`'s tests assert exact byte sequences (`TestWrapKnownFixture`,
   `TestWrapHighByte`), not round-trip behaviour, and `TestWrapMatchesReference`
   compares against `binary.BigEndian` across lengths 0–512.
3. **QEMU version drift** — the exact version is pinned in the evidence header.
4. **`syscall-linux` is Linux-specific by construction** — unchanged and still true.
   Its s390x form validates the trap shape on Linux and sits in U1. It is *not*
   evidence about WTO. One new observation belongs to U2's future, not to this
   hypothesis: Go's `SYSCALL` mnemonic assembles to `SVC 0`, and since the assembler
   has no `SVC` mnemonic, `SVC 35` will require hand-encoding (`BYTE $0x0A; BYTE
   $0x23`) by the same technique used for `TR`.

## What remains open

- **C3** until a Hercules-hosted Linux s390x guest runs the same binaries.
- **The performance half of Phase 1b** — deferred to real hardware, unchanged. No
  timing was taken under QEMU, deliberately.
- **Hardware confirmation of the hand-encoded `TR`** at rung T2.

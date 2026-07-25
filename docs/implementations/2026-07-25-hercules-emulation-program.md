# Hercules emulation program: two-track strategy, decision/hypothesis methodology, and the E-ladder

**Date:** 2026-07-25
**Status:** shipped (documentation and methodology; TK5 install pending owner execution)

---

## 1. Problem / Motivation

Two things were blocked. **Phase 1b** (port the five exercises to s390x) waited on
LinuxONE Community Cloud access. **Phase 3** (the WTO endgame) waited on z/OS access,
which the roadmap names as the project's single biggest risk. Meanwhile Phase 1 has
been complete since 2026-07-05 and the project had no forward motion available to it
that did not require hardware.

The owner proposed Hercules as a middle ground, and the mentor (Jürgen Holtz)
independently supplied the entry point after consulting a former colleague with
operational Hercules experience.

The obvious form of the proposal is unavailable: **z/OS cannot legally run under
Hercules.** The roadmap already says so, and the research pass confirmed it —
Hercules is freely licensed, but guest operating systems carry their own licenses and
z/OS is licensed to a machine.

Accepting that and stopping would have been the wrong conclusion, because it answers
a question that was never the useful one. The gap this change addresses is the
reframing: *not* "can Hercules replace z/OS" (no), but **"which specific unknowns in
the WTO call path can Hercules retire before mainframe access?"** — to which the
answer turns out to be *most of them*.

A second, quieter gap: the repo had `docs/implementations/` but no mechanism for
recording *decisions* or for stating claims falsifiably before testing them. For a
project whose subject matter is unusually easy to be fooled by — tolerant emulators,
an ancestor interface that looks compatible until it isn't, a console message that
looks like success — that was a real methodological hole.

## 2. What Changed

Documentation and methodology only. **No code was modified.** No exercise, assembly
file, test, or `go.mod` was touched.

| File | Change |
|---|---|
| `docs/decisions/README.md` | ADR format, rules, and index — methodology ported from CASSANDRA. |
| `docs/decisions/0001-emulation-strategy-hercules-two-track.md` | The governing decision: U1/U2/U3 unknown decomposition, z/OS-on-Hercules ruled out, Track M primary / Track L time-boxed, the E0–E3 ladder, and the provenance rule. |
| `docs/hypotheses/README.md` | Pre-registration format, rules, and index. |
| `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` | Pre-registers the oracle claim ADR 0001 depends on, with four sub-claims, three evidence lines, a three-outcome decision rule, and a negative control. |
| `docs/hypotheses/002-s390x-port-equivalence.md` | Pre-registers Phase 1b correctness under emulation, and explicitly refuses the performance claim. |
| `docs/evidence/README.md` | Mandatory provenance header, capture rules, both ladders. |
| `docs/research-briefs/README.md` | The Gemini interface: division of labour and the rules that make a return citable. |
| `docs/research-briefs/001-wto-parameter-list-authoritative-layout.md` | Highest-priority brief — the WPL byte layout, MVS 3.8 vs z/OS. |
| `docs/research-briefs/002-prior-art-census-go-assembly-mvs-services.md` | Re-validates the thesis's negative existence claim with a documented search surface. |
| `docs/runbooks/tk5-hercules-setup.md` | TK5-on-Windows setup to rung E0, with ⚠ VERIFY markers and machine-specific traps. |
| `research/README.md` | Where Gemini returns land, stored verbatim, and how to synthesise them. |
| `README.md` | Added the emulation program and the new documentation structure. |
| `docs/mainframe-baseline-strategy.md` | Cross-referenced to the E-ladder; Phase 1b status updated from blocked to unblocked. |
| `memory/MEMORY.md` | Current state, active plan, and the new conventions. |

## 3. Implementation Approach

### The unknown decomposition

The organising idea is that "can we test before hardware?" is unanswerable as posed,
because the WTO call path is not one thing. Split into unknowns, each gets a
different and clearer answer:

- **U1** — does our Go assembly emit correct s390x, and does the Go ABI hold on a
  big-endian 64-bit target? → **fully retirable** (QEMU / Hercules Linux s390x)
- **U2** — is the WTO parameter list byte-correct, does `SVC 35` accept it, what
  comes back in R15? → **substantially retirable** (MVS 3.8j via TK5)
- **U3** — `GOOS=zos`, Malloc31 / below-the-bar, AMODE 31↔64, USS → **not retirable
  by anything**

Every emulated result must declare which unknown it speaks to, in a provenance
header. That single field prevents the failure mode this whole structure exists to
avoid: a result quietly growing into a claim it cannot support.

### The E3 bridge — the load-bearing idea

Go cannot run on MVS 3.8j. But **the bytes can cross**, and that is enough.

A Go program on the laptop emits the parameter list it intends to build — EBCDIC-
converted text, length header, MCS flags — and dumps it as hex. That hex is embedded
in an MVS assembler program as `DC X'...'` constants and handed to a raw `SVC 35`. If
the operator console displays the message, the Go-side byte construction is verified
against a genuine `SVC 35` implementation, without Go ever executing on a mainframe.

Contract: *inputs* — a Go string; *output* — a byte sequence; *side effect* — a
console message on a real supervisor call; *invariant* — the bytes tested on MVS are
byte-identical to the bytes the Go builder will produce on z/OS.

This is a differential test across two machines that share a data format but not an
instruction set. It retires **four of the roadmap's six Phase 3b steps** (EBCDIC
conversion, parameter list construction, `SVC 35` dispatch, R15 evaluation) plus part
of a fifth (R1 linkage), leaving only `Malloc31` and the AMODE context — both U3,
both genuinely z/OS-only.

### Pre-registration

Both hypotheses fix their decision rules before evidence. H001 names three outcomes
(STRONG / PARTIAL / WEAK) with different consequences, a tie-break clause (the
assembler listing beats the manual for the MVS side), and a negative control — submit
a knowingly malformed length field and confirm it *fails*, because a supervisor that
tolerates a bad WPL would let an incorrect list pass and prove nothing.

## 4. Mathematical / Statistical Details

Not applicable — this change is structural and strategic. No formula, statistical
test, or numeric algorithm is introduced.

The nearest thing to a quantitative claim is the deliberately **refused** one: H002
declines to report any performance measurement from emulated execution, because
QEMU and Hercules both implement `TR` as a software loop. Timing `TR` under emulation
measures the emulator's translation loop, not the hardware operation that makes the
instruction interesting. The roadmap's promised `ebcdic` amd64-vs-s390x benchmark
table is therefore deferred to real hardware rather than produced with numbers that
would be meaningless in the specific way that is hardest to notice.

## 5. Design Decisions

| Alternative | Why not |
|---|---|
| **Pursue z/OS under Hercules** | Not legal, and not a timing problem that project progress could solve. Excluded permanently and on the record rather than left ambiguous. |
| **Skip Hercules; wait for LinuxONE / z/OS** | Leaves the project fully blocked on its highest risk, with U2 discovered at T3 on borrowed time in front of an operator console — the most expensive possible place to learn it. |
| **QEMU only; no Hercules** | Faster for U1, but QEMU cannot run MVS 3.8j, so U2 stays untouched. U2 is the unknown with no alternative oracle, which makes Track M the higher-value track despite being the slower one. |
| **Hercules-hosted Linux s390x as a required dependency** | Multi-hour Windows install whose known failure mode (TAP/CTCI networking) has nothing to do with this project's subject matter. Demoted to a time-boxed experiment; QEMU carries the Phase 1b inner loop. |
| **Assert MVS/z/OS WTO compatibility and proceed** | The tempting shortcut, and the one that would quietly compromise the thesis. The WPL length-field semantics could not be established from *any* secondary source in the research pass, which is exactly the signal that this belongs in a hypothesis rather than in a premise. |
| **Report emulated benchmarks with a caveat** | Caveats get dropped when numbers get quoted. Refusing to produce them is the only reliable protection, and it matches the repo's existing honest-benchmark convention. |
| **Write ADRs after doing the work** | Defeats pre-registration. The value of a decision rule is entirely in having fixed it before seeing the result. |

**Why Track M before Track L**, given the roadmap's phase order puts the s390x port
first: Track M is bounded (a download with a bundled emulator) and is the *only*
oracle for U2, whereas Track L's job is already covered faster by QEMU. Ordering by
"which unknown has no alternative oracle" beats ordering by roadmap sequence when the
sequence was written under the assumption that neither was available.

## 6. Verification

This change ships documentation, so verification is structural plus the pending
execution gate.

**Structural — runnable now:**

```powershell
# All new documents exist and are non-empty
Get-ChildItem -Recurse docs\decisions, docs\hypotheses, docs\evidence, docs\research-briefs, docs\runbooks, research |
  Where-Object { -not $_.PSIsContainer } |
  Select-Object FullName, Length
```

```powershell
# No code was touched by this change
git status --short
git diff --stat -- '*.go' '*.s' 'go.mod'   # must be empty
```

**Regression — the Phase 1 exercises must be unaffected:**

```powershell
# from each exercise directory
go vet ./...
go test ./...
```

**Execution gate — owner runs, per the runbook:**

1. Follow `docs/runbooks/tk5-hercules-setup.md` §1–§7.
2. Rung E0 passes when a job's output listing and its return code can be produced.
3. Capture `docs/evidence/E0-tk5-boot-<date>.md` with the full provenance header and
   the §9 checklist — **including findings for every ⚠ VERIFY marker**, which is what
   upgrades the runbook from a draft to a verified document.

**Research gate — owner runs through Gemini:**

Briefs 001 and 002. Brief 001 is the higher priority: rung E2 cannot be designed
without an authoritative answer on the length field.

## 7. Related Docs

- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — the governing decision
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` · `002-s390x-port-equivalence.md`
- `docs/runbooks/tk5-hercules-setup.md`
- `docs/research-briefs/001-…` · `002-…`
- `docs/mainframe-baseline-strategy.md` — the T0→T3 real-hardware ladder this feeds
- `docs/codex-handover.md` — project state and the conventions this change preserves
- `docs/implementations/2026-07-05-phase1-explainer-baseline-handover.md` — predecessor

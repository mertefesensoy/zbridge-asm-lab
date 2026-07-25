# Mainframe Baseline Strategy — Early z/OS Access

**Date:** 2026-07-05
**Status:** active planning document
**Trigger:** Phase 1 is complete and direct IBM Z (z/OS) access may arrive ahead of the
roadmap's original schedule, potentially before Phase 1b (LinuxONE port) and Phase 2
(go-recordio annotation) are done in order.

---

## 1. The governing principle

**Mainframe access time is the scarcest resource in this project.** The roadmap's own
risk register names z/OS access timing as the single biggest risk. Therefore:

> Everything that does not require the mainframe is done before or in parallel with
> mainframe work. Every minute on the machine is spent climbing the test ladder or
> capturing evidence — never writing code that could have been written on the laptop.

This document defines what "ready" means, what gets skipped, what must never be
skipped, and the exact order of operations for the first sessions on real hardware.

---

## 2. How early access changes the roadmap

The original phase order was: **1 → 1b (LinuxONE s390x) → 2 (annotation) → 3 (z/OS)**.
With direct z/OS access, the order bends but the *dependency logic* survives:

| Phase | Original plan | With early z/OS access |
|---|---|---|
| 1b — port exercises to s390x | LinuxONE Community Cloud, two weeks | **Unblocked off-mainframe as of 2026-07-25** (ADR 0001). Correctness is verified under emulation (QEMU s390x inner loop; Hercules Linux s390x time-boxed) — see H002. What remains for hardware is confirmation and the *benchmark table*, which emulation cannot produce honestly. Still absorbed into 3a for the hardware half. |
| 2 — go-recordio `utils.s` annotation | Mid July – early August, reading only | **Unchanged, runs in parallel.** Needs no hardware. This is laptop/evening work and must not consume mainframe time. The SAM31/SAM64 and Malloc31 sections become *prerequisites for T3* (see ladder below). |
| 3a — toolchain validation | Cross-compile `GOOS=zos GOARCH=s390x`, run on USS | **Becomes Day 0.** First thing done on the machine. |
| 3b — WTO scaffold | After 2 + 3a | Unchanged, gated on T2 + the Phase 2 sections that cover Malloc31/SAM31. |
| 3c — validation + docs | After 3b | Unchanged. |

**What is genuinely skippable:** LinuxONE provisioning, *if* the z/OS environment gives
a usable USS shell with either a Go toolchain or a way to run cross-compiled binaries.
LinuxONE remains the fallback if z/OS access is delayed or time-boxed too tightly.

**What is never skippable:**

1. **The `add` checkpoint on the new target (T1).** It found toolchain problems on
   amd64; it will find them on s390x. Never debug WTO and the toolchain simultaneously.
2. **The go-recordio annotation of Malloc31 + SAM31/SAM64.** The WTO parameter list
   must be 31-bit addressable. Writing the scaffold without understanding go-recordio's
   below-the-bar pattern means debugging storage semantics on borrowed machine time.
3. **EBCDIC round-trip verification on the real code page (T2).** The tables were
   verified against ICU on the laptop; the hardware TR implementation must reproduce
   the amd64 outputs byte-for-byte before any message is handed to WTO.

---

## 3. Day-0 checklist (first session on the machine)

Capture facts first, run code second. If the access window closes unexpectedly, the
facts determine what can be prepared offline for the next window.

**Environment facts to record (in `docs/zos-environment.md`, created on Day 0):**

- [ ] Access type: Wazi aaS / shared LPAR / other. SSH into USS? TSO? Job submission?
- [ ] z/OS version (`uname -svr` in USS).
- [ ] Is IBM Open Enterprise SDK for Go installed? (`go version`) Which Go version?
- [ ] If no Go on the box: can cross-compiled `GOOS=zos GOARCH=s390x` binaries be
      uploaded and executed? (This is the assumption 3a validates.)
- [ ] **Console visibility — critical for the endgame.** Can I see the operator console
      or SYSLOG/OPERLOG (SDSF `LOG`)? WTO success is defined as "message appears on the
      console"; without read access to the console or syslog, T3 cannot be *verified*.
      If visibility is missing, raise it with the mentor immediately — it is an access
      requirement, not a nice-to-have.
- [ ] User authority: problem state, key 8 is expected and sufficient (unauthorized WTO
      carries an `@`-prefix on the console — cosmetic, not a block). Note whether APF
      authorization exists anyway.
- [ ] Time limits, idle timeouts, storage quotas of the environment.

**Then immediately climb to T1 (below) in the same session if time allows.**

---

## 4. The baseline test ladder

Each rung gates the next. A rung is "passed" when the listed evidence is captured and
committed to the repo (output logs go in `docs/evidence/`).

> **The E-ladder now runs before this one.** ADR 0001 adds an off-mainframe ladder
> E0→E3 on emulated hardware (TK5/MVS 3.8j for WTO semantics, QEMU/Hercules s390x for
> instruction correctness). Its purpose is to arrive at T1/T2/T3 with the work already
> debugged. In particular, **E3 retires four of the six Phase 3b steps** before any
> mainframe time is spent, re-scoping T3 from "construct a parameter list" to "port a
> verified one." See `docs/decisions/0001-emulation-strategy-hercules-two-track.md` §6.
> No E-rung result may be presented as a z/OS result.

| Rung | Gate | Command shape | Evidence to capture |
|---|---|---|---|
| **T0** | Pure-Go binary runs on target | `GOOS=zos GOARCH=s390x go build` a hello-world; run it on USS | stdout capture, `go version`, binary size |
| **T1** | `add` passes on the target | `go test ./add/` on-box, or cross-compiled test binary (`go test -c`) uploaded and run | full `go test -v` output |
| **T2** | All five exercises pass with real `_s390x.s` bodies | port order: `ebcdic` (TR) → `strmanip` (STH+MVC) → `regs` (R15/R14/R13) → `bytecmp` (CLC) → trap exercise | `go test -v` per module; benchmark table amd64 vs s390x for `ebcdic` (the roadmap's promised side-by-side: lookup loop vs TR) |
| **T3** | WTO scaffold: message on the operator console | Phase 3b steps 1–6 (Malloc31 → AtoE → param list → R1 → SVC 35 → R15) | console/SYSLOG screenshot or log extract showing the message (with its `@` prefix), return-code table |

Port-order rationale for T2: `ebcdic` first because TR is the highest-value single
instruction and the tests are a pure oracle (round-trip + ICU-verified tables);
`strmanip` second because the big-endian header becomes natively testable;
`regs` third because it *documents* the register conventions T3 depends on;
`bytecmp` fourth (CLC + `EX`/`EXRL` is the trickiest encoding, worth doing after
warm-up); the trap exercise last since its z/OS form *is* the start of T3.

Time-boxing: if a rung fails and the diagnosis is not obvious within the session,
capture the failure (exact command, exact output, `go env`), get off the machine, and
diagnose offline. Machine time is for experiments, not for staring.

---

## 5. Preparation to complete *before* access lands (laptop-only)

1. **Write the five `_s390x.s` implementations now**, on the laptop, as best-effort
   drafts. They cannot be tested here, but `GOOS=linux GOARCH=s390x go vet ./...` and
   `go build` (cross-compile) catch syntax and frame-contract errors. Arriving with
   drafts converts T2 from "writing assembly on the mainframe" into "debugging
   assembly on the mainframe" — a much cheaper activity.
2. **Phase 2 annotation, at least the WTO-relevant sections**: Malloc31, SAM31/SAM64
   around BALR, and the parameter-list construction in go-recordio's `utils.s`.
   Deliverable: notes identifying which patterns are SSREQ-specific vs. transferable.
3. **Script the ladder**: a small shell script per rung (build, upload, run, capture)
   so on-machine sessions are mechanical.
4. **Confirm with the mentor**: console/SYSLOG visibility (see Day-0 checklist), and
   whether the access path is Wazi aaS, a shared LPAR, or the "more open Z environment"
   mentioned in the roadmap — the table below adjusts expectations per path.

### Access-path notes

| Path | Expectation | Watch out for |
|---|---|---|
| **Wazi aaS** (IBM Cloud) | Full z/OS dev/test image; most likely to have USS + SDSF + freedom to install Go | entitlement/cost limits; instance lifecycle (deprovisioning between sessions — keep everything in git) |
| **Shared LPAR** | Real production-like system | least freedom: Go may be absent (cross-compile path), console visibility may need to be requested; be a polite tenant — WTO writes to a *shared* console, keep messages obviously tagged and low-volume |
| **LinuxONE Community Cloud** (fallback) | s390x Linux, easy self-service | it is *not z/OS*: validates T1/T2 instruction work (Linux s390x) but not the z/OS syscall conventions or WTO; the T2→T3 jump still needs z/OS |
| **Z Xplore** | ruled out by the roadmap | not flexible enough for custom assembly |

---

## 6. Baseline for the actual Go ↔ IBM Z bridge (pending scope)

The user will provide the concrete scope for the production bridge module separately.
What is already fixed by the roadmap, independent of that scope:

- **Shape:** a standalone public Go module (working name in the roadmap:
  `go-mvscalls` family), BSD-3-Clause to match go-recordio.
- **First service:** `WTO(message string) error` — pure Go assembly for s390x/z/OS,
  no cgo, no Language Environment dependency beyond the Malloc31 touchpoint.
- **Candidate second wave (roadmap Phase 4):** WTOR (ECB → Go channel bridging),
  Name/Token services (IEANTCR / IEAN4RT), and the polished standalone EBCDIC package.
- **Testing doctrine carried over from the lab:** oracle tests where a reference
  implementation exists; honest benchmarks with the pure-Go baseline; per-arch
  assembly behind filename build constraints; stubs that fail the build if targeted.

When the scope arrives, the baseline work is: freeze the API surface, define the
error model (R15/reason-code → typed Go errors), and set the repo skeleton with CI
that cross-compiles for `zos/s390x` on every push even when tests can't run there.

**Open item for the user:** provide the bridge scope (which services, which API
surface, single module vs. family) — tracked in the Codex handover as the first
blocked task.

---

## 7. Risk register (delta from the roadmap)

| Risk | Change vs. roadmap | Mitigation |
|---|---|---|
| Access window is short/time-boxed | new — early access may be trial-shaped | ladder + scripts prepared offline; evidence captured immediately; everything in git before logging off |
| Go missing on z/OS image | unchanged | cross-compile path validated at T0 (this is exactly what 3a exists to prove) |
| Console not visible to my user | new — verification blocker for T3 | raise with mentor at Day 0, not at T3 |
| Skipping Phase 2 under time pressure | new temptation with early access | hard gate: T3 does not start until the Malloc31/SAM31 annotation sections exist |
| LinuxONE skipped, then z/OS access slips | new | LinuxONE provisioning remains the standing fallback; nothing in this plan burns that bridge |

---

## Related docs

- [ADR 0001 — Hercules two-track emulation strategy](decisions/0001-emulation-strategy-hercules-two-track.md) — the E-ladder that feeds this one.
- [H001 — MVS 3.8j as a WTO oracle](hypotheses/001-mvs38j-svc35-wto-oracle.md) · [H002 — s390x port equivalence](hypotheses/002-s390x-port-equivalence.md).
- [TK5 setup runbook](runbooks/tk5-hercules-setup.md) — Track M, rung E0.
- [Interactive module explorer](interactive/zbridge-module-explorer.html) — what each Phase 1 module does and how.
- [Codex handover](codex-handover.md) — machine-readable project state for AI-assisted continuation.
- [Roadmap PDF](../zbridge-asm-roadmap.pdf) — at the repo root since 2026-07-25. Phase definitions, the WTO endgame rationale, the risk register, and the mentor's open questions.

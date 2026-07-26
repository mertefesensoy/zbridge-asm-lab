# ADR 0001 · Hercules is a two-track semantic laboratory, not a z/OS substitute

- Status: Accepted
- Date: 2026-07-25
- Owner: Mert Efe Şensoy · Mentor: Jürgen Holtz (IBM), who supplied the Hercules
  entry point (<https://hercules-390.github.io/html/>) via a former colleague with
  operational Hercules experience
- Resolves: the "middle ground before real hardware" question raised 2026-07-25
- Builds on: the roadmap PDF §Risks ("Hercules is not legal for z/OS"),
  `docs/mainframe-baseline-strategy.md` (the T0→T3 ladder and the scarcity
  principle), `docs/codex-handover.md` §3 (the WTO call-path decomposition)

## Context

Phase 1 is complete: six exercises rehearse the WTO call path on amd64. Two things
are blocked. Phase 1b (port to s390x) waits on LinuxONE Community Cloud access.
Phase 3 (the WTO endgame) waits on z/OS access, which the roadmap names as the
single biggest risk in the project.

The owner proposed Hercules as a middle ground so that system-building can proceed
before real hardware, on the model of the automated, documented workflow used on the
CASSANDRA project. The mentor independently supplied the Hercules documentation
entry point.

*(Note on novelty: Research Brief 002 returned 2026-07-25, exhaustively confirming that no public Go-assembly WTO implementation exists. This upholds the project's novelty claim and raises the value of the emulation ladder described below, as the artifact being built is a documented first.)*

The naive form of the proposal — *run z/OS under Hercules* — is not available, and
the roadmap already says so. Accepting that and stopping would be the wrong
conclusion, because it answers a question that was never the useful one. The useful
question is narrower and has a much better answer:

> **Which specific unknowns in the WTO call path can be retired off-mainframe, and
> which oracle retires each one?**

This ADR answers that question and defines the program that follows from it.

## Evidence

All verified 2026-07-25 against primary or near-primary sources.

1. **Hercules emulates S/370, ESA/390, and z/Architecture (64-bit).** The actively
   maintained fork is SDL Hercules 4.x "Hyperion"
   (<https://github.com/SDL-Hercules-390/hyperion>); it is OSI-certified open source
   under the Q Public License. The site the mentor supplied documents the 4.0
   Hyperion lineage and links onward to the turnkey distributions and communities.
2. **Guest operating systems carry their own licenses.** The Hercules FAQ is explicit
   that running mainframe software under Hercules requires respecting that software's
   license terms. z/OS, z/VM and z/VSE are licensed to a machine; there is no legal
   path to a z/OS image for this project. This is a licensing fact, not a timing or
   effort problem, so no amount of project progress changes it.
3. **MVS 3.8j is freely runnable and is the direct ancestor of z/OS.** IBM
   distributed it as a no-charge product; in the United States it has been treated as
   public domain, and outside the US as "copyrighted software provided at no charge."
   It has been run openly under Hercules for over two decades. The maintained turnkey
   distribution is **TK5** (Rob Prins), Update 5, released 2026-02-18, which bundles
   Hercules SDL 4.9.1 64-bit for Windows along with configuration and manuals.
4. **Linux for s390x runs under Hercules.** Debian and Ubuntu s390x installs under
   Hercules are documented by multiple independent parties; Hercules 4.2+ is reported
   working with Ubuntu s390x.
5. **Go targets s390x as a first-class port.** `GOOS=linux GOARCH=s390x` builds out
   of the box, and s390x Go container images exist, so cross-building and emulated
   execution need no special toolchain work.
6. **TK5's assembler is Assembler XF (IFOX00).** High Level Assembler (ASMA90) is a
   licensed IBM program product and is *not* part of TK5. IFOX00 assembles S/370
   instructions and lacks HLASM's extensions (dependent/named USING, long
   displacement, relative-immediate forms).
7. **Unauthorized WTO messages are console-prefixed, not blocked.** Descriptor code
   and authorization state determine a leading blank, `@`, `*`, or `+` on the
   console. This corroborates the roadmap's premise that problem-state WTO is
   functionally complete and only cosmetically marked.

## Decision

### 1. z/OS under Hercules is ruled out, permanently and on the record

Not deferred, not "pending a license" — excluded. Every document in this repo that
touches emulation states this. The project's z/OS work happens on real, entitled
z/OS access or it does not happen.

### 2. Hercules is adopted as a two-track laboratory

| Track | Guest | Architecture | What it is for |
|---|---|---|---|
| **Track M** (primary) | MVS 3.8j via TK5 | S/370, 24-bit | `SVC 35` / WTO semantics, the operator console, MVS linkage conventions, JCL and the mainframe operational model |
| **Track L** (time-boxed) | Debian/Ubuntu s390x | z/Architecture, 64-bit | Executing Go assembly on a real big-endian 64-bit z/Architecture implementation |

### 3. The unknown decomposition governs what each track may claim

This table is the load-bearing content of the ADR. It defines the boundary of every
claim any emulated result is allowed to make.

| Unknown | Statement | Retirable off-mainframe? | Oracle |
|---|---|---|---|
| **U1** | Does our Go assembly emit correct s390x, and does the Go ABI / frame-size contract hold on a big-endian 64-bit target? | **Yes, fully** | Track L; QEMU/Docker s390x for the fast loop |
| **U2** | Is the WTO parameter list byte-correct; does `SVC 35` accept it; what does it return in R15? | **Yes, substantially** | Track M |
| **U3** | `GOOS=zos` toolchain, Malloc31 / below-the-bar allocation, AMODE 31↔64 switching, z/OS extended-WPL behaviour, USS execution | **No. Nothing emulates this.** | Real z/OS only |

### 4. Track M is primary and starts immediately

Track M is a bounded download-and-run with a bundled emulator, and it is the only
oracle in existence for U2. U2 is also the most expensive unknown to carry onto
borrowed mainframe time, because it is the one that would otherwise be discovered by
trial and error at rung T3 with an operator console watching.

### 5. Track L is time-boxed; QEMU is the Phase 1b inner loop

Installing Linux under Hercules on Windows is a multi-hour exercise whose known
failure mode is TAP/CTCI networking, not anything to do with this project's subject
matter. Phase 1b's actual gate — "do the `_s390x.s` bodies produce correct results on
s390x?" — is answered faster and just as validly by cross-compiling and running under
QEMU s390x emulation. Therefore:

- **QEMU/Docker s390x is the inner loop** for Phase 1b: minutes to set up, seconds
  per test cycle.
- **Track L is a scheduled experiment with an 8-hour box**, valuable for full-system
  fidelity (real IPL, channel I/O, 3270) and for keeping one emulator in the toolkit,
  but it is *not a dependency* of any other work. If the box expires, Track L is
  recorded as deferred and nothing downstream is blocked.

### 6. A new emulated ladder, E0→E3, precedes the existing T0→T3

The existing ladder in `docs/mainframe-baseline-strategy.md` governs real-hardware
time and is unchanged. The E-ladder runs before it, off-mainframe, and each rung
produces committed evidence under `docs/evidence/`.

| Rung | Gate | Track | Retires |
|---|---|---|---|
| **E0** | TK5 IPLs; operator console readable; a JCL job submits and its output is retrievable | M | operational fluency |
| **E1** | A hand-written assembler program using the `WTO` macro puts a message on the MVS operator console | M | toolchain + console verification path |
| **E2** | The same message reaches the console from a **hand-built parameter list and a raw `SVC 35`**, with no `WTO` macro anywhere | M | **U2** — proves we understand the WPL, not merely how to invoke a macro |
| **E3** | The byte sequence our **Go** code produces is fired at a real `SVC 35` and reaches the console | M | the Go-side construction of steps 2, 3, 5, 6 of Phase 3b |

**E3 is the reason this program is worth running.** Go cannot execute on MVS 3.8j —
but the *bytes* can cross. A Go program on the laptop emits the parameter list it
intends to build (EBCDIC-converted text, length header, flags), dumps it as hex, and
that hex is embedded in an MVS assembler program as `DC X'...'` constants and handed
to `SVC 35`. If the operator console displays the message, the Go-side byte
construction is verified against a genuine `SVC 35` implementation without Go ever
running on a mainframe.

Measured against the roadmap's own Phase 3b step list:

| Phase 3b step | Retired by E3? |
|---|---|
| 1. Allocate parameter buffer below the bar via `Malloc31` | ❌ U3 — z/OS only |
| 2. Translate UTF-8 → EBCDIC IBM-1047 via `AtoE` | ✅ |
| 3. Construct the WTO parameter list | ✅ |
| 4. Load R1 with the parameter list address | ⚠️ partial — the linkage is verified, the AMODE context is not |
| 5. Issue `SVC 35` | ✅ |
| 6. Read R15, map to a Go error | ✅ |

**Four of six steps, plus half of a fifth, are retirable before z/OS access exists.**
That is the return on this program, and it converts rung T3 from *"invent a parameter
list on borrowed machine time"* into *"port a verified one."*

### 7. No emulated result may be presented as a z/OS result

Every file under `docs/evidence/` carries a provenance header naming the machine,
guest OS, architecture, and which of U1/U2/U3 the result speaks to. In the thesis and
in any published artifact, emulated findings are labelled as such. This convention is
non-negotiable: the project's contribution is credibility about a gap nobody has
documented, and a single overstated claim costs more than the entire emulation
program returns.

## Scope and what this decision does not claim

- **It does not claim MVS 3.8j's WTO is identical to z/OS's WTO.** That is a
  hypothesis, pre-registered separately as H001, with a divergence table and a
  decision rule. This ADR commits to *testing* the oracle relationship, not to
  assuming it.
- **It does not claim Track M validates 64-bit instruction work.** IFOX00 assembles
  S/370 only. All z/Architecture instruction validation (TR, CLC, MVC, STH, long
  displacement, `EX`/`EXRL`) belongs to Track L and QEMU, not to Track M.
- **It does not claim addressing-mode equivalence.** MVS 3.8j is 24-bit: parameter
  lists live below the 16 MB line. z/OS wants below the 2 GB bar in AMODE 31. These
  are the same *category* of constraint with different numbers, and the AMODE 31↔64
  switching that go-recordio performs has no MVS 3.8j analogue at all.
- **It does not reduce the need for Phase 2.** The go-recordio `utils.s` annotation
  still gates T3, because Malloc31 and SAM31/SAM64 sit squarely in U3 where no
  emulator reaches.
- **It does not change the roadmap's endgame or its phase definitions.** It supplies
  a preparation program that runs underneath them.

## What would reopen or reverse this decision

1. **Legitimate z/OS access arrives early and is generous.** If a Wazi aaS
   entitlement or an open Z environment gives unmetered, long-lived z/OS access, the
   marginal value of Track M drops sharply and the program should compress to
   whatever is still cheaper to do off-machine. Track M's value is inversely
   proportional to available mainframe time.
2. **H001 falsifies.** If MVS 3.8j's WPL diverges from z/OS's in a way that makes E2
   and E3 results non-transferable, Track M's scope narrows to operational learning
   and the E-ladder stops above E1. The divergence itself would still be worth
   publishing.
3. **Track L's time-box expires twice.** If Hercules-hosted Linux s390x fails a
   second scheduled attempt, it is abandoned in favour of QEMU permanently and
   recorded as a negative result about the toolchain, not a project blocker.
4. **A licensed emulated z/OS path appears.** If IBM offers a legally entitled
   emulated z/OS for students or ambassadors, it supersedes Track M for U2 entirely.
   This ADR should be revisited, not worked around.

## Links

- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — the oracle claim this ADR
  deliberately does not assert
- `docs/hypotheses/002-s390x-port-equivalence.md` — Phase 1b under emulation
- `docs/runbooks/tk5-hercules-setup.md` — the Track M execution runbook
- `docs/research-briefs/001-wto-parameter-list-authoritative-layout.md` — the Gemini
  brief that must land before E2
- `docs/mainframe-baseline-strategy.md` — the T0→T3 real-hardware ladder this
  E-ladder feeds
- Hercules documentation (mentor-supplied): <https://hercules-390.github.io/html/>
- SDL Hercules 4.x Hyperion: <https://github.com/SDL-Hercules-390/hyperion>
- MVS TK5: <https://www.prince-webdesign.nl/tk5>

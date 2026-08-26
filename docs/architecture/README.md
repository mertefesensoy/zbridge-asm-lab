# Architecture — start here

**Who this is for:** someone who can read Go and has written assembly for *some*
architecture, but has never touched a mainframe and doesn't yet know what "z/OS" or
"WTO" mean. If you already know this project cold, the other documents in this folder
go deeper; this one just gets everyone to the same starting line.

**What this folder is not:** a restatement of the ADRs in `docs/decisions/`, the
evidence files in `docs/evidence/`, or the roadmap. Those are the primary sources —
this folder explains *how the system that resulted from them actually works*, in the
order a new reader needs it, with the ADRs and evidence cited rather than repeated.
Where this document simplifies for a first read, it says so and points at the file
that has the precise version.

---

## 1. The one-paragraph version

This project builds a single Go function — `WTO(message string) error` — that puts a
line of text on the operator's screen on an IBM mainframe running z/OS, by directly
issuing a low-level mainframe instruction called a *supervisor call*. It does this in
**pure Go assembly**, with **no C code anywhere in the call path**. That second part —
no C — is the entire reason the project is interesting: every existing way to do this
from a language other than IBM's own assembler or C goes *through* C. Nobody has
published a way to skip it. This project is that missing path, and proving it is the
thesis.

If that paragraph made sense, skip to [§4](#4-the-central-problem-no-cgo). If any part
of it was unfamiliar, §2 and §3 build up the vocabulary first.

---

## 2. Background, for readers new to mainframes

### 2.1 What "a mainframe" actually is

A mainframe is not "a really big PC." It's a computer designed from the ground up for
a different job than the laptop this is probably being read on: running thousands of
business transactions per second, for decades, without going down, while multiple
completely independent programs share the same physical machine safely. IBM's current
mainframe line is called **IBM Z**; the machines run an operating system called
**z/OS**. z/OS traces its lineage back through a chain of operating systems —
OS/VS2 → MVS → MVS/XA → MVS/ESA → OS/390 → z/OS — that goes back to the 1960s. That
lineage matters later in this document, because it means an *old* system in that same
family can stand in for testing purposes, which is a big part of how this project made
progress without owning a mainframe.

### 2.2 What "the operator console" is

Someone has to be able to watch a mainframe while it runs and be told when something
needs attention — a job failed, a tape needs mounting, a disk is nearly full. That
person is the **operator**, and the screen (historically, and still today in emulation,
literally a terminal) they watch is the **operator console**. Any program running on
the system — an application, a subsystem, the operating system itself — can put a line
of text on that console. Writing to it is one of the most fundamental things a
z/OS program can do, on par with writing to `stderr` on Linux, except that on a
mainframe it's a formal, audited, access-controlled operation with its own dedicated
low-level instruction rather than a file descriptor.

### 2.3 What a "supervisor call" (`SVC`) is

On z/OS (and its ancestors), a normal program cannot ask the operating system to do
privileged things — allocate memory below a certain address, talk to a device, write to
the console — by calling a function directly. Instead it executes a special CPU
instruction, `SVC n`, where `n` selects which operating-system service is being asked
for. This is the mainframe's equivalent of a Linux syscall (`SYSCALL` on amd64, or the
`SVC 0` trap that Go's own `SYSCALL` instruction assembles to on s390x Linux — see
§2.4). The CPU stops running the calling program, the operating system's supervisor
takes over, does the privileged work, and control returns to the program afterward.

**`SVC 35` is the supervisor call for "Write To Operator"** — WTO. Writing a message to
the console is service number 35 in the table of supervisor calls. This project's
entire endgame is: build the exact bytes `SVC 35` expects, in Go assembly, and execute
that instruction.

### 2.4 What "s390x" is, and why it's not the same question as "z/OS"

**s390x is a CPU architecture** — the instruction set. It's the 64-bit big-endian
architecture IBM Z machines run. **z/OS is an operating system** that runs on that
architecture. This distinction matters enormously in this project, because:

- **Linux also runs on s390x.** IBM ships and supports Linux distributions for the same
  physical hardware and the same instruction set that z/OS runs on.
- Go **already** supports `GOOS=linux GOARCH=s390x` as a first-class cross-compilation
  target. Building and even *emulating execution* of s390x machine code is something
  you can do today, for free, with no mainframe and no license, because it's just
  Linux.
- Go **does not** support `GOOS=zos` at all — see §4.2.

So "does our assembly run correctly on s390x hardware" and "does our assembly work on
z/OS" are two genuinely different questions with two genuinely different answers, and
keeping them separate is one of this project's central design decisions (see §5).

### 2.5 What "EBCDIC" is

Every mainframe concept so far has an ASCII/Linux analogue. Character encoding is the
one place mainframes are flatly different. z/OS's native text encoding is **EBCDIC**
(Extended Binary Coded Decimal Interchange Code), not ASCII/UTF-8. The letter `'A'` is
byte `0xC1` in EBCDIC IBM-1047 (the specific EBCDIC variant z/OS's Unix-like subsystem
uses), not `0x41`. Any text this project sends to the console — including every WTO
message — has to be translated from UTF-8 to EBCDIC first. This project's `codepage`
package does that translation (see [`zbridge-module.md` §2](zbridge-module.md#2-codepage--the-one-finished-package)).

### 2.6 Glossary, for quick lookup

| Term | Meaning |
|---|---|
| **Mainframe / IBM Z** | IBM's large-scale, high-reliability computer line |
| **z/OS** | IBM's flagship mainframe operating system |
| **MVS 3.8j** | A 1981, no-charge ancestor of z/OS. Still runnable today under emulation. See §6 |
| **s390x** | The 64-bit big-endian CPU architecture IBM Z machines use. A Go `GOARCH` value |
| **SVC** | Supervisor Call — the mainframe instruction that asks the operating system to do privileged work. Analogous to a Linux syscall trap |
| **WTO** | Write To Operator — the specific mainframe service (`SVC 35`) that puts a line on the console |
| **WPL** | WTO Parameter List — the block of bytes `SVC 35` reads to know what to print |
| **Operator console** | The screen an operator uses to monitor and control a running mainframe |
| **EBCDIC** | The mainframe-native text encoding. Not ASCII. `'A'` is `0xC1`, not `0x41` |
| **JCL** | Job Control Language — the scripting language used to submit work (a "job") to a mainframe |
| **Hercules** | A free, open-source software emulator of mainframe hardware. See [`emulation-harnesses.md`](emulation-harnesses.md) |
| **AMODE / "below the bar"** | z/OS's addressing-mode system. "Below the bar" means below the 2 GB address boundary — some services require parameters to live there |
| **cgo** | Go's mechanism for calling C code from Go. This project does not use it. See §4 |
| **Plan 9 assembly** | Go's own, portable assembly dialect — not the target CPU's native assembler syntax. See §4.3 |

---

## 3. Why this is a Go project at all

Go is popular for backend services, and organizations that already run z/OS often want
to run modern services *near* their mainframe data — sometimes literally on z/OS itself,
under its Unix System Services (USS) subsystem, which can run a fairly ordinary-looking
Unix-like environment including a Go runtime, once IBM's z/OS-targeting Go toolchain is
available (see §4.2). A Go program running there that wants to tell the operator "job's
done" or "something's wrong" needs a way to call `SVC 35`. Today, nothing publishes a
way to do that without dropping into C. This project is building that missing piece as
a small, well-tested, honestly-documented library, and using the process of building it
as the subject of a mentored thesis.

---

## 4. The central problem: no cgo

### 4.1 What cgo is, and why the obvious path goes through it

**cgo** is Go's built-in bridge to C: you write (or link against) a C function, and Go
generates the glue to call it. IBM's officially supported way to invoke `SVC 35`-style
services from a high-level language is exactly this shape — a C-callable runtime
library (the **Language Environment**, LE) that already knows how to do the low-level
call correctly, wraps it in a normal C function like `__console2()`, and handles all
the register and addressing-mode bookkeeping for you. The "easy" way to give Go a WTO
function is: write a two-line C wrapper around `__console2()`, use cgo to call it from
Go, done. Every comparator this project's research turned up — including the Python
and Java ecosystems on z/OS (`docs/research-briefs/002-...md`) — takes exactly this
route: call through the C/LE runtime layer rather than issuing the supervisor call
directly.

### 4.2 Why this project refuses that path

The project's standing rule, stated plainly in
[ADR 0004](../decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §3: **never
cgo, not even as a fallback.** Two reasons, one practical and one architectural:

1. **cgo doesn't actually insure against anything real.** The cgo route needs the exact
   same hard dependencies as the direct route — IBM's Go fork and entitled z/OS access
   — *plus* a working C compiler and the full Language Environment runtime. If either
   of the real risks (no Go fork, no z/OS access) actually happens, the cgo version
   can't be built either. It costs the project its central claim while covering nothing
   that can actually go wrong. (ADR 0004 §3, roadmap errata S1.)
2. **The gap cgo would paper over is the actual thesis.** `ibmruntimes/go-recordio`
   (IBM's own BSD-3-Clause reference implementation) already proves that calling a
   mainframe system service directly from Go assembly — no C, no cgo — is *possible*,
   for one specific service (`IEFSSREQ`, a subsystem-request call). But nobody has
   published the same thing for an operator-facing service like WTO. That's not a
   trivia gap — WTO is one of the most fundamental things any mainframe program does,
   and "how do you reach it from a language other than assembler or C" is a genuinely
   open, useful question. Answering it *is* the project. Falling back to cgo when the
   going gets hard would mean answering a different, already-answered question instead.

### 4.3 What "pure Go assembly" means concretely

Go ships its own assembly dialect, called **Plan 9 assembly** after the operating
system it originated on. It is *not* the target CPU's native assembler syntax — it's a
portable-looking pseudo-assembly that Go's own assembler (`go tool asm`) translates
into real machine code for whichever `GOARCH` you're building. Writing "pure Go
assembly" means every non-Go instruction in this project's call path is written in this
dialect, assembled by Go's own toolchain, with zero C compiler involved anywhere in the
build. The catch, covered in depth in
[`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md), is that Go's s390x assembler simply
has no mnemonic for `SVC` at all — so "pure Go assembly" for this project additionally
means hand-encoding raw instruction bytes where the assembler's vocabulary runs out,
the same technique IBM's own `go-recordio` and `x/sys/unix` already use for their own
un-mnemonic'd instructions.

---

## 5. The central engineering problem: there is no test system

A normal software project tests against the real target. This project's real target —
a Go toolchain that can produce z/OS binaries, running on an actual z/OS system — does
not exist yet for this project. Upstream Go has never shipped a `zos/s390x` build
target (verified directly against the toolchain source, not assumed —
`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` finding F3); reaching one requires
IBM's own Go fork. Actual z/OS access requires an entitlement this project does not
currently have (see [ADR 0002](../decisions/0002-zos-under-hercules-permitted-by-ibm-backing.md),
which is **withdrawn** — read its §10 before citing it). So: how do you build and gain
confidence in a mainframe-calling function when you can neither compile for the
mainframe nor run anything on one?

### 5.1 The move that makes the rest of the project possible: split the problem into named unknowns

Instead of treating "we have no mainframe" as one big blocking fact, the project asks a
sharper question, laid out in
[ADR 0001](../decisions/0001-emulation-strategy-hercules-two-track.md):

> **Which specific unknowns in the WTO call path can be retired without a mainframe,
> and what stands in for the mainframe for each one?**

The answer is that the overall problem decomposes into exactly three independent
unknowns, and — this is the useful part — **two of the three turn out not to need real
z/OS at all**:

| | Unknown | Plain-English question | What answers it |
|---|---|---|---|
| **U1** | Assembly correctness | Does our hand-written Go assembly actually produce correct s390x machine code, and does Go's calling convention hold up on a real big-endian 64-bit CPU? | **QEMU**, a CPU emulator — no mainframe, no license needed. [Retired.](evidence-ladder.md) |
| **U2** | The WTO parameter list | What bytes does `SVC 35` actually expect, in what order, and does it accept bytes our Go code built? | **MVS 3.8j**, z/OS's freely-runnable 1981 ancestor, under an emulator called Hercules. [Retired.](evidence-ladder.md) |
| **U3** | z/OS-specific behavior | Does a Go program actually run *on z/OS itself* — the toolchain, the memory allocation rules, the addressing-mode switches, the Unix-like subsystem? | **Nothing emulates this.** Needs real, entitled z/OS access. **Still open.** |

Every piece of work in this project can be labeled with which of these three it's
retiring — including, honestly, "none of them, and here's why it's still worth doing."
That label is mandatory on every pull request
([`docs/team/charter.md`](../team/charter.md) §5).

### 5.2 Why U1 and U2 are answerable without a mainframe

**U1** is really a question about the CPU *architecture* (s390x), not about z/OS the
*operating system*. Since Linux also runs on s390x, and Go already cross-compiles to
`linux/s390x`, U1 can be tested by compiling for Linux-on-s390x and running the result
under **QEMU**, a general-purpose CPU emulator, entirely off a laptop. See
§2.4 above for why that separation (architecture vs. operating system) is valid.

**U2** exploits the mainframe lineage mentioned in §2.1. z/OS's `SVC 35` almost
certainly didn't spring into existence with z/OS — the same service, doing
substantially the same thing, has existed since the 1960s/70s ancestor systems. **MVS
3.8j**, a 1981 release, is old enough to be freely distributable today and is commonly
run under a mainframe hardware emulator called **Hercules**. If bytes built by this
project's Go code are handed to a *real* `SVC 35` implementation on MVS 3.8j and it
accepts them and puts a message on the console, that's strong (though not proof-positive
— MVS 3.8j is not z/OS) evidence the parameter-list format is understood correctly.
[`evidence-ladder.md`](evidence-ladder.md) walks through exactly how far this actually
got pushed, and it went further than "strong evidence" — it got to **byte-for-byte
verification against IBM's own macro output.**

**U3 stays open on purpose.** Nothing about memory allocation below a specific address,
switching between 31-bit and 64-bit addressing modes, or the z/OS-specific Unix
subsystem exists on MVS 3.8j (a 24-bit system with none of those mechanisms) or under
QEMU (which emulates a CPU, not an operating system). There is no honest shortcut here;
this piece waits for real, entitled z/OS access. See
[`emulation-harnesses.md`](emulation-harnesses.md) §4 for exactly where the emulation
boundary sits.

---

## 6. How the pieces fit together

```
                     ┌─────────────────────────────────────────────┐
                     │   docs/architecture/  (this folder)          │
                     │   the explanations — READ THESE FIRST        │
                     └─────────────────────────────────────────────┘
                                          │
              ┌───────────────────────────┼───────────────────────────┐
              ▼                           ▼                           ▼
  ┌────────────────────┐   ┌──────────────────────────┐   ┌────────────────────────┐
  │  zbridge/            │   │  E0 → E3 evidence ladder │   │  Emulation harnesses    │
  │  the production      │   │  proof the design works, │   │  QEMU (U1) and Hercules │
  │  Go module            │   │  captured under          │   │  (U2) — what each does  │
  │  (the "what we ship") │   │  docs/evidence/           │   │  and does NOT prove     │
  └────────────────────┘   └──────────────────────────┘   └────────────────────────┘
              │                           │                           │
              └───────────────────────────┴───────────────────────────┘
                                          │
                                          ▼
                         ┌───────────────────────────────┐
                         │  docs/decisions/ (ADRs)        │
                         │  docs/hypotheses/               │
                         │  the actual decision record —   │
                         │  this project never edits its    │
                         │  own roadmap PDF; it supersedes  │
                         │  lines with ADRs instead          │
                         └───────────────────────────────┘
```

Two more things worth knowing before reading further:

- **`zbridge-asm-lab` is two things sharing one repository.** One part (`add/`,
  `ebcdic/`, `strmanip/`, `regs/`, `bytecmp/`, `syscall-linux/` at the repo root) is a
  **teaching ladder** — six small, standalone Go modules, each with its own `go.mod`,
  that rehearse one piece of the eventual problem (byte translation, string internals,
  register discipline, byte comparison, raw syscalls) on progressively more realistic
  targets. The other part (`zbridge/`) is the **production library** the thesis
  actually ships. The lab modules are deliberately kept separate — see
  [`zbridge-module.md`](zbridge-module.md) §1 for why that separation is a hard rule,
  not an accident.
- **The roadmap PDF (`zbridge-asm-roadmap.pdf`, repo root) is the mandate and is never
  edited.** Where reality has moved past what it says, the correction lives in
  [`docs/roadmap-errata.md`](../roadmap-errata.md) and an ADR — never a silent PDF edit.
  If something in these architecture docs seems to disagree with the PDF, the errata
  file is very likely why, and it is authoritative over the PDF where the two conflict.

---

## 7. Where to go next

Read in this order — each builds on the last:

1. **[`zbridge-module.md`](zbridge-module.md)** — a package-by-package tour of the
   `zbridge/` Go module: what each package does, what it depends on, and why several
   packages are deliberately incomplete (and how they fail on purpose, honestly, rather
   than guessing).
2. **[`evidence-ladder.md`](evidence-ladder.md)** — the E0→E3 sequence of concrete,
   captured proofs that retired U1 and U2, in the order they actually happened, with
   what each one does and does not establish.
3. **[`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md)** — the single most important
   technical mechanism in the project, end to end: how a Go string becomes bytes on an
   operator's screen, one step at a time.
4. **[`emulation-harnesses.md`](emulation-harnesses.md)** — how QEMU and Hercules
   actually work as tools, what each one is (and is not) emulating, and exactly which
   claims each one can and cannot support.
5. **[`c4/`](c4/README.md)** — the same system, redrawn as C4 model diagrams at four
   levels of zoom, for readers who think better in pictures than in prose.
6. **[`testing.md`](testing.md)** — how the seven different kinds of "test" in this
   repository work, from ordinary unit tests through to evidence rungs that aren't
   `go test` at all.

For "how do I actually run any of this myself," see the repository's
`RUN.md`.

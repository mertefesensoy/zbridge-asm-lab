# The emulation harnesses: QEMU and Hercules

**Read [`README.md`](README.md) first**, specifically §5. This document goes deep on the
two tools that made [`evidence-ladder.md`](evidence-ladder.md) possible — what each one
actually is, what it's genuinely emulating versus what it just happens to run, and
exactly where each one's evidentiary authority stops. That last part is the important
part: **this project's central discipline is never claiming more than a tool actually
proved**, and you can't judge that without knowing precisely what each tool is.

---

## 1. Why two completely different tools, not one

It would be simpler to have one emulator that answers everything. This project uses two
different tools for a specific reason: **U1 and U2 are different *kinds* of question**,
and each tool is a much better fit for one of them than the other.

- **U1** ("does our assembly emit correct s390x machine code?") is a question about the
  **CPU instruction set**. It has nothing to do with any particular operating system.
- **U2** ("does `SVC 35` accept the bytes we built?") is a question about a specific
  **operating-system service's behavior**. It has almost nothing to do with which CPU
  architecture happens to be running underneath — a service call's *contract* is an OS
  concept, not a hardware one.

A tool that's excellent at emulating raw CPU instructions quickly (QEMU) is a poor fit
for "does this specific 1970s operating-system service behave a certain way," and a
tool built to run a whole vintage operating system faithfully (Hercules) is comparative
overkill — and much slower to iterate with — for "did this one instruction encode
correctly." Using the right tool for each question is what makes both U1 and U2
answerable in the same afternoon rather than one of them taking a week.

---

## 2. QEMU — answers U1, and only U1

### 2.1 What QEMU actually is

QEMU is a general-purpose **machine emulator and virtualizer**. This project uses it in
one specific mode: **user-mode emulation** (`qemu-s390x`, not `qemu-system-s390x`),
which does something narrower and faster than emulating a whole computer. In user-mode,
QEMU translates the CPU instructions of *one Linux binary* for a foreign architecture
into instructions the host CPU actually understands, on the fly, and forwards that
binary's system calls straight to the host Linux kernel. There is no boot process, no
virtual disk, no virtual network card, and critically — **no guest operating system at
all.** The "OS" the program thinks it's talking to is real: it's the same WSL2 Linux
kernel the whole harness runs under, just relaying calls from translated s390x code.

| Property | Value |
|---|---|
| Mode | User-mode emulation (`qemu-s390x`), not full-system |
| Version pinned in evidence | 10.2.1 (Debian 1:10.2.1+ds-1ubuntu3.1) |
| What it emulates | The s390x **instruction set** only |
| Guest operating system | **None** — there isn't one. Syscalls pass straight to the host Linux kernel |
| What runs under it | Individual, statically-linked, cross-compiled Go test binaries |

### 2.2 What this setup can prove

Exactly one thing, and it happens to be exactly what U1 asks: **given a sequence of
bytes our Go assembler produced for `GOARCH=s390x`, does a real s390x CPU (emulated at
the instruction level) execute them the way we intended, including Go's own calling
convention?** `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` answers yes, for all
five ported lab modules, 29 tests, including proof that a hand-encoded, un-mnemonic'd
instruction (`TR` — see [`zbridge-module.md`](zbridge-module.md) §2) both disassembles
back to the intended instruction *and* executes correctly. Full write-up in
[`evidence-ladder.md`](evidence-ladder.md) §6.

### 2.3 What this setup flatly cannot prove

Nothing about z/OS, MVS, or any mainframe operating-system behavior at all — there is
no operating system under emulation here, mainframe or otherwise. It also proves
nothing about **timing**: QEMU's user-mode CPU emulation does not run at anything
resembling real hardware speed, and specific instructions (see §4 below) are emulated
as software loops rather than single hardware operations, which would make any
stopwatch measurement here a measurement of QEMU, not of an s390x CPU. That's why the
project's performance comparison (`docs/roadmap-errata.md` entry S2) is an
*instruction-count and encoding* comparison, not a timing table, and why the actual
timing table is deferred to real hardware.

---

## 3. Hercules (running MVS 3.8j via TK5) — answers U2, partially

### 3.1 What Hercules actually is

Hercules is fundamentally a different kind of tool from QEMU: a **full mainframe
hardware emulator**. It emulates an entire IBM System/370, ESA/390, or z/Architecture
machine — CPU, memory, and crucially, the mainframe I/O model (channels and control
units) well enough that real, unmodified mainframe *operating systems* boot and run on
it, believing they're talking to real hardware. That's a much bigger claim than QEMU's
user-mode emulation makes, and it's why Hercules can answer a question QEMU
structurally cannot: **how does a specific operating-system service actually behave?**

Hercules itself is open source (OSI-certified, Q Public License) and has never been in
any legal doubt for this project. The actively maintained fork is **SDL Hercules 4.x
"Hyperion."** The mentor supplied the project's entry point into Hercules documentation
directly.

### 3.2 The guest: why MVS 3.8j, and specifically not z/OS

Here's the question a first-time reader almost always asks: *if Hercules can run a real
mainframe operating system, why not just run z/OS on it and skip all this ancestor
business?*

**Because the operating system itself carries its own software license, completely
separate from Hercules's own license, and z/OS is licensed to a specific, entitled
machine.** There is no legal path to a z/OS installation image for this project without
that entitlement — this was checked directly (Hercules's own FAQ is explicit that
running mainframe software under it still requires respecting that software's license)
and was tested against reality once: [ADR 0002](../decisions/0002-zos-under-hercules-permitted-by-ibm-backing.md)
briefly explored whether this project's relationship with IBM constituted such an
entitlement, and was **withdrawn within a day** when the owner's own research
established that no such agreement exists. [ADR 0001](../decisions/0001-emulation-strategy-hercules-two-track.md)
§1 is the standing rule: **z/OS under Hercules is ruled out for this project,
permanently, not "pending a license."**

**MVS 3.8j is different, and it's different for a specific, checkable reason.** It's a
1981 release, several operating-system generations before z/OS, and IBM distributed it
at no charge; in the United States it has been treated as public domain, and it has
been run openly under Hercules by hobbyists for over two decades with no license
dispute. Crucially for this project, it is a **direct ancestor** of z/OS — not a
different product, but an earlier point on the same lineage described in
[`README.md`](README.md) §2.1 (OS/VS2 → MVS → ... → z/OS). That lineage is *why*
running WTO on MVS 3.8j is informative about z/OS's WTO at all, rather than being a
completely unrelated exercise.

This project runs MVS 3.8j via **TK5** (maintained by Rob Prins), a turnkey
distribution that bundles a specific, pinned Hercules build together with a
pre-configured, ready-to-boot MVS 3.8j system — no manual DASD (virtual disk) setup, no
manual IPL configuration. TK5 Update 5 (released 2026-02-18) bundles Hercules
4.9.1.0-SDL.

### 3.3 What this setup can prove, and how far the proof reaches

**Directly:** whether a real, working `SVC 35` implementation — on the actual ancestor
system, not a reimplementation or a guess — accepts a given byte sequence as a valid
WTO parameter list and honors it. [`evidence-ladder.md`](evidence-ladder.md) walks
through exactly how far this got pushed: not just "a macro works" (E1), but "a
hand-built list with no macro works" (E2), and finally "bytes a Go program built are
accepted" (E3) — with the exact byte layout read straight out of the macro's own
assembler expansion, not inferred (see
[`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md) §3).

**Indirectly, and this is the load-bearing assumption the whole Track M program rests
on:** that MVS 3.8j's `SVC 35` behavior is similar enough to z/OS's `SVC 35` behavior
that a result on one says something useful about the other. This assumption has a name
in the project's own doctrine — **H001**, the pre-registered hypothesis that MVS 3.8j
is a valid stand-in oracle for z/OS's WTO — and it is explicitly *not* simply assumed
true. It's tested piece by piece, and where it's been tested, the answer has been
"partially, with at least one confirmed divergence": see §3.4.

### 3.4 What this setup explicitly, provably cannot show

Three separate limits, and they're different in kind:

1. **Nothing below the 24-bit addressing line.** MVS 3.8j predates z/OS's memory model
   entirely — there is no "below the 2 GB bar" concept, no AMODE 31↔64 switching, none
   of the addressing-mode machinery z/OS's WTO call path actually needs
   ([`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md) §4–5). This isn't a gap in the
   testing — it's a gap in what the *guest operating system itself* has any notion of.
   This is squarely **U3**.
2. **No 64-bit instruction validation.** TK5's assembler is **IFOX00 (Assembler XF)**,
   an S/370-only assembler — not the licensed, IBM-only High Level Assembler (HLASM),
   and not capable of assembling z/Architecture's 64-bit instruction extensions at all.
   Any claim about 64-bit s390x instruction *encoding* correctness comes from QEMU
   (§2), never from this track.
3. **A confirmed behavioral divergence, not just an untested gap.** Reading the primary
   IBM manual directly (`GC28-0683-2` p.210) established that MVS 3.8j issues **no
   return code** for the single-line WTO this project uses. Whether z/OS's WTO does
   the same is a real, open question with real evidence for "no, it's different" — see
   [`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md) §7 and
   [ADR 0004](../decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §2.2. This
   is the single clearest instance in the whole project of *why* the ancestor-oracle
   assumption has to be tested claim-by-claim rather than trusted wholesale.

**And, same rule as QEMU: no performance numbers, ever.** Hercules, like QEMU,
implements instructions like `TR` as software loops rather than single hardware
operations. Timing anything under it measures Hercules's own software translation
speed, not a mainframe's.

---

## 4. Side by side

| | QEMU (Track L) | Hercules / TK5 (Track M) |
|---|---|---|
| **What it emulates** | The s390x instruction set (user-mode) | An entire mainframe: CPU, channels, DASD |
| **Guest OS** | None — syscalls pass to the real host kernel | MVS 3.8j (1981, z/OS's ancestor) |
| **Speaks to** | **U1** — assembly correctness, Go ABI | **U2** — the WTO parameter list, `SVC 35` |
| **Iteration speed** | Seconds per test run | Minutes (full IPL, JCL job cycle) |
| **64-bit instructions** | Yes — this is exactly what it's for | **No** — IFOX00 is S/370 (24-bit) only |
| **z/OS-specific behavior (U3)** | No | No — MVS 3.8j has no below-the-bar/AMODE concept at all |
| **Performance numbers trustworthy?** | **No** — `TR` etc. are software loops | **No** — same reason |
| **Legal status of the guest** | N/A — no guest OS | MVS 3.8j: no-charge, public-domain-in-the-US, two decades of open community use |

---

## 5. How a job actually reaches MVS 3.8j — headlessly, no operator

One more thing worth understanding, because it's what makes the E0–E3 rungs
*reproducible* rather than a one-time, hand-operated demo: every rung in
[`evidence-ladder.md`](evidence-ladder.md) after E0 runs through a fully scripted
pipeline with **no 3270 terminal and no human operator at any point**:

1. A JCL (Job Control Language — the mainframe's job-submission scripting language) file
   is written or generated on the Windows host.
2. It's piped directly at a TCP port TK5 exposes as a **virtual card reader** — bash's
   built-in `/dev/tcp` is literally enough (`cat job.jcl > /dev/tcp/127.0.0.1/3505`); no
   special client software is required.
3. MVS's job scheduler, **JES2**, picks it up from the virtual reader exactly as if it
   had been fed on physical punched cards, queues it, and runs it.
4. Output — including anything written to the operator console — is captured to plain
   files: a virtual printer device writes to `prt/prt00e.txt`, and a second device
   configured specifically as a **hardcopy log** captures console traffic to
   `log/hardcopy.log`.
5. The harness polls the output file until it stops growing (job completion and output
   availability are two different events — `$HASP395` means the job *ended*, not that
   its output has finished being written) and then reads the result back.

The whole thing runs inside a WSL2 Linux environment specifically so the emulator's
virtual disks live on a real Linux filesystem rather than under OneDrive or on NTFS —
both of which introduce failure modes (sync interference, path/locking issues) that
have nothing to do with the mainframe work itself. Full operational detail, including
several real mistakes made and turned into standing rules (most importantly: **only
`HHC01412I Hercules terminated` in the log counts as proof of a clean shutdown** — a
process simply no longer running is not proof of anything), is in
`docs/evidence/E0-tk5-boot-2026-07-26.md` and the runbook at
`docs/runbooks/tk5-hercules-setup.md`.

---

## 6. What's left after both tracks

**U3, entirely.** Below-the-bar storage allocation, AMODE 31↔64 switching, the
`GOOS=zos` Go toolchain, and Unix System Services execution are all z/OS behaviors with
no stand-in in either tool described here. Nothing in this document's two tracks
touches U3, by design — see [`README.md`](README.md) §5.2 for why that's an honest
limit rather than a gap someone forgot to close.

Next: **[`c4/README.md`](c4/README.md)** for the same system redrawn as C4 diagrams, or
back to **[`README.md`](README.md)** for the full map of this folder.

# The evidence ladder: how U1 and U2 actually got retired

**Read [`README.md`](README.md) first**, specifically §5, if you haven't — this document
assumes you already know what U1/U2/U3 mean and why they're the project's real progress
measure, not the roadmap's phase numbers.

Knowing *that* two of three unknowns are retired is one thing. This document is about
*how* — walking through, rung by rung, exactly what was run, on what, and precisely what
each result does and does not prove. This project treats that distinction as load-bearing:
every claim has to be **cited**, **evidenced**, or explicitly **registered** as an open
assumption (`docs/goal-prompt.md` §4.1) — there is no fourth category, and "probably
works the same way" is not an acceptable substitute for any of the three.

---

## 1. Why a ladder, and why off-mainframe rungs come first

Mainframe access — and, before this project had *any* emulation set up, even emulator
setup time — is the project's scarcest resource. The guiding principle
([ADR 0001](../decisions/0001-emulation-strategy-hercules-two-track.md) §4, restated in
`docs/goal-prompt.md` §3) is: **everything that doesn't strictly require the mainframe
gets done before anything that does.** Concretely, that means an **E-ladder** (E for
"emulated") runs entirely off-mainframe and produces committed, citable evidence for as
much of the design as it possibly can — before a **T-ladder** (T for hardware) ever
touches real machine time. This document is entirely about the E-ladder, because as of
this writing the T-ladder hasn't started (it's gated on real z/OS access, same as U3
generally).

Two separate emulated tracks feed the E-ladder, because they answer two separate
unknowns and need two separate kinds of stand-in machine:

- **Track L** (QEMU) → answers **U1** (does the assembly emit correct s390x machine
  code?). Doesn't appear as "E0–E3" below; it's a parallel track covered in §6.
- **Track M** (Hercules, running MVS 3.8j) → answers **U2** (is the WTO parameter list
  byte-correct; does `SVC 35` accept it?). This is the **E0 → E1 → E2 → E3** sequence
  that makes up most of this document.

Every rung's result lives in `docs/evidence/` as a Markdown file with a mandatory
**provenance header** — machine, guest operating system, CPU architecture, emulator
name and version, and which unknown the result speaks to. This isn't paperwork for its
own sake: it's the mechanism that makes "no emulated result is ever presented as a z/OS
result" (ADR 0001 §7) checkable rather than just promised. If you read nothing else in
this document, read the **"what this is not"** callout at the top of every rung below —
it's there in the actual evidence files too, and it's there on purpose.

---

## 2. Rung E0 — can this project drive an MVS system at all?

**What this is not:** a WTO result. E0 says nothing about `SVC 35` or the parameter
list. It's a pure operational-capability gate: before trusting *any* later result, can
the harness actually IPL a system, read its console, and get a job's output back —
**without a human sitting at a terminal**, since a human-operated step can't be
reproduced by teammates or rerun unattended?

**How:** [TK5](../runbooks/tk5-hercules-setup.md) (a turnkey MVS 3.8j distribution, see
[`emulation-harnesses.md`](emulation-harnesses.md) for what it actually is) running
under its own bundled Hercules, entirely inside a WSL2 Linux environment — no 3270
terminal, no operator. Three concrete gates:

1. **TK5 IPLs unattended.** Historically the biggest risk here is that MVS's boot
   sequence stops and waits for an operator to answer a prompt. This project's harness
   arms Hercules' "Automatic Operator" (a scriptable auto-responder) with the expected
   prompts *before* triggering the boot, so the outstanding-reply prompt
   (`IEA101A SPECIFY SYSTEM PARAMETERS...`) gets answered without a human ever seeing
   it.
2. **The operator console is readable as a file**, not just as pixels on a terminal —
   critical, because a later rung (E1) needs to *capture* a console message as
   evidence, not just glance at a screen. A device configured as a virtual 1403 printer
   captures everything to `log/hardcopy.log`.
3. **A job goes in and its result comes back out — through a socket, not a keyboard.**
   TK5's virtual card reader is configured to listen on a TCP port; the harness submits
   a JCL job by literally piping it at that port (`cat job.jcl > /dev/tcp/127.0.0.1/3505`
   — bash's built-in `/dev/tcp`, no extra tools needed), and reads the result back from
   a printer-output file once the job finishes. The test job is `IEFBR14`, the classic
   MVS no-op: it allocates nothing and returns 0. If it comes back `COND CODE 0000`,
   the whole pipeline — JCL in, JES2 (the job scheduler) queues it, output comes back
   out — works.

**Result:** all three gates passed, in one clean cycle, entirely headlessly, on
2026-07-26. Full detail, including several *operational* mistakes made and corrected
afterward (a detached Hercules process dying when its holding shell exited; two
emulator instances briefly running against the same disk state; a false "it stopped"
signal that turned out to mean nothing) is in
`docs/evidence/E0-tk5-boot-2026-07-26.md`. Those mistakes matter to this document for
one reason: they're why the project's rule became **"`HHC01412I Hercules terminated` in
the log is the only accepted proof of a clean stop — not the absence of a running
process."** Process absence is equally consistent with a clean shutdown and with a kill
mid-write, and only one of those leaves the emulated disk trustworthy.

---

## 3. Rung E1 — get the WTO parameter list layout from a primary source

**What this is not:** proof that hand-built (non-macro) bytes work. That's E2. E1 is
about *finding out the layout in the first place*.

**The problem E1 solves:** nobody could say, with a citable source, what bytes `SVC 35`
actually expects. The obvious place to look — the primary IBM manual, `GC28-0683-2` —
turned out **not to document the parameter list's byte layout at all**
(`docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md`; a fully independent
research return had also claimed a specific quotation from that manual, and the
quotation simply wasn't there on inspection — the project's "verify the form number,
then verify the page" rule exists because of exactly this incident). Mainframe
parameter lists like this one are conventionally built by an assembler **macro**, not
hand-assembled by the programmer, and IBM's documentation describes what keywords the
macro accepts — not the raw bytes the macro expands into. So where does the byte layout
actually live?

**The answer, and it's a satisfying one: in the macro's own output.** Any assembler can
be told to print the fully-expanded source it actually assembled — every mainframe
assembler has some form of a `PRINT GEN` directive for this — which means the exact
bytes a macro generates are sitting in the assembler's own listing output, in plain
sight, for anyone who assembles the macro and asks to see the expansion. **That
listing is the primary source.** Nothing "documents" it in the sense of a manual
paragraph; the macro processor's own literal output *is* the specification.

**What was actually run:** an MVS assembler program consisting of a single WTO macro
invocation, `WTO 'ZBRIDGE TEST HELLO',MF=L` (`MF=L` — "list form" — asks the macro to
expand into data statements rather than executable code, which is exactly the shape
that reveals the parameter list layout), assembled with `PRINT ON,GEN,DATA`. The `DATA`
part matters specifically: without it, the assembler truncates long constants in its
printed listing, which would have silently cut off part of the very bytes being looked
for.

**Result:** the listing showed, unambiguously, a 2-byte length field, then a 2-byte
flags field, then the EBCDIC message text — full layout and the reasoning behind the
"+4" in the length field are in
**[`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md)**, which is dedicated to that byte
layout specifically rather than repeating it here. The same assembled program was then
linked and run, and its WTO macro call **did put a message on the console**, captured
in `log/hardcopy.log` as `+ZBRIDGE TEST E1 WTO REACHED THE CONSOLE` — the leading `+` is
itself a small, independently useful finding (it's how MVS marks an unauthorized,
"problem-state" WTO on the console — present, not blocked, just cosmetically flagged).

Full detail, including the actual listing output and a second finding about routed
(multi-console) WTOs: `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` §1.

---

## 4. Rung E2 — prove the layout is *understood*, not just that a macro works

**What this is not:** proof Go's construction of the bytes is correct. That's E3. E2
only proves a *human*, working from what E1 revealed, can build the parameter list by
hand.

**Why this rung has to exist separately from E1:** getting a macro to put a message on
the console proves the macro works. It does *not* prove the underlying byte layout
E1 read out of the listing is actually correct and complete — you could misread a
listing and still have the macro "work" around your misunderstanding, because you never
actually relied on your own reading of it. E2 closes that gap by **removing the macro
entirely** and building the parameter list by hand, from constants written directly in
the source:

```assembler
ZBE2SVC  CSECT
         SAVE  (14,12)
         LR    12,15
         USING ZBE2SVC,12
         LA    1,WPL          * R1 = address of the parameter list (MVS linkage)
         SVC   35              * WTO, issued directly — no macro anywhere
         RETURN (14,12),RC=0
         DS    0F
WPL      DC    AL2(MSGEND-WPL) * length: let the assembler compute it
         DC    XL2'0000'       * MCS flags: zero, minimal form
MSGTEXT  DC    C'ZBRIDGE TEST E2 RAW SVC 35 NO MACRO'
MSGEND   EQU   *
```

Note `AL2(MSGEND-WPL)` for the length field — the assembler computes it from the
distance between two labels rather than a human doing the arithmetic, which is
deliberate: it makes the "+4" rule (see
[`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md)) self-evident from the source rather
than something that could be gotten wrong by hand and coincidentally still work.

**Result:** the message reached the console — `+ZBRIDGE TEST E2 RAW SVC 35 NO MACRO` —
with **no `WTO` macro anywhere in the program.** This is the rung
[ADR 0001](../decisions/0001-emulation-strategy-hercules-two-track.md) §6 designates as
the one that actually retires U2, precisely because it proves comprehension of the
format rather than successful macro invocation.

---

## 5. Rung E3 — the rung the whole program was built to reach

**What this is not:** a z/OS result, or a claim that this exact byte sequence is
correct on z/OS. MVS 3.8j is z/OS's distant ancestor, not z/OS itself — see
[`emulation-harnesses.md`](emulation-harnesses.md) for exactly how far that similarity
can be trusted.

**Why this is "the reason this program is worth running,"** in the ADR's own words:
Go cannot execute *on* MVS 3.8j — there's no Go toolchain for a 1981 operating system,
and there never will be. But Go's **output** — the bytes a Go function decides to build
— can absolutely be moved onto MVS 3.8j and handed to a real `SVC 35`. That's the whole
trick: a Go program, running on an ordinary laptop, calls `console.EncodeWPL` (see
[`zbridge-module.md`](zbridge-module.md) §3.2) to build the exact parameter list it
would send to a real WTO call. Those bytes get formatted as assembler `DC X'...'`
(hex constant) statements — by `console.FormatDC`, the other half of this same seam —
and embedded directly into an MVS assembler program, which is then assembled, linked,
and run under Hercules.

The actual generation is automated end-to-end by `zbridge/cmd/gen-e3`, which emits a
complete, ready-to-submit JCL job:

```assembler
ZBE3GO   CSECT
         ...
         LA    1,WPL
         SVC   35
         DS    0F
WPL      DC    X'00250000E9C2D9C9C4C7C540E3C5E2E3'
         DC    X'40C5F340C2E8E3C5E240C2E4C9D3E340'
         DC    X'C2E840C7D6'
```

**No human wrote those hex bytes, and IBM's `WTO` macro was not involved anywhere in
producing them.** They came out of `EncodeWPL`, running on Windows.

**Result:** `+ZBRIDGE TEST E3 BYTES BUILT BY GO` — on the console, from a real `SVC 35`,
fed a parameter list Go built. Every step (assemble, link, run) returned condition code
0. This is verified twice over: once on MVS 3.8j (an external, one-time proof captured
as evidence), and — critically — **permanently, on every subsequent `go test` run**, by
`console/wpl_oracle_test.go`, which hard-codes the exact bytes IBM's own macro produced
in E1 as an "oracle" and asserts `EncodeWPL`'s output matches them byte-for-byte. That
second part means E3's property doesn't just sit in a one-time evidence file — it's a
regression test. If anyone ever changes `EncodeWPL` in a way that would break
compatibility with what a real `SVC 35` accepts, `go test` fails immediately, without
needing an emulator at all.

Full detail, including the operational mistakes hit and fixed along the way (a linker
flag override that silently stopped producing an object file; a job-card field limited
to 20 characters that a longer name overflowed silently into a JCL error), is in
`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` §3 and §6.

---

## 6. The other track: E-L, and how U1 got retired

U1 — does the hand-written assembly emit correct s390x machine code, and does Go's own
calling convention survive on a real big-endian 64-bit target — doesn't need a
mainframe *operating system* at all, only the right CPU *architecture* (see
[`README.md`](README.md) §2.4 for why that distinction is what makes this possible).
So this rung runs on an entirely different tool: **QEMU**, a general-purpose CPU
emulator, running plain Linux binaries — no MVS, no Hercules, no z/OS ancestry
involved.

**What was run:** the five lab modules' `_s390x.s` assembly bodies (`ebcdic`,
`strmanip`, `regs`, `bytecmp`, `syscall-linux`), cross-compiled on the Windows host with
`GOOS=linux GOARCH=s390x`, then executed — not just compiled, **executed** — under
`qemu-s390x`, a user-mode emulator that runs individual Linux binaries for a foreign
architecture directly.

**Result:** all 29 tests across all five modules passed. One of them deserves special
mention because it's the direct precedent for how `SVC 35` itself will eventually be
handled: `ebcdic`'s s390x implementation needs the `TR` (Translate) instruction, and Go's
s390x assembler simply has no mnemonic for it (see
[`zbridge-module.md`](zbridge-module.md) §2). The bytes were hand-encoded, and their
correctness was checked three independent ways — disassembled back to `tr` by GNU
`objdump`, run under QEMU against a differential byte-loop reference at eleven buffer
lengths, and checked against all 256 table entries — rather than trusted because they
"looked right." Full output, including the actual disassembly:
`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`.

This rung also produced the finding that makes U3 the hard remaining unknown: `go tool
dist list` on the exact Go version this project uses (go1.26.3) lists `linux/s390x` and
**no `zos/s390x`** — checked directly against the toolchain, not inferred. See
[`README.md`](README.md) §4.2.

---

## 7. What the ladder adds up to

| Rung | Track | Gate | Retires |
|---|---|---|---|
| E-L | L (QEMU) | 29 tests pass on real s390x machine code, emulated | **U1** |
| E0 | M (Hercules/MVS) | Unattended IPL, readable console, job in/output out | operational fluency (prerequisite for E1–E3) |
| E1 | M | `WTO` macro's own expansion reveals the byte layout | the ground-truth source for the layout |
| E2 | M | Hand-built list + raw `SVC 35`, no macro | **U2**, formally — comprehension proven, not just macro invocation |
| E3 | M | **Go-built** bytes accepted by a real `SVC 35` | the Go-side construction verified against a real implementation |

**Two of three unknowns retired, and neither one used a minute of real mainframe time.**
That's not a lucky accident — it's the entire payoff of decomposing "we have no
mainframe" into three separately-answerable questions in the first place (see
[`README.md`](README.md) §5).

**What's still open, honestly:** everything genuinely z/OS-specific. MVS 3.8j is a
24-bit ancestor system with no below-the-bar allocation concept, no AMODE 31↔64
switching, and no Unix System Services. Nothing in this ladder touches those — they're
U3, and [`emulation-harnesses.md`](emulation-harnesses.md) draws the line precisely.
There's also one specific, documented **divergence** rather than an unknown: MVS 3.8j
issues *no return code at all* for the exact kind of single-line WTO this project uses
(confirmed by reading `GC28-0683-2` p.210 directly), while z/OS's own WTO macro
documentation confirms it **does** return a result in R15 — meaning rung E3, however
convincing, structurally *cannot* verify what a Go caller should do with that return
register, because the ancestor system has nothing there to read. `zbridge.Error.HasCode`
(see [`zbridge-module.md`](zbridge-module.md) §6) exists specifically because of this
divergence.

Next: **[`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md)** — the byte-level detail of
exactly what E1 found and exactly what E3 proved, worked through step by step.

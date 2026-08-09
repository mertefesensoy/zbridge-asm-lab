# Phase checkpoint · 2026-07-27

**For:** Jürgen Holtz (IBM) · **From:** Mert Efe Şensoy
**Covers:** since our last meeting (~2026-07-05), against the phase plan in
`zbridge-asm-roadmap.pdf`
**Companion:** `2026-07-27-zos-access-and-toolchain.md` — the access questions.
This document is the progress report; that one is the ask.

---

## The 30-second version

Three things to take away, in order of importance:

1. **Phase 1 and Phase 1b are both complete** — and Phase 1b's stated blocker, LinuxONE
   Community Cloud access, was **dissolved rather than waited out.** All five exercises
   are ported to s390x and *executed*, on a laptop.
2. **Four of the six Phase 3b implementation steps are already validated**, ahead of z/OS
   access, through a structure the roadmap did not contain. The WTO parameter list has
   been built by our Go code and accepted by a real `SVC 35`.
3. **Phase 2 has not started**, and it is the roadmap's own "most important pre-z/OS
   phase." Its window is mid-July to early August. That is the honest schedule risk and
   it is next.

---

## Phase-by-phase status

| Phase | Roadmap window | Roadmap status at last meeting | Status now |
|---|---|---|---|
| **0** Foundation | folded into 1+2 | deferred | Unchanged — absorbed as planned |
| **1** Go assembly on x86 | through late June | in progress | **COMPLETE** (2026-07-05). Six exercises, all vet + test clean |
| **1b** Port to s390x | late June – mid July | **blocked on LinuxONE access** | **COMPLETE for correctness** (2026-07-25) via emulation. Benchmark deliverable deferred — see below |
| **2** go-recordio `utils.s` annotation | mid July – early August | not started | **NOT STARTED.** Still inside its window. Now the critical path |
| **3** WTO on z/OS | mid Aug – late Sept | dependent on 2 + z/OS access | **Substantially pre-validated.** 4 of 6 3b steps done off-mainframe |
| **4** Stretch | Q4 2026+ | not started | One item arguably shipped early (`codepage`) |

---

## Phase 1b — the blocker was removed by re-framing

The roadmap says Phase 1b is *"blocked on access"* to LinuxONE Community Cloud, and asks
whether the ambassador programme is a faster route.

**We did not need either.** `GOOS=linux GOARCH=s390x` is a first-class Go port, so the
exercises cross-compile on a Windows laptop, and `qemu-s390x` runs the resulting
big-endian binaries directly. **29 tests across the five modules pass on `linux/s390x`.**

What that changes for you: the LinuxONE question drops from *blocking* to *useful*. It is
still wanted, but only for one thing — see the next paragraph.

### What Phase 1b did NOT deliver, and why

The roadmap's Phase 1b deliverable includes *"the ebcdic exercise gets a side-by-side
comparison: the amd64 lookup loop versus the s390x `TR` instruction, **with benchmark
numbers attached**."*

**That number cannot be produced honestly under emulation**, and I am not going to
produce it. QEMU implements `TR` in software, as a loop — timing it measures QEMU's
translation loop, not the hardware operation that makes `TR` interesting. Depending on
the emulator's internals the result could show a speedup, a slowdown, or nothing, and
none of the three would say anything about an IBM Z machine.

So Phase 1b split into two deliverables with different completion dates: **correctness is
done; the performance table waits for real hardware.** That was written down as a
pre-registered hypothesis *before* any code ran, not as an excuse afterwards.

**This is the one place LinuxONE access still buys something specific.**

---

## Phase 3 — pre-validated ahead of schedule

The roadmap places Phase 3 after Phase 2 and after z/OS access. Most of subphase 3b turned
out to be reachable earlier, using MVS 3.8j under Hercules as an oracle for the parameter
list. Measured against the roadmap's own six steps:

| Phase 3b step | Status | How |
|---|---|---|
| 1. Allocate the buffer below the bar via `Malloc31` | **Not done** | Needs real z/OS. No emulator reaches it |
| 2. Translate UTF-8 → EBCDIC IBM-1047 via `AtoE` | **Done** | Verified on s390x; the EBCDIC in the accepted parameter list is our `AtoE` output |
| 3. Construct the WTO parameter list | **Done** | Layout read from IBM's own macro expansion; our Go construction accepted by a real `SVC 35` |
| 4. Load R1 with the list address | **Done on MVS** | Standard MVS linkage, verified twice |
| 5. Issue `SVC 35` | **Done on MVS** | Issued raw, no macro |
| 6. Read R15 and translate to a Go error | **Not validatable here** | See the correction below — this step needs re-scoping |

**Four of six, plus the linkage half of a fifth**, before z/OS access exists. The value is
that rung T3 on borrowed machine time changes from *"invent a parameter list while an
operator watches"* into *"port a verified one."*

---

## Three corrections to the roadmap, from direct evidence

These matter more than the progress, because they are places where building the plan
exactly as written would have produced something wrong. I would rather bring them to you
now than discover them on machine time.

### 1. Phase 3b step 3 describes the parameter list incompletely

The roadmap says: *"Construct the WTO parameter list (2-byte length header followed by
EBCDIC message text, per IBM docs)."*

**There is a 2-byte MCS flags field between the length and the text.** Building it as
written would produce a parameter list that is wrong from byte 2 onward. From IBM's own
macro expansion on MVS 3.8j:

```
000000 0016              DC  AL2(22)                TEXT LENGTH
000002 0000              DC  B'0000000000000000'    MCS FLAGS
000004 E9C2D9C9C4C7C540  DC  C'ZBRIDGE TEST HELLO'  MESSAGE TEXT
```

Layout: **length (2) · MCS flags (2) · EBCDIC text**, fullword-aligned, and the length
field is `len(text) + 4` because it counts its own two bytes and the two flag bytes as
well as the text. Confirmed at two message lengths.

There is also a trap inside the routed form: with descriptor and routing codes present,
MCS flag bit 0 is set and two more halfwords are appended **after** the text — and the
length field *still* reads 22. It never covers them.

**And "per IBM docs" is not available.** The layout is not documented in prose in the MVS
supervisor manual at all. I retrieved GC28-0683-2 and read it page by page to confirm
that. The authoritative instrument is the macro expansion itself (`PRINT GEN`), which is
what we used. **If you know of a z/OS publication that does document the byte layout, that
is the single most useful pointer you could give me.**

### 2. Phase 3b step 6 cannot mean the same thing on both systems

The roadmap says step 6 is *"Read R15 for the return code and translate to a Go error"*,
and page 2 lists SVC dispatch as *"return code read from R15."*

**MVS 3.8j issues no return code at all for a single-line, non-MLWTO WTO** —
GC28-0683-2 p.210 — and R1 comes back holding the message identification number instead.
z/OS does return one in R15. So this is a genuine divergence between ancestor and
descendant, not a gap in our testing, and step 6 is the one step the off-mainframe work
provably cannot retire.

Our error type carries a `HasCode` flag alongside the return code so that "the service
returned nothing to map" is a representable state rather than indistinguishable from
"returned zero". **Confirming z/OS's actual behaviour here is one of my questions for
you.**

### 3. The console prefix is `+`, not `@`

Page 1 says unauthorized WTO messages *"carry an at-sign prefix on the console."*
Observed on MVS 3.8j:

```
+ZBRIDGE TEST E2 RAW SVC 35 NO MACRO
```

A **plus sign**. The substance of the roadmap's claim is confirmed — the message is
console-*prefixed*, not blocked, so the architectural payoff is full — but the specific
character differs on the ancestor. Whether z/OS uses `@` I have not verified.

---

## A structural change worth explaining

The roadmap's risk register names z/OS access timing as *"the single biggest risk"* and
notes correctly that **Hercules is not legal for z/OS.** I tested that boundary this week
and confirmed it: I briefly recorded a decision that our project's IBM backing might
permit a z/OS guest, researched it, found no such agreement exists, and withdrew the
decision the next day. Nothing was published in the meantime because the claim was flagged
as unverified from the moment it was written.

What came out of that is the useful part. Instead of asking *"can we get z/OS?"*, the
work was decomposed by **which unknown each question actually is**:

| | Question | Retired by | Status |
|---|---|---|---|
| **U1** | Does our Go assembly emit correct s390x; does the Go ABI hold on big-endian 64-bit? | QEMU | **Retired** |
| **U2** | Is the WTO parameter list byte-correct; does `SVC 35` accept it? | MVS 3.8j under Hercules | **Retired** |
| **U3** | `GOOS=zos`, `Malloc31`, AMODE 31↔64, USS | Real entitled z/OS only | **Open** |

Two of three fell without a mainframe. That decomposition is why the licensing answer
being "no" cost the schedule nothing.

It also produced a four-rung ladder that runs entirely off-mainframe — system up, macro to
the console, hand-built parameter list with raw `SVC 35`, then **our Go code's bytes** fed
to a real `SVC 35`. All four passed on 2026-07-26. Go never ran on the mainframe; its
bytes crossed.

---

## Where the schedule actually stands

**Phase 2 is the risk, and it is the next thing I do.** The roadmap calls it the most
important pre-z/OS phase, its window is mid-July to early August, and it has not started.
It needs no hardware and no access — it is reading `ibmruntimes/go-recordio/utils/utils.s`
line by line and annotating it.

It also now has a second reason to be urgent: it is the only route to the two Phase 3b
steps still open. `Malloc31` and the SAM31/SAM64 switching around the call are exactly
what that file contains.

Phase 1 finished about a week later than the roadmap's "late June", and Phase 1b about ten
days after "mid July". Phase 3's mid-August start is still reachable **if** Phase 2 runs
in the next two weeks and z/OS access resolves in parallel.

---

## Questions for you

Updated from the roadmap's page 8, with the ones that have since answered themselves
removed.

1. **Is there a citable z/OS publication for the WTO parameter list byte layout?**
   (*z/OS MVS Data Areas*, WPL structure, was suggested to me but I have not been able to
   verify a form number.) If the honest answer is "read the macro expansion", that is a
   useful answer and I will stop looking.
2. **Does a single-line WTO on z/OS actually return a code in R15?** This decides whether
   Phase 3b step 6 is a step or a non-step.
3. **How do I obtain and build IBM's Go fork for z/OS?** Upstream Go has no `zos/s390x`
   target — I verified this against the toolchain source. This is a *separate* question
   from a z/OS licence, and it is now the critical dependency for anything z/OS-side.
4. **What is the realistic z/OS access path and timeline?** ZD&T, Wazi aaS, a shared LPAR,
   the more open Z environment Lucas mentioned — or something else. (Roadmap Q2, still
   open.)
5. **Would you or Lucas review the Phase 2 annotated walkthrough when it ships?**
   (Roadmap Q3, unchanged — and now closer.)
6. **Is hand-encoding `SVC 35` via `BYTE` directives the accepted approach?** Go's s390x
   assembler has no `SVC` mnemonic, so it must be emitted as literal bytes. `x/sys/unix`
   does exactly this for `SVC 08`, which I take as a strong signal, but you would know
   whether it is idiomatic or merely expedient.
7. **Contacts at IBM doing low-level Go on s390x.** (Roadmap Q4, unchanged.)

### Roadmap Q1 — closed on our side

**"Is WTO the right scope, or should it be downscaled to WTO via `__console2()` through
cgo first?"** — I have closed this and removed the cgo fallback from the plan
([ADR 0004](../decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §3). I would
value you telling me if that is wrong, but here is the reasoning.

A fallback is insurance only if it survives the failure it covers. Lining up what each
route actually requires:

| Requirement | Direct `SVC 35` | cgo via `__console2()` |
|---|---|---|
| IBM's Go fork for z/OS | required | required |
| Access to a z/OS system | required | required |
| A C compiler on z/OS | not needed | **required** |
| Language Environment | `Malloc31` touchpoint only | **full dependency** |
| Parameter-list knowledge | required — **and obtained** | not needed |

**The cgo route's dependencies are a strict superset of the direct route's.** The two risks
still open are z/OS access and the Go fork — and if either bites, the cgo version cannot be
built or run either. It insures against nothing that can actually happen, while costing the
thesis its stated "no cgo, no Language Environment dependency."

The one risk the hedge genuinely covered — *"we may not be able to work out the parameter
list"* — is exactly the risk rungs E1–E3 retired.

**What I kept:** the `__console2()` comparison Phase 3c already plans, reframed as a
**measurement baseline** rather than a fallback deliverable. "As fast as the C path, with
no C, no LE, and a static binary" is a stronger result than either half alone, and it needs
cgo only in a test harness, never in the shipped module.

---

## Evidence, if you want to check anything

| Claim | File |
|---|---|
| s390x port, 29 tests, QEMU | `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` |
| System up, job in and out | `docs/evidence/E0-tk5-boot-2026-07-26.md` |
| WPL layout, raw `SVC 35`, Go bytes accepted | `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` |
| GC28-0683-2 read page by page | `docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md` |
| Why Hercules, and what it may not do | `docs/decisions/0001-…-hercules-two-track.md` |
| The z/OS-guest decision and its withdrawal | `docs/decisions/0002-…-ibm-backing.md` §10 |
| Visual walkthrough of all of it | `docs/interactive/wto-explainer.html` |

Every emulated result carries a provenance header naming the machine, guest OS,
architecture and emulator version. **Nothing in this project is presented as a z/OS result
until it runs on z/OS.**

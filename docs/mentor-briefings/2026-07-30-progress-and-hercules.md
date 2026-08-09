# Email draft · 2026-07-30 · Progress, and what Hercules could and could not do

> **STATUS: SENT.** A shortened, edited version of this draft went out the week of
> 2026-08-02, and Jürgen replied. The text below is the draft as it was prepared and
> reviewed pre-send — it is **not** a transcript of what actually went out (the sent
> version was shorter: two questions rather than four, and no mention of Lucas or the
> go-recordio walkthrough review). Neither the exact sent text nor Jürgen's reply is
> filed in this repo yet; whether and how to file that correspondence is a separate,
> owner-scoped decision, tracked in [issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1)
> (Phase 0).
>
> **Context this draft assumed at the time:** the 2026-07-27 meeting did not happen, so
> neither `2026-07-27-phase-checkpoint.md` nor `2026-07-27-zos-access-and-toolchain.md`
> had been delivered. Every question in them was still open as of this draft.
>
> *(Original pre-send review notes, preserved for the record:)*
>
> **This is the short version, by owner request (2026-07-30).** The long-form detail was
> not deleted — it lives in the two 2026-07-27 briefings and in `docs/evidence/`, which
> this email points at. If Jürgen wants depth, the depth is one click away and does not
> need to be in his inbox.
>
> **Before sending, check:** (1) does Jürgen already know the meeting slipped, or should the
> opening acknowledge it differently; (2) are you happy naming Lucas; (3) do you want to
> include the repository link.

---

**To:** Jürgen Holtz (IBM)
**From:** Mert Efe Şensoy
**Subject:** zbridge update — the WTO parameter list is working, and two questions I'd value your view on

---

Hi Jürgen,

I'm sorry we couldn't make the 27th work. Rather than let things sit until the next slot,
here's a short written update — there's a good result to share, and two questions that have
become genuinely blocking.

**The good news.** Our Go code's bytes have been accepted by a real `SVC 35`, and the message
appeared on an operator console. That was MVS 3.8j under Hercules, so it isn't a z/OS result
and I won't present it as one — but it means four of the six implementation steps in my Phase
3b plan are now validated before I've used a minute of mainframe time. The parameter list is
no longer guesswork.

**Hercules deserves the credit here.** The pointer you passed on has turned out to be the most
useful thing in the project so far. What made it work was narrowing the question: instead of
asking *"can we get a mainframe?"*, I asked which unknowns could be retired without one. Two of
three fell — whether our assembly emits correct s390x (QEMU answered that), and whether the WTO
parameter list is byte-correct (MVS answered that). Only the genuinely z/OS-specific parts are
left.

**What it couldn't do, honestly.** It can't run z/OS. I looked into whether this project's IBM
backing covered a z/OS guest, found that it didn't, and withdrew the assumption the next day —
nothing had been published in the meantime, because I'd flagged it as unverified when I wrote
it. TK5's assembler is S/370 only, so no 64-bit instruction validation there. And I won't quote
any performance figure from it: both Hercules and QEMU implement `TR` as a software loop, so
timing it would measure the emulator rather than the machine. That benchmark has moved to the
hardware phase.

**The workaround** was to run the whole thing headlessly — TK5's own bundled Linux Hercules
inside WSL2, JCL in through the socket card reader, output back as files, no 3270 terminal and
no operator at any point. It cost a few days of fighting the emulator (unclean shutdowns,
overlapping instances corrupting disk state, a JCL parameter that silently stopped producing an
object deck), but it's all scripted now and reproducible, and the failures are written up as
rules so we don't repeat them.

**One correction worth your attention.** My roadmap described the WTO parameter list as a 2-byte
length header followed by EBCDIC text. **There's a 2-byte MCS flags field between them**, which
I only found by reading the macro expansion. Building it as I'd written it would have produced
something broken from byte 2 onward — and I'd have discovered that on borrowed machine time. I
also read GC28-0683-2 page by page and it doesn't document the layout at all, which genuinely
surprised me. Two smaller corrections are written up in the repository.

**A small organisational note:** the project has grown from just me into a five-person team,
with me leading it — four other computer engineering students at TEDU. Nothing changes on your
side: same single point of contact, same questions, same cadence. I mention it only so the
repository makes sense if you look at it.

**What I'd value your view on**, roughly in order of how much it changes what I build next:

1. **How would I go about obtaining and building IBM's Go fork for z/OS?** Upstream Go has no
   `zos/s390x` target at all, so this is now the critical dependency — and it's a separate
   question from access to a system. Related: since Go's s390x assembler has no `SVC` mnemonic,
   I'm emitting it as literal bytes, the way `x/sys/unix` and go-recordio both do. Is that the
   accepted approach, or is there an IBM-side convention I should follow?
2. **What's a realistic path and timeline for z/OS access?** ZD&T, Wazi aaS, a shared LPAR, the
   more open environment Lucas mentioned — I honestly don't know which of these is reachable
   for a thesis project.
3. **Is there a z/OS publication that documents the WTO parameter list layout — and does a
   single-line WTO on z/OS return anything in R15?** MVS returns nothing, which would make one
   step of my plan a non-step. If the honest answer on the layout is "read the macro
   expansion", that's genuinely useful and I'll stop looking for a manual.
4. **Would you or Lucas still be up for reviewing the go-recordio walkthrough when it ships?**
   That's what I'm working on now.

No rush on any of this, and I'm happy to talk it through on a call if that's easier than
writing it out.

Thanks as always for the guidance — the Hercules suggestion in particular has paid for itself
several times over.

Best regards,
Mert

---

### If you'd like the detail

| | |
|---|---|
| The parameter list, raw `SVC 35`, and our Go bytes accepted | `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` |
| The system coming up, and everything that went wrong after | `docs/evidence/E0-tk5-boot-2026-07-26.md` |
| The s390x port, 29 tests under QEMU | `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` |
| Why Hercules, and what it is not allowed to claim | `docs/decisions/0001-emulation-strategy-hercules-two-track.md` |
| The three roadmap corrections in full | `docs/decisions/0004-roadmap-corrections-and-cgo-scope-closure.md` |
| The plan for the coming academic year | `docs/roadmap-2026-27.md` |

Every emulated result carries a header naming the machine, guest OS, architecture and emulator
version. Nothing is presented as a z/OS result until it runs on z/OS.

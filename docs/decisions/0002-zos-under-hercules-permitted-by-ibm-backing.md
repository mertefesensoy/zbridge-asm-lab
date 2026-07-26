# ADR 0002 · z/OS under Hercules is permitted for this project; ADR 0001 §1 is superseded

- Status: **WITHDRAWN — 2026-07-26.** The premise did not survive the owner's own
  research. ADR 0001 §1 is **restored and governs again**: z/OS under Hercules is out.
  This was not an override; it is reopen-condition #1 of this ADR firing exactly as
  written. **See §10 (Withdrawal) at the end before reading anything above as current.**
- Date: 2026-07-25 (accepted), 2026-07-26 (withdrawn)
- Decided by: **Mert Efe Şensoy (owner)**, 2026-07-25, in session
- Supersedes: **ADR 0001 §1** in full. Amends ADR 0001 §3 (the U3 row), §5, and
  reopen-condition #4.
- Evidentiary status of the central premise: **OWNER-ASSERTED — citation pending.**
  Read §2 before relying on this ADR for anything external.

---

## 1. What changed

ADR 0001 §1 read:

> **z/OS under Hercules is ruled out, permanently and on the record.** Not deferred,
> not "pending a license" — excluded. […] The project's z/OS work happens on real,
> entitled z/OS access or it does not happen.

The owner has determined that this project is **officially backed and supported by
IBM**, and that this backing makes running z/OS under Hercules permitted for it.
On that basis §1 is superseded. Hercules is a mandated tool of this project, and the
z/OS guest is no longer excluded from it.

**Note what did *not* change, because it was never in dispute.** ADR 0001 never
treated Hercules itself as legally doubtful. It records Hercules as OSI-certified
open source under the Q Public License, adopts it as the project's two-track
laboratory, and makes Track M — MVS 3.8j via TK5, which runs *on* Hercules — the
**primary** track that "starts immediately." The exclusion was always about the
licensing of the **guest operating system**, not about the emulator. Any reading of
ADR 0001 as hostile to Hercules is a misreading, and this ADR is the place that says
so on the record.

## 2. The evidentiary status of the premise — read this before citing this ADR

Goal-prompt §4.1 permits exactly three kinds of claim: **cited**, **evidenced**, or
**registered** as an explicit open assumption. There is no fourth category.

The premise "IBM's backing of this project permits running z/OS under Hercules" is
**registered, not cited.** It rests on the owner's statement in session on
2026-07-25. No agreement, entitlement document, licence text, or written statement
from IBM or from the mentor has been produced, read, or filed in this repository.

This is not a challenge to the owner's decision — the owner is the decision authority
for this project and has made the call, which this ADR records and acts on. It is the
standard this project holds *itself* to, and it exists because of a documented
failure earlier the same day: research brief 001 returned a claim with a plausible
form number attached, and the manual turned out not to contain the quotation at all
(`docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md`). The lesson recorded
from that was **"verify the form number, then verify the page."** A licensing premise
deserves the same treatment.

**Citation slot — to be filled by the owner:**

```
IBM backing / entitlement basis
  Instrument:      ____________________  (agreement, programme entitlement, written
                                          statement from IBM or from the mentor)
  Identifier:      ____________________  (agreement number, programme name, date)
  Scope:           ____________________  (which products, which environments, which
                                          period, which named individuals)
  Filed at:        ____________________  (path in docs/evidence/ or docs/decisions/)
  Verified on:     ____________________
```

Until that block is filled, the following restrictions apply and are not optional:

1. **Nothing in the thesis, the public repository, a talk, a post, or any
   mentor-facing document may state that IBM permits z/OS under Hercules.** Publishing
   an unsourced licensing claim attributed to IBM is the single most damaging thing
   this project could do to its own credibility, and it is also a claim about a third
   party's legal position.
2. Evidence files produced from a z/OS-under-Hercules guest carry the provenance
   header unchanged (ADR 0001 §7 survives in full) **plus** a line reading
   `entitlement: OWNER-ASSERTED, uncited (ADR 0002 §2)`.
3. This ADR is the only document in the repository that asserts the premise. Other
   documents link here rather than restating it.

## 3. What this unlocks, precisely

The U3 row of ADR 0001 §3 read "**No. Nothing emulates this.**" That is now partly
wrong, and partly still right. The corrected decomposition:

| Unknown | Statement | Retirable under Hercules-hosted z/OS? |
|---|---|---|
| **U1** | Correct s390x emission; Go ABI on big-endian 64-bit | Already retired — `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` |
| **U2** | WTO parameter list byte-correctness; `SVC 35` acceptance; return contract | **Yes, and now directly** — against z/OS's own `SVC 35` rather than MVS 3.8j's, which removes the entire ancestor-divergence question H001 exists to manage |
| **U3a** | `Malloc31` / below-the-bar allocation, AMODE 31↔64 switching, extended WPL, USS execution | **Yes** — these are z/OS behaviours, and a z/OS guest exhibits them |
| **U3b** | **The `GOOS=zos` toolchain** | **No. Unchanged.** A licence does not produce a compiler |

**U3b is the finding that matters, and it is now the critical path.** Verified today
against the local toolchain: `go tool dist list` (go1.26.3) offers `linux/s390x` and
**no `zos/s390x`**. The string `"zos"` appears in `internal/syslist`, so `_zos.go`
filename constraints parse, but `zos/s390x` is absent from `internal/platform`'s
supported list. The z/OS files under `GOROOT/src/cmd/vendor/golang.org/x/sys/unix/`
are vendored library sources, not toolchain support.

So the sequence of blockers has reordered rather than cleared:

- **Before this ADR:** blocked on z/OS access, which was blocked on licensing.
- **After this ADR:** the licensing blocker is removed by owner decision; the binding
  constraint becomes **obtaining and building IBM's Go fork for z/OS**, plus obtaining
  a z/OS installation image, plus getting it to IPL under Hercules.

That reordering is the real content of this decision. It converts a legal question
into three engineering questions, which is a very large improvement — engineering
questions have schedules.

## 4. The engineering reality, stated separately from the licensing question

These are true regardless of what any licence permits, and none of them is an
argument against the decision. They are here so the plan is built on the actual
difficulty rather than on the removal of the legal one.

1. **Hercules 3.13 is not the right build.** The Hercules installed in this project's
   WSL environment on 2026-07-25 is 3.13 (Ubuntu package, dated 2015). Modern z/OS
   needs current z/Architecture facility support; the actively maintained fork is
   **SDL Hercules 4.x "Hyperion"**, which ADR 0001 §Evidence already identifies. TK5
   bundles SDL 4.9.1. Any z/OS attempt uses Hyperion, not the distribution package.
2. **A z/OS guest is a large systems-programming exercise, not a download.** DASD
   volume sizing, IODF and device definitions, IPL parameters, and missing hardware
   facilities (notably crypto) are all real work, and community reports of z/OS under
   Hercules describe substantial effort even where the image is in hand. Budget it as
   a multi-session project with a real chance of not completing.
3. **The image is a separate question from the permission.** Being permitted to run
   z/OS does not produce installation media. Where the image comes from is an open
   item for the owner, and it is a distinct question from the entitlement in §2.
4. **The mentor is the right reviewer for all of this.** Jürgen Holtz is an IBM
   employee mentoring this project; the entitlement question, the image question, and
   the ZD&T / Wazi aaS alternatives are all better answered by him than inferred here.

## 5. What happens to Track M and the E-ladder

**Track M continues unchanged and remains the immediate next step.** The owner
selected rung **E0 via TK5 on Windows** as the next Hercules milestone, and nothing
in this ADR displaces it:

- TK5 is a bounded download-and-run that works today, with no image question and no
  entitlement question attached.
- The E-ladder rungs E0→E3 were designed to retire U2 cheaply, and they still do.
- Operational fluency with Hercules — IPL, console, JCL submission, output retrieval —
  is a prerequisite for a z/OS guest attempt, not an alternative to it. Learning it on
  TK5 costs a fraction of learning it on z/OS.

The E-ladder gains rungs above E3 rather than losing the ones below it:

| Rung | Gate | Guest | Retires |
|---|---|---|---|
| E0–E3 | unchanged (ADR 0001 §6) | MVS 3.8j / TK5 | operational fluency, U2 via the ancestor |
| **E4** | z/OS IPLs under SDL Hyperion; console readable; job submits | z/OS | the guest-installation question |
| **E5** | `WTO` reaches the z/OS console from a hand-built parameter list and raw `SVC 35` | z/OS | **U2 directly** — H001's ancestor-divergence question becomes moot rather than answered |
| **E6** | A Go binary built with IBM's fork runs on the z/OS guest | z/OS | **U3b**, and it gates everything above it |

E4 and E5 are pre-registered as gated on §2's citation slot being filled before any
result from them is published.

## 6. Consequences for H001

H001 asks whether MVS 3.8j's `SVC 35` is a valid oracle for the z/OS WTO parameter
list. If rung E5 becomes reachable, **H001 stops being load-bearing**: the question
"does the ancestor tell us about the descendant" is replaced by asking the descendant.

H001 is **not** retired or withdrawn on that basis, for two reasons. It is not yet
known whether E4/E5 will actually be reached, and H001's open C4 finding — that
GC28-0683-2 p.210 says a single-line WTO issues no return code — is an independent
documentary result that stands on its own and still bears on ADR 0001 §6's claim
about Phase 3b step 6. That question is unaffected by this ADR and remains open.

## 7. What this decision does not claim

- **It does not claim IBM has licensed z/OS under Hercules generally**, for anyone
  else, or outside this project. It records a determination made by the owner about
  this project. See §2.
- **It does not claim a z/OS guest will work.** §4 lists concrete reasons it may not.
- **It does not weaken ADR 0001 §7.** Every provenance rule survives intact, and §2
  adds a field rather than removing any. An emulated z/OS result is still an emulated
  result and must be labelled as one — a z/OS guest under Hercules is *not* real
  hardware, and the T-ladder's purpose is unchanged.
- **It does not retire U3.** It splits U3 into U3a (retirable by a z/OS guest) and
  U3b (the toolchain, retirable by nothing but IBM's Go fork).
- **It does not change the thesis.** The contribution is still a pure-Go-assembly WTO
  with no cgo. An easier path to a test environment does not alter what is being
  built or why it is novel.

## 8. What would reopen or reverse this decision

1. **The citation slot in §2 cannot be filled.** If no instrument, entitlement, or
   written statement can be produced, this ADR is reduced to a record of an
   unsubstantiated premise and §1's supersession is withdrawn. The restrictions in §2
   are designed so that nothing published depends on the premise in the meantime.
2. **The mentor contradicts the premise.** Jürgen Holtz's position governs; bring it
   here rather than working around it.
3. **ZD&T or Wazi aaS becomes available.** IBM's own entitled emulated-z/OS products
   are a cleaner path than Hercules for the same goal, and they come with the
   entitlement question already answered. If either is offered, it supersedes this
   ADR's z/OS-under-Hercules route (this is ADR 0001's original reopen-condition #4,
   which survives).
4. **A z/OS guest attempt fails twice under SDL Hyperion.** Then the route is recorded
   as a negative result about emulation, Track M continues on TK5, and the T-ladder
   reverts to being the only path to U3a.

## 10. Withdrawal — 2026-07-26

**This ADR is withdrawn. ADR 0001 §1 is restored and governs: z/OS under Hercules is
ruled out.** Everything in §1 through §9 above is retained unchanged as the record of a
decision that was made, tested against reality, and reversed.

### What happened

The owner researched the entitlement question and reported on 2026-07-26:

> *"For the license problem I have done some research but unfortunately we do not have
> that kind of agreement or a special permission for us to use."*

That is §8's **reopen-condition #1** firing precisely as written: *"The citation slot in
§2 cannot be filled. If no instrument, entitlement, or written statement can be
produced, this ADR is reduced to a record of an unsubstantiated premise and §1's
supersession is withdrawn."*

So this is not an override of a decision and it is not a mistake being cleaned up after.
It is the decision working. The ADR named its own failure condition in advance, the
condition occurred, and the documented consequence was applied.

### The cost of having been wrong for one day: nothing

This is worth stating plainly, because it is the strongest argument for the way this
project handles claims.

§2 required the premise to be treated as **registered, not cited**, and imposed three
restrictions while the citation slot was empty:

1. Nothing in the thesis, the repository, a talk, a post, or any mentor-facing document
   may state that IBM permits z/OS under Hercules.
2. Evidence from a z/OS guest must carry `entitlement: OWNER-ASSERTED, uncited`.
3. This ADR is the only document that asserts the premise; others link to it.

**All three held.** No z/OS guest was ever created, so restriction 2 was never
exercised. No document outside this ADR asserted the premise, so nothing else needs
correcting. Nothing was published. The withdrawal touches this file, four
status/pointer documents, and no technical work whatsoever.

Compare the counterfactual: had the premise been written into the ADR as a cited fact,
it would by now be in the goal prompt, the handover, the roadmap narrative, and
plausibly in something shown to a mentor at IBM — and the correction would have been a
retraction of a licensing claim about IBM rather than an internal status change.

### What this does NOT cost the project

**The critical path is untouched.** ADR 0002 was an accelerant, never a dependency:

- **Track M was always the primary track and is entirely unaffected.** MVS 3.8j is
  freely runnable under Hercules; that has never been in question. Rungs E0–E3 stand
  exactly as ADR 0001 §6 defines them.
- **The WPL layout — the project's actual blocker — was never going to be solved by a
  z/OS licence anyway.** The runbook's §11 already identifies the real source:
  *"E1's assembler listing shows the exact parameter-list bytes IBM's own macro
  generates, which is the ground truth this project needs and which no manual can be
  fully trusted for."* That listing comes off MVS 3.8j. No entitlement required.
- **U1 is already retired** on QEMU, and QEMU needs no licence.
- **The proposed rungs E4–E6 are withdrawn** along with §5's table. E0–E3 are restored
  as the whole of the E-ladder.

### What is restored

| Item | Restored to |
|---|---|
| ADR 0001 §1 | In force: z/OS under Hercules excluded; z/OS work happens on entitled access or not at all |
| ADR 0001 §3, U3 row | **"No. Nothing emulates this."** The U3a/U3b split proposed in §3 above is withdrawn |
| ADR 0001 §6 | E-ladder is E0–E3 only |
| ADR 0001 reopen-condition #4 | In force again, and it is now the live question: *"If IBM offers a legally entitled emulated z/OS for students or ambassadors, it supersedes Track M for U2 entirely"* |
| H001 | **Fully load-bearing again.** §6 above speculated that reaching a z/OS guest would make the ancestor-divergence question moot. It will not be reached, so H001 is the mechanism for U2 and its C4 finding still stands |

### The one finding that survives the withdrawal

**U3b — that upstream Go cannot target z/OS — is unaffected and remains true.**
`go tool dist list` (go1.26.3) offers `linux/s390x` and no `zos/s390x`; `"zos"` is in
`internal/syslist` but the target is absent from `internal/platform`. That was verified
against the toolchain, not inferred from the licensing premise, so it stands on its own
evidence (`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`, finding F3).

It keeps its practical consequence: **obtaining IBM's Go fork is a hard dependency of
the endgame, and it is a separate question from the z/OS licence.** A compiler that
emits z/OS binaries is not the same entitlement as a z/OS system to run them on, and it
is worth asking the mentor about both.

### Open, and going to the mentor

The owner consults Jürgen Holtz on Monday 2026-07-27. The questions are collected in
`docs/mentor-briefings/2026-07-27-zos-access-and-toolchain.md`. This ADR should not be
revived on a verbal "that's probably fine" — reviving it requires the §2 citation slot
filled with an actual instrument, which is the same bar it failed today.

## 11. Links

- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — §1 superseded here;
  §2–§7 stand, and §7's provenance rules are reaffirmed
- `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` — the U1 retirement, and the
  source of the `zos/s390x` toolchain finding (F3)
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — becomes non-load-bearing if E5 is
  reached; its C4 finding is unaffected
- `docs/runbooks/tk5-hercules-setup.md` — rung E0, the owner's chosen next milestone
- SDL Hercules 4.x Hyperion: <https://github.com/SDL-Hercules-390/hyperion>

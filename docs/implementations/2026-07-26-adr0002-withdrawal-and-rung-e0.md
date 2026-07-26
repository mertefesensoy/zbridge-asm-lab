# 2026-07-26 · ADR 0002 withdrawn, mentor briefing prepared, and rung E0 passed headlessly

**Scope:** three things — retract a decision whose premise failed, prepare the owner's
mentor consultation, and pass the lowest unpassed rung on the track that never depended
on the retracted premise.

---

## 1. Problem / motivation

On 2026-07-25 the owner determined that this project's IBM backing permitted running
z/OS under Hercules, and ADR 0002 recorded that, superseding ADR 0001 §1.

On 2026-07-26 the owner reported:

> *"For the license problem I have done some research but unfortunately we do not have
> that kind of agreement or a special permission for us to use. I will consult this to
> my mentor tomorrow Monday but until then we need to defer it and make progress so
> that I can tell my mentor how I am going around that problem."*

Three needs follow: retract the ADR correctly, give the owner something to take to the
mentor, and demonstrate that the project is not blocked — by actually advancing it.

## 2. What changed

| File | Change |
|---|---|
| `docs/decisions/0002-…-ibm-backing.md` | Status → **WITHDRAWN**; new §10 recording the withdrawal, what it cost, and what is restored. §1–§9 retained unchanged as the record of a decision that was made and reversed |
| `docs/mentor-briefings/2026-07-27-zos-access-and-toolchain.md` | **New.** The owner's briefing for the Monday consultation: licensing position, what shipped without it, five questions for the mentor |
| `docs/evidence/E0-tk5-boot-2026-07-26.md` | **New.** Rung E0 evidence with provenance header, all three gates, every ⚠ VERIFY item answered, and an operational-lessons section recording three unclean stops |
| `docs/runbooks/tk5-hercules-setup.md` | Header status → E0 PASSED with a pointer to §12; **new §12** with the headless WSL2 procedure, two new traps, and the verified device/port table |
| `docs/goal-prompt.md` | Directive 4 corrected; U3 unsplit; current-position block rewritten for E0 and the withdrawal |
| `CLAUDE.md` | Hercules directive corrected — Hercules mandated, z/OS guest ruled out |
| `memory/MEMORY.md` | E0 passed; ADR 0002 withdrawn; E1 identified as the next and most important rung |

No source code changed. No test changed. That is the point of §3.1.

## 3. Implementation approach

### 3.1 The withdrawal was pre-registered, so it was cheap

ADR 0002 §8 reopen-condition #1 read: *"The citation slot in §2 cannot be filled. If no
instrument, entitlement, or written statement can be produced, this ADR is reduced to a
record of an unsubstantiated premise and §1's supersession is withdrawn."*

That condition occurred, so the documented consequence was applied. This is not a
decision being overridden; it is a decision working.

The reason it cost nothing is §2's three restrictions, imposed while the citation slot
was empty: no published surface may state the premise, evidence from a z/OS guest must
be labelled `OWNER-ASSERTED, uncited`, and this ADR is the only document that asserts
it. **All three held.** No z/OS guest was created, nothing was published, no other
document restated the premise. The withdrawal touched one ADR and four pointer
documents and **zero lines of technical work**.

The counterfactual is the argument for the discipline: had the premise been written in
as a cited fact, it would be in the goal prompt, the handover and the roadmap narrative
by now, and the correction would have been a retraction of a licensing claim about IBM
in front of an IBM mentor.

**One finding survives the withdrawal**, because it was verified against the toolchain
rather than inferred from the licence: upstream Go has no `zos/s390x` target. IBM's Go
fork is therefore a hard dependency, and it is a *separate* entitlement question from a
z/OS system. Asking about only one of them was the trap to avoid.

### 3.2 Choosing what to advance

The project's own priority rule (goal-prompt §Action-selection) is: unsynthesised
research return → unclosed hypothesis with evidence → **lowest unpassed rung** → laptop
prep → gap in the record. The lowest unpassed rung was **E0**, on Track M, which runs on
MVS 3.8j and has never depended on a z/OS entitlement.

Better still, runbook §11 already identified that **E1's assembler listing is the
ground-truth source for the WTO parameter-list byte layout** — the project's actual
blocker — and that it needs no entitlement. So the licensing deferral costs nothing on
the critical path, and E0 is the step that opens it.

### 3.3 E0 without a terminal or an operator

The runbook describes a Windows, interactive, 3270-terminal path. Reading the TK5
package changed the plan:

1. **TK5 ships Linux startup scripts and a Linux x86-64 Hercules binary**
   (`hercules/linux/64/bin/hercules`, SDL Hyperion 4.9.1.0-SDL). So it runs inside WSL2
   on ext4 — which eliminates runbook Traps 1, 2 and 3 outright (OneDrive DASD
   corruption, non-ASCII paths, antivirus scanning). `/mnt/c` was deliberately avoided:
   9p DASD I/O is slow and would have reintroduced all three.
2. **`scripts/ipl.rc` arms HAO (Hercules Automatic Operator) against `IEA101A` and
   `IEA305A` before issuing the IPL.** So TK5 IPLs fully unattended — answering the
   runbook's §3 ⚠ VERIFY item, which had flagged the `IEA101A` prompt as a possible hang.
3. **Device `000C` is `3505 ${RDRPORT:=3505} sockdev`** — a card reader on a TCP socket.
   JCL pushed at port 3505 enters the JES2 queue as if punched. Printers write to
   *files*. So the entire ladder needs no 3270 client: JCL → 3505 → JES2 →
   `prt/prt00e.txt`.

The gate job was `IEFBR14`, the canonical MVS no-op — it allocates nothing and returns 0,
so a pass is unambiguously about the system rather than the job.

### 3.4 §8's automation was not built after E0 — it was E0

The runbook says to build the socket-reader pipeline only once E0 has passed by hand,
reasoning that automating an unrun pipeline means debugging two things at once. In
practice E0 was passed *by* that pipeline, which leaves it proven end to end. What
remains for a `Submit-MvsJob` wrapper is packaging, not discovery.

## 4. Verification

Every command was run. Raw output is in the evidence file.

- **Gate 1 (IPL):** `HHC00811I Processor CP00: architecture mode S/370`,
  `/IEA101A SPECIFY SYSTEM PARAMETERS…` appearing and being auto-answered,
  `$HASP493 JES2 QUICK-START IS IN PROGRESS`, TK5 Update 5 banner.
- **Gate 2 (console):** MVS answered `/d t` with
  `IEE136I LOCAL: TIME=10.29.54 DATE=2026.207`.
- **Gate 3 (job in, output out):** `IEF403I ZBE0T01 - STARTED`,
  `ZBE0T01 STEP1 IEFBR14 RC= 0000`, `IEF142I … COND CODE 0000`, `$HASP395 … ENDED`,
  and the full printed listing retrieved from `prt/prt00e.txt` with JES2 statistics
  (7 cards read, 21 SYSOUT print records).
- **Clean shutdown:** `HHC01422I Configuration released`,
  `HHC01424I All termination routines complete`, `HHC01425I Hercules shutdown complete`,
  `HHC01412I Hercules terminated`; all 16 DASD volumes flushed.
- **Restart:** a second IPL at 12:42:38 reached JES2 quick-start and answered `/d t`.

Package integrity: `mvs-tk5.zip`, 498,312,872 bytes, SHA-256
`710d002843631322810a276dd42c793fda458548dc64d86e2914a62db7425f84`, `unzip -t` clean.

## 5. What went wrong, and what it changed

**Three unclean stops occurred after the gate passed.** They are recorded rather than
tidied away, because the corrections are now rules.

1. **A detached Hercules died on stdin EOF.** Hercules reads console commands from
   stdin; the launcher held stdin open with a `sleep infinity` writer into a FIFO, and
   when that helper was reaped with its process group Hercules saw EOF and exited
   *without running MVS's shutdown*. Runbook Trap 5 by an indirect route.
2. **Two instances overlapped.** A leftover detached instance was still live when a later
   script replaced the `/root/mvs-tk5` tree. Two emulators on one set of CCKD volumes
   makes the state untrustworthy however either stopped.
3. **The wrong success signal.** A script inferred a clean stop from *process absence*.
   Process absence proves nothing — it is equally consistent with a kill mid-write.

Rules adopted, now in runbook §12 as Traps 6 and 7:

- `HHC01412I Hercules terminated` is the **only** accepted proof of a clean stop.
- Release the stdin holder **only after** that message appears.
- Assert exactly one instance (`pgrep -cf 'hercules -f conf/tk5.cnf'` = 1) before
  touching the tree; never modify it while an instance is live.
- Leave the system **down** between sessions — WSL2 idles its VM out, and an idle-out
  with buffered DASD writes is Trap 5 again. Keep nothing in `/tmp` (tmpfs).

**Resting state:** `/root/mvs-tk5` was re-extracted from the verified zip and left
**never booted**. A system that has never IPLed cannot have been damaged by an unclean
stop, which makes it the only resting state that needs no argument.

**Total cost of the episode: a three-minute re-extract**, because the zip was kept and
its hash recorded. That is the first genuinely useful test of the runbook's own advice.

## 6. Design decisions

| Decision | Alternatives | Why |
|---|---|---|
| Append a §10 withdrawal rather than delete ADR 0002 | Delete it; rewrite §1–§9 | An ADR is a record, not a status field. Keeping the reversed reasoning visible is what makes the next premise get tested too |
| Run TK5 in WSL2 on ext4 | Windows `C:\mvs-tk5` per the runbook; WSL on `/mnt/c` | Eliminates three of five traps. `/mnt/c` would have reintroduced them and been slow |
| Pass E0 headlessly via the socket reader | Wait for the owner at a 3270 terminal | The scarcity principle applies to owner time too, and E1's listing can now be produced unattended |
| `IEFBR14` as the gate job | A TK5 sample job (runbook §7's suggestion) | TK5's samples live in datasets, not `jcl/`, which would have needed a TSO logon. `IEFBR14` isolates "system works" from "job works" |
| Re-extract rather than trust a snapshot | Keep the post-episode snapshot; run integrity tools | The snapshot was taken from a state that may already have been contaminated. A never-booted extract needs no argument |

## 7. What was NOT done

- **E1 was not attempted.** No assembler was run and no `WTO` macro was expanded.
- **A repeated clean shutdown is not evidenced.** One is; the later attempts ended
  uncleanly. Runbook §9's checklist item is partially met and will be closed at the
  start of E1 using the corrected procedure.
- **No TSO logon.** So the runbook's §5 credential ⚠ VERIFY item is still open.
- **No WPL layout was written.** Unchanged: `console/wpl.go` still returns
  `ErrLayoutUnverified`. E1 is what changes that, not this session.
- **H002 claim C3 is still untested.** No Hercules-hosted Linux s390x guest exists; the
  Hercules now in use is TK5's, running MVS 3.8j on S/370.
- **No performance figure was taken**, here or anywhere, and none should be.

## 8. Related docs

- `docs/evidence/E0-tk5-boot-2026-07-26.md`
- `docs/mentor-briefings/2026-07-27-zos-access-and-toolchain.md`
- `docs/decisions/0002-…-ibm-backing.md` §10 — the withdrawal
- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — restored in full
- `docs/runbooks/tk5-hercules-setup.md` §12 — the procedure that ran
- `docs/implementations/2026-07-25-s390x-port-and-bridge-scaffold.md` — the prior session

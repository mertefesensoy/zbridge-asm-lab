# `mvsjob.sh down`: shut MVS down first, and make the clean-stop proof unforgeable

**Date:** 2026-10-21
**Author:** Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

On 2026-10-21, `mvsjob.sh down` hung at the end of the IEZWPL hunt, on a fresh,
disposable TK5 tree. It sent `script scripts/poweroff`. Hercules began its shutdown,
closed every disk, and then stopped making progress: `HHC01424I All termination routines
complete` never appeared. The harness gave up after its 6-minute wait. A second `quit`
then printed `HHC01412I Hercules terminated`, but the process stayed alive. SIGTERM
printed `HHC01412I` once more, and still nothing exited. Only SIGKILL ended it. On
2026-08-10 the same power-off path had stopped cleanly twice.

This change answers three questions: which part of the shutdown hangs, how to avoid it,
and what a safe escalation is once the disks are closed. Answering them turned up two
separate defects and a hazard in the docs:

1. **Powering off with MVS still running hangs Hercules.** `scripts/poweroff` is only
   `stopall`, `sysclear` and `quit`, so MVS itself is never shut down. With MVS up,
   Hercules' shutdown deadlocks while it detaches TK5's 2703 dial-in lines. This
   happened in 4 stops out of 4 on 2026-10-21, one of them reproduced on purpose.
2. **The clean-stop test could be satisfied by a forged message.** The old `down` counted
   a stop as clean as soon as `HHC01412I` appeared in the log. But a second `quit`, or a
   SIGTERM, re-enters Hercules' shutdown, skips the termination routines, and still prints
   `HHC01412I` (§3.2). In that state the process does not exit.
3. **RUN.md told the operator to "simply run `down` again"** after a timeout. With the old
   script, that sent a second `quit`. That forged `HHC01412I` would then be reported as
   `CLEAN STOP confirmed`, a false "clean" with no visible warning.

Separately, `~/mvs-tk5` (the owner's install) was untrustworthy, and its CCKD headers
prove it. It was IPL'd on 2026-08-10 at 09:09 and never shut down; its console log
simply stops on 2026-08-11 at 16:40. Six of its volumes still had the CCKD OPENED bit
set: `tk5res.390`, `tk5cat.391` and `work01`–`work04`.

## 2. What Changed

| File | Change |
|---|---|
| `docs/runbooks/mvsjob.sh` | `down` rewritten. It shuts MVS down first with TK5's own steps, each logged with its own timeout. It then powers off once. It accepts a stop only when one complete shutdown is followed by the process exiting. If Hercules hangs anyway, it kills it only after every disk and printer file is closed, then verifies the disks. It has three exit codes and writes a report per stop. Design notes 1 and 5 corrected; 6 and 7 added. |
| `docs/runbooks/tk5-hercules-setup.md` | §12: Trap 7 rule 1 corrected; new Trap 8 (the hang and the forged message); Trap 6 path note; procedure blocks moved from `/root` to `~` with the MVS-first shutdown; VERIFY-table row for §6 shutdown annotated. |
| `RUN.md` | §6.4 rewritten: what `down` does, its three exit codes, and why `quit` must never be typed by hand. §7.2 timing is now ~2 min. §7.4 step 4 no longer suggests leaving MVS running. §8: two rows replaced, one added for the pre-2026-10-21 script. §4.4: the copy step now strips CRLF line endings. Without that, a script copied from a Windows checkout fails in bash; this was pre-existing, and the new §8 row would have led straight to it. |
| `docs/architecture/emulation-harnesses.md` | The clean-stop rule restated with the tightened proof. |
| `docs/architecture/evidence-ladder.md` | The same tightening, as a dated note. |
| `docs/diagrams/wto-call-path.md` | Sequence-diagram note and the paragraph under it restated. |
| `docs/architecture/c4/sequence-state-object-diagrams.md` | §2 state diagram gains a `ForcedStop` state. ShuttingDown and CleanlyStopped redefined. §2.3 no longer implies forcing a stop by hand. |
| `docs/implementations/2026-10-21-mvsjob-down-mvs-first-shutdown.md` | This document. |

Operational changes, outside the repo, all approved by the owner:

- `~/mvs-tk5` moved aside to `~/mvs-tk5.untrusted-2026-08-11`, and a fresh,
  hash-verified extract put in its place, never booted.
- `~/mvsjob.sh` replaced by this version; the old one is kept as
  `~/mvsjob.sh.2026-08-10`.
- The raw lab and validation output copied to the private
  `zbridge-journey-report/evidence/2026-10-21-shutdown-lab/`, with `SHA256SUMS`.
- The disposable trees deleted afterwards.

## 3. Implementation Approach

### 3.1 Diagnosis: which termination routine hangs

There was no stack trace to read. `ptrace_scope=1` and a password-protected `sudo` rule
out gdb inside WSL. So the hang was located from four sources, each of them readable
without privileges.

1. **A thread ledger built from the console log.** Hercules logs every thread start
   (`HHC00100I`) and end (`HHC00101I`). Subtracting the ended threads from the started
   ones gives the threads still alive.
   - Cleared: the HTTP server (`http_server` ended), the 3270 console listener
     (`console_connect` ended), and all CCKD threads.
   - Never ended: `0:0670 communication thread` ended, but 0671 through 068F never did.
   - `HHC01427I` and `HHC01422I Configuration released` never appear. Those are printed
     at the end of `release_config`, the termination routine that detaches every
     device. So the hang is inside `release_config`, at line 0671.
2. **What those devices are.** In TK5's `conf/tk5_default.cnf`, 0670–068F are 2703
   (`commadpt`) dial-in lines: four BSC RJE, four TTY and eight 2741. zbridge uses none
   of them.
3. **`devlist` just before the stop.** In every lab run, all 16 lines showed
   `OP=ENABLE … open busy`. Each one carries an ENABLE that MVS issued, waiting for a
   caller.
4. **`/proc/<pid>/task/*/{comm,wchan}` during the hang.**
   - `quit_thread`, the thread running the shutdown, is in `futex_wait_queue`.
   - The 0671–068F comm threads are idle in `do_select`.
   - Each line's CCW thread (`dev 06xx thrd`) is in `futex_wait_queue`.

The SDL Hercules source fits all of this. `detach_devblk` holds the device lock while it
calls `commadpt_close_device`, and that function can wait on `ipc_halt` or `ipc` with no
timeout. Separately, `sysclear`'s `device_reset` clears the subchannel but never calls
the 2703 halt handler, so the ENABLE stays parked in the CCW thread. The exact lock
cycle cannot be proven without stacks, and this document does not claim one.

### 3.2 Why a second `quit` prints `HHC01412I`

In `hscmisc.c`, `do_shutdown_now()` prints `HHC01420I`, then `HHC01423I`, then calls
`hdl_atexit()`, and then prints `HHC01424I`, `HHC01425I` and `HHC01412I`. In `hdl.c`,
`hdl_atexit()` begins with `if (hdl_shutting) return;`.

So the second shutdown returns from `hdl_atexit()` immediately, while the first one is
still stuck inside it, and it prints all three "complete" messages. Then it calls
`exit(0)`, which blocks behind the wedged first shutdown. SIGTERM takes the same path.
That is exactly what the hunt log shows: two extra `HHC01420I … HHC01412I` blocks, and no
`HHC01422I` anywhere.

### 3.3 The new `down`: contract

**Inputs.** The same environment as before: `TK5_HOME`, `TK5_FIFO` and `TK5_OUT`, with
the same `$HOME`-based defaults.

**Outputs.**

| Exit | Meaning | Files touched |
|---|---|---|
| 0 | Clean: lines logged after the power-off hold exactly one `HHC01420I`, plus `HHC01422I`, `HHC01424I` and `HHC01412I`, and the Hercules process has exited | report `$OUT/down-<timestamp>.txt` |
| 2 | Forced, not clean: Hercules hung, was SIGKILLed after no disk, printer or punch file was open, and every CCKD volume then passed the check | report; one `MVSJOB: … FORCED STOP, NOT CLEAN …` line appended to `log/3033.log` after the process is dead |
| 1 | Not verified: files were still open (nothing killed, stdin still held); or Hercules exited without one complete sequence; or the disk check failed | report; the log line too, if a kill happened |

**Stages.**

1. **Identify the instance.** Exactly one Hercules may be running, and it must be this
   tree's: `/proc/<pid>/cwd` must equal `$TK5_HOME`. The stdin holder is identified as
   the `sleep` whose fd 1 is `$TK5_FIFO`. No `pkill` patterns remain.
2. **If a shutdown has already begun** (`HHC01420I` is in the log, which `up` truncates
   at each IPL), send nothing. Watch, and escalate if needed. This is what makes
   re-running `down` safe.
3. **Otherwise, shut MVS down.** If MVS is not yet in a disabled wait, `down` first waits
   (240 s at most) for `TK5 system initialization complete`, because `up` returns about
   80 s earlier than that. Then it runs four steps. Each issues a command, waits for a
   fixed success string, and fails fast on a fixed rejection string where one has been
   observed:

   | Command | Wait for | Timeout | Fail fast on |
   |---|---|---|---|
   | `/f bsppilot,shutnow` | `$HASP099` | 300 s | `IEE341I` |
   | `/$pjes2` | `$HASP085` | 120 s | (none) |
   | `/z eod` | `IEE334I` | 60 s | (none) |
   | `/quiesce` | `HHC00809I` | 60 s | (none) |

   If a step fails, `/d a,l` output goes into the report and `down` powers off anyway,
   saying that a hang may follow.
4. **Power off once.** `script scripts/poweroff`, then watch up to 180 s for the process
   to exit.
5. **If it is still alive,** write a snapshot to the report: every thread's state, wait
   channel and name; open files; and the log tail. Then:
   - If any file under `dasd/`, `prt/` or `pch/`, or `log/hardcopy.log`, is still open,
     return 1 and kill nothing.
   - Otherwise SIGKILL. Confirm death, and release the holder (only now: note 2). Verify
     every `CKD_C370` volume: OPENED bit clear, and `cckdcdsk -ro -3` rc=0 on all. Append
     the `MVSJOB:` line, and return 2, or 1 if verification failed.
6. **Never** send a second `quit` or a SIGTERM.

**Invariants.**
- The holder is released only after the Hercules process has exited.
- No signal ever goes to a process outside this tree.
- Nothing is killed while a disk or printer file is open.

`up`, `run` and `cmd` are unchanged.

## 4. Mathematical / Statistical Details

**The OPENED bit.** A CCKD file starts with a 512-byte CKD device header, followed by
the compressed-device header `CCKD_DEVHDR`. Its options byte `cdh_opts` is at offset 3
of that header, so at file offset 512 + 3 = **515**.

`cckd.h` defines `CCKD_OPT_OPENED = 0x80` and `CCKD_OPT_OPENRW = 0x40`. In
`cckddasd.c`:
- the first track write after a harden sets OPENED;
- `cckd_harden()` clears it after it has written the header, the L1 table and the
  free-space chain, which happens at close and also periodically.

So for each volume *v*, write `opts(v)` for the byte at offset 515. Then:

```
OPENED(v) = (opts(v) AND 0x80) ≠ 0
```

`OPENED(v)` false means *v*'s metadata was last written consistently, with no track
write since. The post-kill check requires it to be false for every volume, **and**
`cckdcdsk -ro -3` must return rc=0 for every volume.

**What the run counts do and do not show.**
- Stock power-off with MVS up: 4 hangs in 4 runs (the hunt, 0-1, 0-2, and V2's
  deliberate reproduction).
- The MVS-first stop: clean in all *n* = 7 runs (lab B ×3, V1 ×4).

With zero failures in *n* trials, the "rule of three" gives an approximate 95% upper
confidence bound of 3/*n* on the true failure rate. For *n* = 7 that is about 43%, a
weak bound, so the case does **not** rest on the counts. It rests on the mechanism.
Before every stop, `devlist` showed 16 of 16 lines busy with ENABLE. After the MVS
shutdown, it showed 0 of 16 busy, and the hang requires a busy line. The counts confirm
the mechanism; on their own they would not establish it.

**Worst-case `down` duration** is the sum of its timeouts:
240 + 300 + 120 + 60 + 60 + 180 s = 960 s, which is 16 min. A failed step adds about
5 s for `/d a,l`, and the disk check adds a few seconds. A normal stop takes about 2 min
(§6).

## 5. Design Decisions

- **MVS-first (B) over removing the 2703 lines (A).** Both were 3 clean stops out of 3
  in the lab. A is faster (16 s), and it also closes 16 listening ports. But it changes
  TK5's device configuration, and the project's evidence rests on stock TK5. A is easy
  to reproduce if it is ever wanted: TK5's own hook `INCLUDE
  ${CNF101A:=conf}/tk5_${REP101A:=default}${CMD101A}.cnf` accepts `CNF101A=<dir>`,
  pointing at a copy of `tk5_default.cnf` with lines 0670–068F commented out. No shipped
  file has to change. B is also simply the correct operating procedure: TK5's own
  `scripts/shutdown` runs these same MVS steps first, and the clean 2026-07-26 E0 stop
  went that way.
- **The harness issues the MVS steps itself, rather than calling `scripts/shutdown`.**
  That script chains the steps through HAO triggers. On 2026-08-10 it twice stopped
  partway, and the log of those runs was later overwritten, so which link broke can no
  longer be examined. Running the steps one at a time gives every link a timeout, a
  report line and a `/d a,l` on failure.
- **Fail-fast strings only where observed.** `IEE341I BSPPILOT NOT ACTIVE` appears in
  real TK5 logs. No rejection text for `$PJES2` was observed, so none is guessed; that
  step relies on its timeout.
- **`down` waits for TK5 initialisation; `up` is unchanged.** SHUTNOW was only tested
  after `TK5 system initialization complete`. Making `up` wait for it would add about
  80 s to every demo boot. Waiting in `down` costs time only when `down` comes early,
  and V1 cycle 1 tests exactly that.
- **The clean test is tightened, not replaced.** The owner's rule, "HHC01412I and the
  process exiting", stands. It now also requires a single complete shutdown, so a forged
  `HHC01412I` can never pass.
- **The kill gate checks `/proc/<pid>/fd`, not the log.** An open file descriptor is
  direct evidence; a log message is indirect. Printer, punch and hardcopy files are
  included because they are this project's evidence outputs. `3033.log` is excluded,
  because it is Hercules' own stdout.
- **Exit 2 is "not clean".** The disks were verified, but the project rule only trusts a
  clean stop, so docs say to re-extract before any evidence run. Exit 2 is distinct from
  1, so scripts can tell a verified forced stop from an unverified one.
- **Instances are found by cwd and fd, not by pidfiles or `pkill -f`.** A pidfile can go
  stale. A `pkill` pattern can hit another tree's Hercules (rule 3) or an unrelated
  `sleep`. The cwd and fd checks need no state of their own.
- **Hercules itself is not patched.** The deadlock is in SDL Hercules' `commadpt`, which
  TK5 ships as a binary. Avoiding it is in scope; fixing it upstream is a follow-up.
- **The 2026-08-10 clean power-offs are left unexplained.** The one hypothesis that
  reconciles them is that those power-offs followed `scripts/shutdown` attempts whose
  SHUTNOW had already released the lines. It cannot be checked, because that log was
  overwritten by the next `up`.

## 6. Verification

### 6.1 Lab: which fix works (2026-10-21, 17:23–17:57)

Every run used a fresh extract of `~/tk5dl/mvs-tk5.zip`, sha256
`710d002843631322810a276dd42c793fda458548dc64d86e2914a62db7425f84`, in its own
directory, with one Hercules at a time. `~/mvs-tk5` was never used. Each run did:
`up` → one IEFBR14 job → `devlist` → stop variant. Variants were interleaved:
0, A, B, 0, A, B, A, B.

| Run | Variant | Outcome | Power-off to exit | Termination messages | Disks after |
|---|---|---|---|---|---|
| 0-1* | stock poweroff | hung on 0671; forced kill, no file open | — (killed at 240 s) | 1420 only | 15/15 OPENED clear, `cckdcdsk` rc=0 |
| 0-2 | stock poweroff | hung on 0671; forced kill, no file open | — (killed at 240 s) | 1420 only | 15/15 OPENED clear, `cckdcdsk` rc=0 |
| A-1, A-2, A-3 | lines 0670–068F left out | CLEAN ×3 | 16 s each | 1420, 1422, 1424, 1412 | 15/15 OK |
| B-1, B-2, B-3 | MVS first | CLEAN ×3 | 16 s each | 1420, 1422, 1424, 1412 | 15/15 OK |

\* 0-1 was stopped before TK5 initialisation finished: a driver bug matched the echoed
`hao tgt MVS038J` line. It hung all the same.

In B, the MVS steps took the same time in all three runs:
- SHUTNOW to `$HASP099`: 72 s
- `$PJES2` to `$HASP085`: 6 s
- `Z EOD` to `IEE334I`: 2 s
- `QUIESCE` to `HHC00809I`: about 0 s

`devlist` showed 16 of 16 lines busy before every stop, and 0 of 16 after B's MVS
shutdown.

### 6.2 Validation of the new `down`

Validation ran on 2026-10-21, 18:06–18:26, with this script. The edits made after the
run touched comment lines only, which a `diff` with comment lines excluded confirms. Both parts
used fresh hash-verified extracts, one Hercules at a time, and the project's real demo
job, `zbe3go.jcl`.

**V1: four consecutive `up` → `run` → `down` cycles on ONE tree.** Cycles 2–4 therefore
re-IPL a tree that the new `down` had just stopped. Cycle 1 used RUN.md's demo timing:
`down` straight after `run`, before TK5 had finished initialising.

| Cycle | `up` | MVS steps (SHUTNOW / `$PJES2` / `Z EOD` / `QUIESCE`) | Power-off → exit | `down` total | Exit | Log: 1420 / 1422 / 1424 / 1412 |
|---|---|---|---|---|---|---|
| 1 (demo timing) | 20 s | waited 48 s for TK5 init, then 73 / 6 / 2 / 0 s | 16 s | 145 s | 0 | 1 / 1 / 1 / 1 |
| 2 | 19 s | 38 / 6 / 2 / 0 s | 16 s | 62 s | 0 | 1 / 1 / 1 / 1 |
| 3 | 19 s | 38 / 6 / 2 / 0 s | 16 s | 64 s | 0 | 1 / 1 / 1 / 1 |
| 4 | 19 s | 72 / 7 / 2 / 0 s | 16 s | 97 s | 0 | 1 / 1 / 1 / 1 |

After every cycle:
- no Hercules was running, no stdin holder remained, and the FIFO had been removed;
- every volume had its OPENED bit clear;
- no JES2 operator prompt appeared on any re-IPL.

After cycle 4, all 15 volumes passed `cckdcdsk -ro -3` with rc=0. The message counts
were taken independently from each cycle's saved `3033.log`, not from `down`'s own
output.

The demo job itself succeeded every time, with `RC= 0000` from IFOX00, IEWL and GO in
all four cycles, including cycle 1, submitted before TK5 initialisation finished. Every
run also logged a `$HASP190 ZBE3GO SETUP -- PUNCH1` message. That predates this change,
and it did not stop the job or its listing.

**V2: the escalation path, on a second fresh tree.** After `up` and `run`, the OLD stop
was sent by hand (`script scripts/poweroff` with MVS still up), to reproduce the hang.
Then `down` ran:

- It saw the shutdown already in progress, sent nothing, and watched for 180 s.
- It wrote the thread snapshot to its report.
- It found no disk or printer file open, and sent SIGKILL.
- It verified all 15 volumes, OPENED clear and `cckdcdsk` rc=0, released the holder,
  removed the FIFO, appended the `MVSJOB: … FORCED STOP, NOT CLEAN …` line, and exited
  **2**.

The console log holds exactly one `HHC01420I` and **zero** `HHC01412I`, so nothing was
forged. A second `down` printed `already down (NOTE: this is not proof of a clean stop)`
and exited 0.

**Counting both parts:** the MVS-first stop was clean in 7 runs out of 7 (lab B ×3,
V1 ×4). By the rule of three (§4), that gives an approximate 95% upper bound on the
hang rate of about 43%. That number is why the argument rests on the mechanism and not
on the counts. Stock power-off with MVS up hung in 4 runs out of 4: the hunt, lab 0-1,
lab 0-2, and V2's deliberate reproduction.

### 6.3 How to repeat it

On a fresh tree (never `~/mvs-tk5` itself), in WSL:

```bash
cd ~ && sha256sum tk5dl/mvs-tk5.zip     # must print 710d0028...25f84
mkdir ~/val && unzip -q tk5dl/mvs-tk5.zip -d ~/val && chmod -R +x ~/val/mvs-tk5
export TK5_HOME=~/val/mvs-tk5 TK5_FIFO=~/val/tk5.fifo TK5_OUT=~/val/out
./mvsjob.sh up && ./mvsjob.sh run zbe3go.jcl && ./mvsjob.sh down; echo "down rc=$?"
```

This passes when `down` exits 0 and its last line reads `CLEAN STOP confirmed`. Repeat
the three commands on the same tree to test re-IPL. To exercise the escalation path, run
`up` and `run`, and then reproduce the hang on purpose with
`echo 'script scripts/poweroff' > ~/val/tk5.fifo`. Then run `./mvsjob.sh down`. It must
send nothing, kill only after all output files are closed, and exit 2.

## 7. Related Docs

- [`docs/runbooks/mvsjob.sh`](../runbooks/mvsjob.sh): design notes 1–7
- [`docs/runbooks/tk5-hercules-setup.md`](../runbooks/tk5-hercules-setup.md) §12, Traps 6–8
- [`RUN.md`](../../RUN.md) §6.4, §7.2, §7.4, §8
- [`docs/architecture/c4/sequence-state-object-diagrams.md`](../architecture/c4/sequence-state-object-diagrams.md) §2
- [`2026-08-10-mvsjob-permission-and-shutdown-fixes.md`](2026-08-10-mvsjob-permission-and-shutdown-fixes.md):
  the change that made `poweroff` the default, which this one supersedes for shutdown
- [`docs/evidence/E0-tk5-boot-2026-07-26.md`](../evidence/E0-tk5-boot-2026-07-26.md):
  the original clean-stop rule and the clean E0 stop

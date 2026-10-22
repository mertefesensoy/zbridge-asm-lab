# Runbook · TK5 (MVS 3.8j under Hercules) on Windows — Track M setup to rung E0

**Date:** 2026-07-25, executed 2026-07-26
**Status:** **E0 PASSED 2026-07-26** — but by a different route than this document
describes. See §12, added after execution.
**Governing decision:** `docs/decisions/0001-emulation-strategy-hercules-two-track.md`
**Evidence:** `docs/evidence/E0-tk5-boot-2026-07-26.md`

> ## Read §12 first
>
> This runbook describes a **Windows, interactive, 3270-terminal** path. E0 was
> passed on 2026-07-26 by a **WSL2, headless, socket-reader** path instead, which
> eliminates three of the five traps below and needs no terminal client and no
> operator. **§12 has the procedure that actually ran**, every ⚠ VERIFY item answered
> from TK5's own configuration files, and two new traps this environment adds.
>
> §1–§11 are retained unchanged: they remain correct for a Windows install, and the
> traps they document are real. Where they disagree with §12, §12 is what was tested.

---

## How to read this runbook

Items marked **⚠ VERIFY** are things I could not confirm from a primary source and
that you should read out of TK5's own bundled documentation rather than trust here.
They are marked because a runbook that sounds equally confident about everything is
worse than useless — it hides where the risk is.

Items marked **🛑 TRAP** are failure modes specific to *this machine and this
project*. Read all of §1 before downloading anything.

The goal of this runbook is exactly **rung E0** from ADR 0001: TK5 IPLs, you can read
the operator console, and you can submit a job and get its output back. Nothing
about WTO happens here. E1 onward is a separate document written after E0 lands.

---

## §0 · What you are building

Hercules is a hardware emulator. TK5 is a pre-installed MVS 3.8j system that runs on
it — a complete 1981-era IBM mainframe operating system, legally distributable,
bundled with the emulator and ready to IPL. You are not installing an operating
system; you are unpacking a machine that already has one.

The reason this is in the project at all, per ADR 0001: MVS 3.8j is the direct
ancestor of z/OS and implements a real `SVC 35`. It is the only oracle available for
unknown **U2** (is our WTO parameter list byte-correct?) before mainframe access.

What you will have at the end of E0:

- A running mainframe on your laptop
- A 3270 terminal session into it
- The ability to submit JCL and retrieve output
- Evidence captured in `docs/evidence/` for the E0 rung

---

## §1 · Prerequisites and traps — read before downloading

### 🛑 TRAP 1 — Do NOT install TK5 anywhere under OneDrive

Your repo lives at `C:\Users\senso\OneDrive\Masaüstü\IBM Z Project\`. **TK5 must not
go there, or anywhere else inside OneDrive.**

TK5's DASD volumes are multi-hundred-megabyte files that the emulator writes to
continuously while running. OneDrive will attempt to sync every change. The
consequences range from annoying to destructive: sync storms saturating your upload,
file locks causing Hercules I/O errors mid-run, and — the serious one — OneDrive
reconciling a partially-written DASD volume and corrupting the system beyond repair.

**Install to a short, local, ASCII-only path outside OneDrive.** Recommended:

```
C:\mvs-tk5
```

### 🛑 TRAP 2 — Path characters and spaces

Hercules and its shell scripts are decades-old C and batch tooling. The `ü` in
`Masaüstü` and spaces in `IBM Z Project` are both plausible sources of obscure
failures in that stack. `C:\mvs-tk5` avoids the question entirely. This is also why
Trap 1's recommendation is a root-level directory rather than somewhere in your
profile.

### 🛑 TRAP 3 — Antivirus and Windows Defender

Real-time scanning of DASD volume files during emulator I/O causes severe slowdowns
and occasional I/O errors. Add `C:\mvs-tk5` as an exclusion **before** first boot.
Windows Security → Virus & threat protection → Manage settings → Exclusions → Add a
folder.

### 🛑 TRAP 4 — Firewall prompt on first run

Hercules listens on local TCP ports for terminals and the card reader (§4, §8).
Windows will prompt on first launch. **Allow on private networks only.** There is no
reason for this to be reachable from anything but localhost, and an MVS 3.8j system
has 1981-era security — it should never be exposed.

### What you need

| Requirement | Notes |
|---|---|
| Disk | ⚠ VERIFY against the TK5 page — budget **~2 GB** unpacked to be safe |
| RAM | Trivial by modern standards; MVS 3.8j ran in 16 MB |
| Hercules | **Bundled with TK5** (SDL Hyperion 4.9.1, 64-bit Windows) — do not install separately |
| 3270 terminal client | ⚠ VERIFY — see §4 |
| Unzip tool | Windows built-in is fine |

---

## §2 · Download and unpack

**Source (the only one to use):** <https://www.prince-webdesign.nl/tk5>

Current level at time of writing: **TK5 Update 5, released 2026-02-18**, which bundles
Hercules SDL 4.9.1 64-bit for Windows.

The page offers two downloads. You want the **complete system** (`mvs-tk5.zip`), not
the update-only package (`mvstk5-update5.zip`) — the latter is for people who already
have TK5 installed.

1. Download `mvs-tk5.zip`.
2. Unpack to `C:\mvs-tk5` (see Trap 1 and 2).
3. Add the antivirus exclusion (Trap 3).
4. **Before running anything, read the bundled documentation.** TK5 ships with its own
   manuals and a README. Where this runbook and TK5's own docs disagree, **TK5's docs
   win** — they describe the exact package you downloaded; this runbook was written
   without it in hand.

Record the SHA-256 of the downloaded zip for the evidence file:

```powershell
Get-FileHash -Algorithm SHA256 C:\mvs-tk5.zip
```

---

## §3 · First boot

From the install directory:

```powershell
C:\mvs-tk5\mvs.bat
```

This starts Hercules and IPLs MVS. A console window appears showing the Hercules
control panel and the emulated mainframe's boot messages.

**What a healthy boot looks like:** a stream of MVS initialisation messages ending
with the system quiescing into a ready state, JES2 started, and no repeating error
loop. MVS 3.8j is chatty at IPL — a lot of output is normal.

**⚠ VERIFY:** whether TK5 IPLs fully unattended or pauses for operator replies at the
console. TK4- historically required a couple of console replies during IPL (typically
answering a prompt with `R 00,...`). If the boot appears to stall, look for an
outstanding WTOR — a message with a reply ID — rather than assuming a hang. This is
your first encounter with the message-reply model that WTOR (roadmap Phase 4) is
built on, so it is worth understanding rather than working around.

**Do not close this window.** See §6 for shutdown.

---

## §4 · Connect a 3270 terminal

The Hercules console window is the *hardware* console. To actually use MVS you need a
3270 terminal session.

**⚠ VERIFY:** whether TK5 bundles a 3270 client and which one. Check the install
directory and TK5's README first. If it does not bundle one, the standard free
Windows options are:

- **wc3270** — the Windows build of the x3270 family. The usual recommendation, and
  scriptable (`ws3270`/`x3270if`), which matters for §8.
- **Vista tn3270** — free for personal use, widely used in the Hercules community.

Connect to:

```
localhost:3270
```

**⚠ VERIFY** the port against TK5's bundled `.conf` file — the terminal device
definitions state it explicitly. Port 3270 is the community convention but the conf
file is authoritative.

A successful connection shows a green-on-black VTAM/TSO logon screen. That screen is
the first thing in this project that looks like a mainframe.

---

## §5 · Log on

**⚠ VERIFY all credentials against TK5's bundled documentation.** The TK4-/TK5
lineage conventionally ships with pre-defined userids `HERC01` through `HERC04` plus
`IBMUSER`, with a shared default password. I am deliberately not stating the password
here rather than risk sending you down a wrong path — it is in TK5's docs, on the
first page or two.

`HERC01` is conventionally the general-purpose user with full access, and is the one
to use for assembler work.

Once logged on you are in TSO. This is where you will edit source, submit jobs, and
browse output from E1 onward.

---

## §6 · Clean shutdown — learn this before you need it

### 🛑 TRAP 5 — Never close the Hercules window to stop the system

DASD writes are buffered. Killing the emulator mid-write can corrupt the volumes and
leave you reinstalling from the zip. This is the single most common way people lose a
TK5 system.

**⚠ VERIFY the exact sequence in TK5's documentation.** The shape of it, common to
the TK4-/TK5 lineage:

1. From the MVS operator console, stop JES2 and quiesce the system — TK5 provides a
   documented shutdown procedure (conventionally a `SHUTDOWN` started task rather
   than issuing the individual commands by hand).
2. Wait for MVS to report the system has quiesced.
3. Only then stop Hercules from its console.

**Practise a clean shutdown and a clean restart before you do any real work.** The
cost of learning this now is five minutes; the cost of learning it later is the
system.

### Back up the moment E0 passes

Once TK5 boots and shuts down cleanly, **shut it down and copy the whole
`C:\mvs-tk5` directory** to a backup location (an external drive or a non-syncing
folder — not OneDrive, per Trap 1). A known-good snapshot means every later
experiment is cheap to recover from. You are about to start writing supervisor-call
code on this system; assume you will break it at least once.

---

## §7 · Submit your first job

The E0 gate is: a job goes in, output comes back. Interactively, from TSO:

1. Create a small JCL member (TK5 ships sample jobs — start by submitting one of
   theirs rather than writing your own, so a failure is unambiguously environmental).
2. `SUBMIT` it.
3. Retrieve the output. **⚠ VERIFY** which spool browser TK5 provides — the TK4-/TK5
   lineage typically ships one for viewing held output.

You have passed E0 when you can point at a job's output listing and say what its
return code was.

---

## §8 · The automation path (design note — build this after E0)

Clicking through 3270 screens does not scale to a documented, repeatable experiment
ladder, and the whole point of this project's workflow is that every rung produces
committed evidence. The mainframe-native answer is the **socket card reader**.

Hercules can attach a card reader device to a TCP socket. JCL pushed to that socket
is read into the job queue exactly as if it had been fed on punched cards. Printer
output is written to a file on the Windows side. Together that gives a fully
scriptable pipeline with no terminal interaction at all:

```
JCL file on laptop  →  socket reader  →  MVS runs the job  →  printer file  →  docs/evidence/
```

**⚠ VERIFY** the reader and printer device definitions and their ports in TK5's
bundled `.conf` file. The community convention maps the card reader to port **3505**
and the printer to **1403** (the historical IBM device numbers, which is a nice
touch), but the conf file is authoritative.

Once E0 passes, this becomes a small PowerShell wrapper — `Submit-MvsJob` — that
takes a `.jcl` file, submits it, waits for the printer output, and drops the listing
into `docs/evidence/` with a provenance header. That script is the Track M analogue of
CASSANDRA's run harness, and it is what makes E1→E3 reproducible rather than
anecdotal.

Do not build it before E0 passes. Automating a pipeline you have not yet run by hand
is how you end up debugging two things at once.

---

## §9 · E0 evidence checklist

Capture into `docs/evidence/E0-tk5-boot-<date>.md` with the provenance header from
ADR 0001 §7:

- [ ] TK5 version/update level and the download SHA-256
- [ ] Bundled Hercules version (from the Hercules console banner at startup)
- [ ] Host: Windows version, CPU, RAM
- [ ] Install path (confirming it is outside OneDrive)
- [ ] IPL console output — full capture, including any operator replies needed
- [ ] 3270 client used and its version; the port you connected on
- [ ] Screenshot of a successful TSO logon
- [ ] The sample job you submitted, and its output listing with return code
- [ ] Clean shutdown output, and confirmation of a successful second IPL afterwards
- [ ] Reader/printer port numbers as found in the `.conf` file (input to §8)
- [ ] **A list of every ⚠ VERIFY item in this runbook, with what you actually found**

That last item is the one that upgrades this runbook from a draft to a document. Send
me those findings and I will fold them in and remove the markers.

---

## §10 · Troubleshooting

| Symptom | Likely cause | Action |
|---|---|---|
| IPL appears to hang | Outstanding WTOR awaiting an operator reply | Look for a message with a reply ID; reply per TK5 docs. Not a hang. |
| 3270 client will not connect | Wrong port, or Hercules not fully up | Check the `.conf` device definitions; wait for IPL to finish first |
| Very slow I/O, intermittent errors | Antivirus scanning DASD volumes | Apply the Trap 3 exclusion, restart |
| Weird path/script errors | Non-ASCII or spaces in install path | Reinstall to `C:\mvs-tk5` |
| System will not IPL after a crash | DASD corruption from an unclean stop | Restore from the §6 backup |
| Sync conflicts, locked files | Installed under OneDrive | Move out of OneDrive entirely (Trap 1) |

---

## §11 · What comes next

E0 passing unblocks **E1**: assemble and run a program using the `WTO` macro, and
capture the assembler listing. That listing is Line 2 evidence for **H001**
(`docs/hypotheses/001-mvs38j-svc35-wto-oracle.md`) — it shows the exact parameter-list
bytes IBM's own macro generates, which is the ground truth this project needs and
which no manual can be fully trusted for.

E1 has its own runbook, written once E0's ⚠ VERIFY findings are in.

---

## §12 · What actually ran — the headless WSL2 procedure (added 2026-07-26, E0 PASSED)

E0 passed with **no 3270 terminal, no operator, and no Windows install.** TK5 bundles
Linux startup scripts *and* a Linux x86-64 Hercules binary, so the whole system runs
inside WSL2 on ext4.

### Why this route is better than §1–§7

| §1 trap | Under WSL2 / ext4 |
|---|---|
| Trap 1 — never install under OneDrive | **Gone.** ext4 inside the WSL VHD; OneDrive cannot see it |
| Trap 2 — path characters and spaces | **Gone.** `/root/mvs-tk5` |
| Trap 3 — antivirus scanning DASD | **Gone in practice.** Defender does not real-time scan inside the WSL2 VHD |
| Trap 4 — Windows firewall prompt | **Not encountered.** Ports bind in the WSL network namespace |
| Trap 5 — never stop the emulator to stop the system | **Fully applies, and bit us three times.** See Trap 7 |

Do **not** install under `/mnt/c`: DASD I/O across the 9p boundary is slow and it
reintroduces Traps 1–3.

### 🛑 TRAP 6 (new) — WSL2 idles its VM out, and `/tmp` is tmpfs

An idle-out while Hercules holds buffered DASD writes is Trap 5 by another route. Keep
everything under your home directory (`/root` until 2026-08-10, `~/mvs-tk5` since),
never `/tmp`, and **leave the system shut down between sessions** rather than running.
*(2026-10-21: with `systemd=true` in `/etc/wsl.conf`, the VM was once seen to stay up
instead. That is not a guarantee either, so the rule stands.)*

### 🛑 TRAP 7 (new) — a detached Hercules dies on stdin EOF

Hercules reads console commands from stdin. If stdin is a FIFO held open by a helper
process and that helper is reaped with its process group, Hercules sees EOF and exits
**without running MVS's shutdown**. This produced three unclean stops on 2026-07-26.

Rules that follow, and they are not optional:

1. **A clean stop is one complete Hercules shutdown ending in `HHC01412I Hercules
   terminated`, followed by the process exiting.** Process absence proves nothing — it
   is equally consistent with a kill mid-write. *(Corrected 2026-10-21: `HHC01412I` on
   its own is not enough either, because it can be forged. See Trap 8.)*
2. **Release the stdin holder only after the process has exited.**
3. **Assert exactly one instance** before touching the tree:
   `pgrep -cf 'hercules -f conf/tk5.cnf'` must be `1`. Never modify the tree while any
   instance is live — two emulators on one set of CCKD volumes makes the state
   untrustworthy however either stopped.

### 🛑 TRAP 8 (new, 2026-10-21) — powering off with MVS still up hangs Hercules

`scripts/poweroff` is only `stopall`, `sysclear` and `quit`, so it never shuts MVS
down. With MVS still running, Hercules' shutdown hangs partway, and it did so in **4
stops out of 4** on 2026-10-21 (one of them reproduced on purpose).

**What the log shows.** `HHC01420I Begin Hercules shutdown`, then `HHC01423I Calling
termination routines`, then a statistics block for every DASD volume (they are closed
cleanly). After that, nothing. `HHC01422I Configuration released` and `HHC01424I All
termination routines complete` never appear, and the process never exits.

**Why.** TK5's `conf/tk5_default.cnf` defines 16 dial-in lines, 0670–068F, as 2703
devices. These are four RJE lines, four TTY lines and eight 2741 lines, and zbridge uses
none of them. While MVS runs, every one of them holds an outstanding ENABLE, waiting for
a caller. Hercules' `devlist` shows `OP=ENABLE … open busy` on all 16. Hercules detaches
devices one at a time, and closing a line in that state deadlocks: line 0670 closed and
0671 never did, in every hang.

**The trap inside the trap: a second `quit` forges the "clean" message.** Hercules'
termination code runs only once (`hdl_atexit()` returns immediately on re-entry). So a
second `quit`, or a SIGTERM, re-enters the shutdown, skips the hung routine, and still
prints `HHC01424I`, `HHC01425I` and `HHC01412I Hercules terminated`. The process then
stays blocked and needs SIGKILL. On 2026-10-21 that forged `HHC01412I` was at first
taken for a clean stop.

Rules that follow:

1. **Shut MVS down before Hercules**, with TK5's own sequence:
   - `/f bsppilot,shutnow` → `$HASP099` (~72 s)
   - `/$pjes2` → `$HASP085`
   - `/z eod` → `IEE334I`
   - `/quiesce` → `HHC00809I`

   After that, all 16 lines are idle, and the power-off completed in 16 s, 7 runs out of
   7. `docs/runbooks/mvsjob.sh down` runs these steps itself, each with a timeout and a
   log line. It does not use `scripts/shutdown`, which chains the same steps through HAO
   triggers but twice stopped partway on 2026-08-10 with no log of which link broke.
2. **Never send a second `quit` or a SIGTERM.** Either one forges `HHC01412I`.
3. **If it hangs anyway, kill only after the disks are closed, then prove it.** Use
   SIGKILL, and only once the process has no file under `dasd/`, `prt/` or `pch/`, and
   not `log/hardcopy.log`, open (`/proc/<pid>/fd`). Then check that every CCKD volume has
   its OPENED bit clear (byte 515, `0x80`) and passes `cckdcdsk -ro -3`. **Record it as a
   forced stop, never as a clean one.** In the lab, every forced stop made this way left
   all 15 volumes intact.

The lab runs behind these numbers (stock, lines removed, MVS first; fresh tree each time)
are in `docs/implementations/2026-10-21-mvsjob-down-mvs-first-shutdown.md`.

### The procedure

*(Paths updated 2026-10-21. The install moved from `/root` to `~/mvs-tk5` on
2026-08-10.)*

```bash
# One-time: fetch and unpack. Record the hash — it is what makes recovery cheap.
mkdir -p ~/tk5dl && cd ~/tk5dl
curl -fL -o mvs-tk5.zip https://www.prince-webdesign.nl/images/downloads/mvs-tk5.zip
sha256sum mvs-tk5.zip     # 710d002843631322810a276dd42c793fda458548dc64d86e2914a62db7425f84
unzip -q mvs-tk5.zip -d ~/ && chmod -R +x ~/mvs-tk5
```

```bash
# IPL, headless. TK5's scripts/ipl.rc arms HAO before the IPL, so IEA101A and IEA305A
# are answered automatically and no operator is needed.
T=~/mvs-tk5; cd $T
mkfifo ~/tk5.fifo
echo CONSOLE > unattended/mode
export PATH="$T/hercules/linux/64/bin:$PATH"
export LD_LIBRARY_PATH="$T/hercules/linux/64/lib:$T/hercules/linux/64/lib/hercules"
export HERCULES_RC=scripts/ipl.rc
setsid nohup bash -c "exec sleep infinity > $HOME/tk5.fifo" &
setsid nohup hercules -f conf/tk5.cnf < ~/tk5.fifo >> log/3033.log 2>&1 &
# ready for jobs when log/3033.log contains: IEE136I LOCAL:   (~15-30s)
# TK5 has finished starting VTAM, TCAM and TSO about 80s later, at:
#   MVS038J MVS 3.8j TK5 system initialization complete
```

```bash
# Submit a job through the socket card reader. No terminal, no netcat needed.
cat job.jcl > /dev/tcp/127.0.0.1/3505
# output appears in prt/prt00e.txt ; console in log/3033.log ; hardcopy in log/hardcopy.log
```

```bash
# Clean shutdown: MVS first, then Hercules (Trap 8). docs/runbooks/mvsjob.sh down does
# exactly this, with a timeout and a report line per step - prefer it. By hand, send
# each line only after the message in its comment appears in log/3033.log:
echo '/f bsppilot,shutnow' > ~/tk5.fifo   # $HASP099  (about 72 s)
echo '/$pjes2'             > ~/tk5.fifo   # $HASP085
echo '/z eod'              > ~/tk5.fifo   # IEE334I
echo '/quiesce'            > ~/tk5.fifo   # HHC00809I (CPU in disabled wait)
echo 'script scripts/poweroff' > ~/tk5.fifo
# Clean = exactly one HHC01420I, then HHC01422I, HHC01424I and HHC01412I, AND the
# hercules process gone. Send nothing else meanwhile: a second quit forges HHC01412I.
# Only then release the stdin holder - the sleep whose stdout is ~/tk5.fifo.
```

### Every ⚠ VERIFY item, answered from `conf/tk5.cnf` and the bundled scripts

| Item | Answer |
|---|---|
| Hercules version | **4.9.1.0-SDL** (SDL Hyperion), built 2025-12-08. **A Linux x86-64 build ships in `hercules/linux/64/bin/hercules`** |
| Disk | **593 MB unpacked**, 498 MB zip — not the ~2 GB budgeted |
| §3 unattended IPL? | **Fully unattended.** `scripts/ipl.rc` uses HAO (`hao tgt IEA101A`, `hao tgt IEA305A`) to auto-reply. `IEA101A` appearing is normal, not a hang |
| §4 3270 client | **Not needed for E0–E3.** Socket reader + printer files replace it entirely |
| §4 console port | **3270 confirmed** (`CNSLPORT ${CNSLPORT:=3270}`) |
| §8 card reader | **Port 3505 confirmed** (`000C 3505 ${RDRPORT:=3505} sockdev ascii trunc eof`) |
| §8 printer | **1403 is a device *type*, not a port.** Printers write to files: `000E`→`prt/prt00e.txt`, `000F`→`prt/prt00f.txt`, `0002`→`prt/prt002.txt`, `030E`→`log/hardcopy.log` |
| §5 credentials | **Not needed for E0** — no logon was performed. Still unverified |
| §6 shutdown | **`scripts/shutdown`**, drivable from the Hercules console; also `quiesce`, `poweroff`, `z_eod`. *2026-10-21: run its MVS steps one at a time, then `poweroff` (Trap 8); `poweroff` alone hangs with MVS up* |
| CPU config | `CPUMODEL 3033`, `CPUSERIAL 000611`, `ARCHLVL S/370`, `MAINSIZE 16`, `NUMCPU 1` |
| Bonus | `HTTP PORT 8038` with `HTTP START` — Hercules serves a web console |
| §2 licence terms | **The TK5 download page states none for MVS 3.8j.** ADR 0001's evidence item 3 rests on its original sources, not on that page |

### §8's automation was not built after E0 — it *was* E0

The design note in §8 says to build the socket-reader pipeline only once E0 passes by
hand. In practice E0 was passed *by* that pipeline, so it is already proven end to end:
JCL file → port 3505 → JES2 → `prt/prt00e.txt` → `docs/evidence/`. What remains for a
`Submit-MvsJob` wrapper is packaging, not discovery.

### What E1 needs next

E1 assembles a program using the `WTO` macro and captures the listing — **the listing is
the ground-truth source for the WTO parameter-list byte layout**, which has no primary
citation on any system. Two constraints to carry in:

- **The assembler is IFOX00 (Assembler XF), not HLASM** (ADR 0001 evidence item 6): no
  dependent or named `USING`, no long displacement, no relative-immediate forms.
- **`log/hardcopy.log` (device `030E`) is where a WTO message lands as a file.** That is
  what makes E1's console evidence capturable without a screenshot.

## Links

- `docs/evidence/E0-tk5-boot-2026-07-26.md` — the E0 evidence, including the three
  unclean stops and the corrections adopted
- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — why Track M exists
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — what E1–E3 are testing
- TK5: <https://www.prince-webdesign.nl/tk5>
- Hercules documentation (mentor-supplied): <https://hercules-390.github.io/html/>
- Jay Moseley's MVS reference material: <https://www.jaymoseley.com/hercules/>

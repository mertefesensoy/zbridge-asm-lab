# Runbook · TK5 (MVS 3.8j under Hercules) on Windows — Track M setup to rung E0

**Date:** 2026-07-25
**Status:** written, not yet executed
**Governing decision:** `docs/decisions/0001-emulation-strategy-hercules-two-track.md`

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

## Links

- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — why Track M exists
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — what E1–E3 are testing
- TK5: <https://www.prince-webdesign.nl/tk5>
- Hercules documentation (mentor-supplied): <https://hercules-390.github.io/html/>
- Jay Moseley's MVS reference material: <https://www.jaymoseley.com/hercules/>

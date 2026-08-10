# RUN.md — how to build, test, and demo this repository, from a clean machine

This document assumes **nothing is installed yet** and walks through every dependency,
every external binary, and every command needed to reproduce everything this project
claims — up to and including putting a message on a real (emulated) mainframe operator
console from Go-built bytes. If you can follow this document top to bottom on a machine
that has never seen this repository, it has done its job.

**Read [`docs/architecture/README.md`](docs/architecture/README.md) first** if you want
to understand *what* you're running and *why* before you run it. This document is the
"how," not the "why."

---

## 1. The two tracks, and what each one proves

| | Track A — laptop only | Track B — full emulated demo |
|---|---|---|
| **What it needs** | Go 1.26+. Nothing else. | Everything in Track A, **plus** WSL2, QEMU, and TK5/Hercules (§3) |
| **What it proves** | All Go code compiles, all unit/oracle/differential tests pass, and every module cross-compiles for `linux/s390x` | The exact same Go-built bytes are **accepted by a real (emulated) `SVC 35`** and a message appears on an operator console |
| **Time to first result** | Under a minute | 20–40 minutes the first time (mostly a ~500 MB download); a few minutes on every run after that |
| **This is what the mentor/committee sees as "it works on paper"** | ✅ | — |
| **This is the actual thesis result — the demo** | — | ✅ |

If you only have a few minutes, run Track A. If you're presenting, you want Track B —
specifically §7, which is the exact sequence that puts a message built by Go on a
1981-era mainframe's operator console.

---

## 2. Dependency inventory — everything, with exact versions and sources

| Dependency | Version used in this repo's own evidence | Where it comes from | Needed for |
|---|---|---|---|
| **Go** | 1.26.3 (the toolchain; `go.mod` in every module requires `go 1.26`) | <https://go.dev/dl/> | Track A and B |
| **WSL2** (Windows only) | Ubuntu 26.04 LTS ("resolute"), kernel `6.6.114.1-microsoft-standard-WSL2` | Built into Windows 11 (`wsl --install`) | Track B |
| **qemu-s390x** (user-mode CPU emulator) | 10.2.1 (Debian package `1:10.2.1+ds-1ubuntu3.1`) | Ubuntu package `qemu-user` (or `qemu-user-static`) inside WSL2 | Track B §6 |
| **qemu-system-s390x** (full-system emulator; installed but not required for anything in this document) | 10.2.1 | Ubuntu package `qemu-system-misc` inside WSL2 | Not required by any command below |
| **s390x-linux-gnu-objdump** (disassembler, for verifying hand-encoded instructions) | 2.46 | Ubuntu package `binutils-s390x-linux-gnu` inside WSL2 | Optional — only needed if you want to re-verify the hand-encoded `TR`/`SVC` bytes yourself, per [`testing.md`](docs/architecture/testing.md) §3 |
| **TK5** (MVS 3.8j + bundled Hercules 4.9.1.0-SDL) | Update 5, released 2026-02-18. Zip size 498,312,872 bytes, **SHA-256 `710d002843631322810a276dd42c793fda458548dc64d86e2914a62db7425f84`** | <https://www.prince-webdesign.nl/images/downloads/mvs-tk5.zip> | Track B §7 |

**No cgo, no C compiler, no Language Environment install of any kind is needed anywhere
in this list.** That absence is not an omission — it's the entire point of the project
(see [`docs/architecture/README.md`](docs/architecture/README.md) §4).

**Verify the TK5 download before trusting it.** Anyone can put a file at a URL; the hash
is what makes this reproducible rather than an act of faith. §4.3 downloads it *inside*
WSL2, so `sha256sum` (below) is the one you'll actually use — the PowerShell form is
here too only because it's easy to reach for `sha256sum` out of habit in a Windows
terminal and get "not recognized as a cmdlet" instead:

```bash
# inside WSL2 — this is the one §4.3 actually uses
sha256sum mvs-tk5.zip
# must print: 710d002843631322810a276dd42c793fda458548dc64d86e2914a62db7425f84
```

```powershell
# if you instead downloaded it on the Windows side — sha256sum doesn't exist in PowerShell
(Get-FileHash mvs-tk5.zip -Algorithm SHA256).Hash
# must print (case-insensitive): 710D002843631322810A276DD42C793FDA458548DC64D86E2914A62DB7425F84
```

---

## 3. Track A — laptop only, no setup beyond Go

From the repository root, each module is built and tested from **its own directory**
(every module has its own `go.mod` — see [`zbridge-module.md`](docs/architecture/zbridge-module.md) §1
for why they're kept separate).

### 3.1 Vet and test every module

```powershell
foreach ($m in 'add','ebcdic','strmanip','regs','bytecmp','syscall-linux','zbridge') {
  Write-Host "===== $m =====" -ForegroundColor Cyan
  Push-Location $m
  go vet ./...
  go test ./...
  Pop-Location
}
```

```bash
# equivalent, if you're running this from Git Bash / WSL instead of PowerShell
for m in add ebcdic strmanip regs bytecmp syscall-linux zbridge; do
  echo "===== $m ====="
  (cd "$m" && go vet ./... && go test ./...)
done
```

**Expected result:** `ok` for every package in every module, zero failures. `syscall-linux`
prints `[no test files]` or skips on non-Linux hosts — that's correct, not a failure; see
[`testing.md`](docs/architecture/testing.md) §4. Exact pass counts, reconciled and
verified: **27 tests in `zbridge`, 18 across the five Windows-buildable lab modules** —
see [`testing.md`](docs/architecture/testing.md) §8.

### 3.2 Cross-compile every module for `linux/s390x`

This is a **build-only** check — no code runs, but `go vet` mechanically verifies the
Go-assembly frame contract for the target architecture (see
[`README.md`](docs/architecture/README.md), "the assembly bug class this repo exists to
prevent").

```powershell
$env:GOOS = 'linux'; $env:GOARCH = 's390x'
foreach ($m in 'ebcdic','strmanip','regs','bytecmp','syscall-linux','zbridge') {
  Write-Host "===== $m (linux/s390x) =====" -ForegroundColor Cyan
  Push-Location $m
  go vet ./...
  go build ./...
  Pop-Location
}
Remove-Item Env:\GOOS, Env:\GOARCH
```

**Expected result:** clean on all six. **`add/` is deliberately excluded and is
*expected* to fail** if you try it — it declares a bodyless function served only by an
amd64 assembly file and was never meant to be ported. A failure there is the gate
working correctly, not a regression (see [`testing.md`](docs/architecture/testing.md) §5).

### 3.3 Run the benchmarks (optional)

```powershell
foreach ($m in 'add','ebcdic','strmanip','regs','bytecmp','syscall-linux','zbridge') {
  Push-Location $m
  go test -bench=. -benchmem -run='^$' ./...
  Pop-Location
}
```

These numbers are meaningful **only when run on real hardware you're currently
standing on** (i.e., this Track A path on your actual laptop CPU). Never run this
under QEMU or Hercules and treat the result as an s390x number — see
[`testing.md`](docs/architecture/testing.md) §6 for exactly why not.

---

## 4. Track B setup (one-time)

### 4.1 Install WSL2 (Windows only; skip if you're already on Linux)

```powershell
wsl --install -d Ubuntu
```

Reboot if prompted, then open the Ubuntu shell it creates and set a username/password.
Confirm the distro version:

```bash
lsb_release -a
uname -r
```

### 4.2 Install QEMU and the s390x binutils, inside WSL2

```bash
sudo apt update
sudo apt install -y qemu-user qemu-user-binfmt binutils-s390x-linux-gnu
```

Confirm versions (should be 10.2.1 / 2.46 or close — a newer point release is fine, it
just means your evidence provenance headers should record whatever you actually get,
per [ADR 0001 §7](docs/decisions/0001-emulation-strategy-hercules-two-track.md)):

```bash
qemu-s390x --version
s390x-linux-gnu-objdump --version
```

### 4.3 Download and unpack TK5

**Do not install under `/mnt/c` or anywhere OneDrive can see.** DASD volume files are
hundreds of megabytes and mutate continuously while Hercules runs; a syncing filesystem
underneath them risks corruption. Install to a plain ext4 path inside WSL2 — **your own
home directory (`~`), not `/root`.** Earlier revisions of this document used `/root`,
which is what this project's first evidence run happened to use — but `/root` is
readable only by the `root` user. If you (like most WSL2 setups) log in as a regular
user, an install under `/root` means every command in this section needs `sudo`, and
you'll get a bare `Permission denied` with no other explanation the first time you try
to skip it. Save yourself that discovery:

```bash
mkdir -p ~/tk5dl && cd ~/tk5dl
curl -fL -o mvs-tk5.zip https://www.prince-webdesign.nl/images/downloads/mvs-tk5.zip
sha256sum mvs-tk5.zip   # must match 710d0028...425f84, see §2
unzip -q mvs-tk5.zip -d ~/
chmod -R +x ~/mvs-tk5
```

Nothing else needs installing — TK5 bundles its own Linux Hercules binary
(`hercules/linux/64/bin/hercules`), so **no separate Hercules install and no 3270
terminal client are needed** for anything in this document.

### 4.4 Copy the harness into WSL2 (one-time)

`mvsjob.sh` lives in the repository under `docs/runbooks/`, which on the Windows side
means reaching it through `/mnt/c/...` — a path that, on most machines, contains spaces
(`Program Files`-style installs elsewhere on the same PATH have caused real failures
here — see the note below) and is easy to mistype. Copy it into WSL2 once, next to the
TK5 install from §4.3, and every command from here on is short and doesn't cross the
Windows/WSL2 boundary at all:

```bash
cp "/mnt/c/path/to/zbridge-asm-lab/docs/runbooks/mvsjob.sh" ~/mvsjob.sh
chmod +x ~/mvsjob.sh
```

`mvsjob.sh` reads its TK5 location from `$HOME/mvs-tk5` by default (override with the
`TK5_HOME` environment variable if yours lives somewhere else) — so as long as §4.3 and
this step used the same user account, no further configuration is needed.

**Why `mvsjob.sh` sanitizes its own `PATH` internally, and why that matters even if you
never notice it:** WSL2 imports your entire Windows `PATH` into Linux by default —
every `Program Files` entry, spaces and all. Hercules' own startup chain breaks on that
(fails with something like `bash: line 1: C:/Program: No such file or directory`) the
moment `mvsjob.sh` prepends to an inherited `PATH` containing those entries and
re-exports it. The script now resets `PATH` to a clean, space-free base before doing
anything else, so this shouldn't surface at all — but if you ever see an error shaped
like that, it means something upstream is still handing Hercules a polluted `PATH`.

---

## 5. Track B — run the s390x unit/oracle/differential tests under QEMU

This reproduces the U1 result (see [`evidence-ladder.md`](docs/architecture/evidence-ladder.md) §6):
real, big-endian s390x machine code, executed, not just compiled.

```powershell
# 1. On Windows: cross-compile test binaries for linux/s390x
$env:GOOS = 'linux'; $env:GOARCH = 's390x'; $env:CGO_ENABLED = '0'
$out = "$PWD\s390x-tests"
New-Item -ItemType Directory -Force $out | Out-Null
foreach ($m in 'ebcdic','strmanip','regs','bytecmp','syscall-linux') {
  Push-Location $m
  go test -c -o "$out\$m.test" ./...
  Pop-Location
}
Remove-Item Env:\GOOS, Env:\GOARCH, Env:\CGO_ENABLED
```

```bash
# 2. In WSL2: run each binary under the QEMU s390x user-mode emulator
cd /mnt/c/path/to/zbridge-asm-lab/s390x-tests   # adjust to wherever step 1 wrote them
for t in ebcdic strmanip regs bytecmp syscall-linux; do
  echo "===== $t ====="
  qemu-s390x "./$t.test" -test.v
done
```

**Expected result:** every test `PASS`, including `TestAtoETRMatchesLoop` and its
siblings — the differential tests that prove the hand-encoded `TR` instruction (no
Go mnemonic exists for it) does what its header comment claims. See
[`testing.md`](docs/architecture/testing.md) §3 for what this test is actually checking.

**Optional — verify the hand-encoded instruction bytes independently of execution:**

```bash
s390x-linux-gnu-objdump -d ebcdic.test | grep -A 14 '<trbody>:'
```

You should see `dc ff 20 00 30 00` disassembled back as `tr 0(256,%r2),0(%r3)` — the
disassembler recovering the instruction from raw bytes Go's own assembler had no
mnemonic to emit directly.

---

## 6. Track B — the actual demo: Go-built bytes accepted by a real `SVC 35`

This is rung **E3**, the project's central result (see
[`evidence-ladder.md`](docs/architecture/evidence-ladder.md) §5). Everything before this
section was preparation; this section is the payoff.

**This section switches between two different shells, twice. Getting this wrong is
the single most common mistake here** — it produces `go: command not found` (if you
run §6.2 inside WSL2, where Go is deliberately not installed) or `mvsjob.sh: command
not found` (if you run §6.1/6.3/6.4 in PowerShell, where there is no bash). Check this
table before typing anything:

| Step | Shell | You'll know you're in the right one because... |
|---|---|---|
| 6.1 Bring MVS up | **WSL2** | prompt looks like `yourname@yourpc:~$` |
| 6.2 Generate the job | **PowerShell** | prompt looks like `PS C:\...>` |
| 6.3 Submit it | **WSL2** | prompt looks like `yourname@yourpc:~$` |
| 6.4 Shut down | **WSL2** | prompt looks like `yourname@yourpc:~$` |

Keep both a PowerShell window/tab and a WSL2 window/tab open side by side — don't try
to do this in one terminal by typing `wsl` and `exit` repeatedly.

### 6.1 Bring MVS up

**Run this in WSL2** (type `wsl` in PowerShell/Windows Terminal, or launch "Ubuntu"
from the Start menu). `mvsjob.sh` (copied into WSL2 once, per §4.4) reads the TK5
install location from the `TK5_HOME` environment variable, defaulting to
`$HOME/mvs-tk5` — which matches §4.3, so in the normal case you don't need to set it
at all:

```bash
~/mvsjob.sh up
```

**Expected output:** `MVS up after ~<N>s` — typically 15–60 seconds on a freshly
extracted install. **No 3270 terminal, no operator, nothing to click.** The script
arms Hercules' automatic-operator feature before IPL so the boot completes unattended;
if it hasn't reported "up" within 5 minutes, something is wrong (see §9).

### 6.2 Generate the E3 job — real bytes, built by Go, right now

```powershell
cd zbridge
go run ./cmd/gen-e3 "ZBRIDGE LIVE DEMO" > zbe3go.jcl
```

Watch the diagnostics on stderr — they print the exact parameter-list bytes
`console.EncodeWPL` just built, in real time, before anything touches a mainframe:

```
message      : "ZBRIDGE LIVE DEMO" (18 chars)
WPL bytes    : 00 16 00 00 ...
length field : 22 = len(text) 18 + 4
MCS flags    : 00 00
layout       : VERIFIED (minimal single-line form) from the IFOX00 assembler expansion ...
```

`zbe3go.jcl` now contains a complete MVS assembler program whose `DC X'...'` constants
are exactly those bytes — see [`wpl-svc35-mechanism.md`](docs/architecture/wpl-svc35-mechanism.md)
§3.6–3.7 for what those bytes mean and how they were verified against a real IBM macro.
You can pass any message up to 124 characters (uppercase, digits, and punctuation are
safest — see [`README.md`](docs/architecture/README.md) §2.5 on EBCDIC) as the argument;
omit it and the job uses a default message.

### 6.3 Submit it and watch it reach the console

`zbe3go.jcl` was written on the Windows side in §6.2, so cross it into WSL2 first —
still inside the same WSL2 terminal from §6.1:

```bash
cp "/mnt/c/path/to/zbridge-asm-lab/zbridge/zbe3go.jcl" ~/
~/mvsjob.sh run ~/zbe3go.jcl
```

**Expected output** — this is the moment the demo is about:

```
submitted ZBE3GO
ZBE3GO ended after ~5s
--------------- console messages for ZBE3GO ---------------
07:39:25 / 5.39.24 JOB    1  $HASP100 ZBE3GO   ON READER1     E3 GO BYTES SVC35
...
07:39:25 / 5.39.25 JOB    1  +ZBRIDGE LIVE DEMO
07:39:25 / 5.39.25 JOB    1  $HASP395 ZBE3GO   ENDED
...
--------------- output saved: ~/mvsjob-out/ZBE3GO.txt (NNN lines) ---------------
IFOX00 RC= 0000   (assembler)
IEWL   RC= 0000   (linker)
GO     RC= 0000   (the SVC 35 itself)
```

That `+ZBRIDGE LIVE DEMO` line — with your actual message — is the whole result. If you
don't see it in the "console messages" block, check the raw log directly before
assuming failure: `grep -a "<your message>" ~/mvs-tk5/log/hardcopy.log`.

**What just happened, in one sentence:** a message built entirely by Go code running on
this laptop was assembled into a mainframe program by a real assembler, linked by a
real linker, and handed to a real `SVC 35` instruction on an emulated 1981 IBM
mainframe — and it printed on the operator console with a return code of zero at every
step. The leading `+` is MVS marking the message as an unauthorized, problem-state
WTO — present and displayed, not blocked (see
[`evidence-ladder.md`](docs/architecture/evidence-ladder.md) §3).

### 6.4 Shut down cleanly

**Do not skip this**, and do not just close the terminal — an unclean stop can corrupt
the emulated disk (see [`emulation-harnesses.md`](docs/architecture/emulation-harnesses.md) §3.4).

```bash
~/mvsjob.sh down
```

**Expected output:** `CLEAN STOP confirmed (HHC01412I) after ~<N>s`, typically under
90 seconds. If it hasn't appeared within six minutes, see §9 — do **not** kill the
process manually. **If it does time out, don't panic and don't force-kill anything
first** — send a harmless command to prove Hercules is still alive
(`echo '/d t' > ~/tk5.fifo`, then check `tail ~/mvs-tk5/log/3033.log` for a reply a
few seconds later) and simply run `~/mvsjob.sh down` again. That sequence — a
responsive system whose shutdown request just didn't land — is a known, recoverable
state, not disk corruption.

---

## 7. The demo script — a suggested live run-through

This section is written to be read from **during** a presentation, not learned from —
keep it open on a second screen. §7.1–§7.2 are done before anyone is watching; §7.3 is
the live sequence; §7.4 is what to do if something goes sideways in front of an
audience; §7.5 is how to land it.

### 7.1 Pre-flight checklist — the night before or the morning of, never live

Do every one of these *before* anyone is watching. None of them belongs in front of an
audience.

- [ ] `go build ./...` and `go test ./...` pass clean in `zbridge/` right now, on the
      exact machine you'll present from (§3.1).
- [ ] WSL2 is up, TK5 is already unpacked at its install path, and its zip hash still
      matches §2 (you verified this once at setup time — no need to redo the download,
      just confirm the install is the one you tested against).
- [ ] MVS is currently **down** — start the day from a known-clean state (§6.1), not
      mid-session from whatever it was left in last time you tested.
- [ ] **Run the entire demo end to end at least once**, today, on this machine, with a
      message you actually intend to use. A dress rehearsal is what turns "I read the
      instructions" into "I've done this."
- [ ] Have this file and [`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`](docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md)
      open in tabs before you start — the second one is your fallback (§7.4).
- [ ] Decide your message in advance even if you plan to take a live suggestion —
      having a known-good fallback message ready costs nothing and removes one
      variable.

### 7.2 Timing budget

| Step | Time |
|---|---|
| MVS boot (§6.1) | 15–60 seconds on a fresh install |
| Generating the job (§6.2) | instant |
| Submit → console message (§6.3) | ~5–20 seconds |
| Shutdown (§6.4) | Typically under 90 seconds; budget up to a few minutes |

If you're closing the demo and moving to Q&A, kick off `~/mvsjob.sh down` first and let
it finish in the background of your closing remarks (§7.5) rather than standing at the
terminal waiting for it.

### 7.3 The script itself

1. **"Here's the library."** Open [`zbridge/console/wpl.go`](zbridge/console/wpl.go) —
   `EncodeWPL` is pure Go, no assembly, runs anywhere.
2. **Run `go test ./...` in `zbridge/`** (§3.1) — point at
   `TestEncodeWPLMatchesIBMMacro` passing: *"this asserts our output matches bytes a
   real IBM macro produced, byte for byte."*
3. **`mvsjob.sh up`** (§6.1) — while it boots, this is a good moment to show
   [`docs/architecture/c4/level1-context.svg`](docs/architecture/c4/README.md) and
   explain the shape of the project.
4. **`go run ./cmd/gen-e3 "<a message the audience suggests>"`** (§6.2) — take a live
   suggestion for the message text, or use your pre-decided fallback message from
   §7.1; show the bytes printed to stderr.
5. **`mvsjob.sh run zbe3go.jcl`** (§6.3) — the payoff. Point at the `+<message>` line
   and the three `RC= 0000` lines.
6. **`mvsjob.sh down`** (§6.4) — always end clean.

Total live time once MVS is up: under a minute per message.

### 7.4 If something breaks live

An emulator, a network socket, and a live audience is a lot of moving parts. If a step
stalls or errors, **don't debug live** — narrate and fall back.

1. **Say it plainly, once:** *"This is an emulator quirk — it's documented in our own
   runbook, not a flaw in the result I'm about to show you."* This is true (§8 lists
   the known failure modes) and said once, confidently, it reads as competence, not
   panic.
2. **Show the captured proof instead.**
   [`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`](docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md)
   §3 has the exact console line, the exact bytes, and all three return codes from a
   real run — walk through that file instead of the live terminal. The result is just
   as real; only the "live" part is missing.
3. **Do not restart the emulator mid-demo.** A second boot attempt live, in front of
   people, burns another 30–60 seconds you don't have and risks a second failure.
   Finish the point from the evidence file and move on.
4. **If it's the shutdown step that hangs** (§8), it's safe to simply leave it running
   and deal with it after the audience has left — an unclean stop is a laptop
   housekeeping problem, not something that affects what you just showed them.

### 7.5 Landing it

Close with the one-sentence version, not a recap of every step:

> *"That's the whole result: bytes built entirely by Go, accepted by a real `SVC 35`,
> with no C anywhere in the path — the first published proof that this is possible."*

Then bridge to questions with: *"I've documented the byte-level format and the one
open question — what happens on the real system's return code — in detail if you want
to go deeper,"* pointing at [`START_HERE.md`](START_HERE.md) §4. That section is
written exactly for this moment.

---

## 8. Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `mvsjob.sh up` never reports "MVS up" within 5 minutes | `TK5_HOME` wrong, or a stale Hercules instance already running | `pgrep -f 'hercules -f conf/tk5.cnf'` — if it prints more than one PID, do **not** proceed; see §9's note on multiple instances |
| `submit failed` in `mvsjob.sh run` | MVS isn't actually up yet, or wrong `TK5_HOME` | Confirm `up` reported success first; check `$T/log/3033.log` for `IEE136I LOCAL:` |
| No `+<message>` line appears after submit | JCL error — check the tail of `$T/log/3033.log` for `IEF452I` or `IEF642I` | The programmer-name field on the JOB card is capped at 20 characters on MVS 3.8j — this is why `gen-e3`'s JOB card is short and fixed |
| `mvsjob.sh down` never sees `HHC01412I` | Something is blocking shutdown | **Do not kill the process.** Wait longer, or consult `docs/runbooks/tk5-hercules-setup.md` §12 Trap 7 |
| Cross-compile (§3.2) fails on `add/` | Expected — see §3.2 | Not a bug; `add/` is excluded by design |
| `qemu-s390x` not found | Package not installed, or wrong distro | Re-run §4.2; confirm you're inside the WSL2 shell, not PowerShell |
| DASD/disk corruption, MVS won't IPL after a crash | An unclean stop happened at some point | Delete `~/mvs-tk5`, re-unpack from the zip (§4.3) — the SHA-256 in §2 means this is always a safe, cheap recovery |
| `Permission denied` on any `mvsjob.sh`/TK5 command | Install lives under `/root` but you're not running as `root` | Move it to `~/mvs-tk5` (§4.3) instead — this is the fix, not `sudo` on every command going forward |
| Shutdown times out without `HHC01412I` | The shutdown request may not have been processed yet | Confirm Hercules is still responsive (`echo '/d t' > ~/tk5.fifo`, check the log a few seconds later), then simply run `~/mvsjob.sh down` again — see §6.4 |

For anything not covered here, `docs/runbooks/tk5-hercules-setup.md` §10 and §12 have
the full, narrated troubleshooting history from when this was first set up, including
three unclean-shutdown incidents and exactly what fixed each one.

---

## 9. Where to go for more detail

| Question | Read |
|---|---|
| What is this project, and why does it exist? | [`docs/architecture/README.md`](docs/architecture/README.md) |
| What does each Go package actually do? | [`docs/architecture/zbridge-module.md`](docs/architecture/zbridge-module.md) |
| How was each of these results actually proven? | [`docs/architecture/evidence-ladder.md`](docs/architecture/evidence-ladder.md) |
| What do the WPL bytes mean, byte by byte? | [`docs/architecture/wpl-svc35-mechanism.md`](docs/architecture/wpl-svc35-mechanism.md) |
| What are QEMU and Hercules actually emulating? | [`docs/architecture/emulation-harnesses.md`](docs/architecture/emulation-harnesses.md) |
| The same system, as diagrams | [`docs/architecture/c4/README.md`](docs/architecture/c4/README.md) |
| How do the tests work, and which kind is which? | [`docs/architecture/testing.md`](docs/architecture/testing.md) |
| The original, narrated TK5/Hercules setup (more detail than this document, less demo-focused) | [`docs/runbooks/tk5-hercules-setup.md`](docs/runbooks/tk5-hercules-setup.md) |
| Raw evidence from the runs this document's expected outputs are drawn from | [`docs/evidence/E0-tk5-boot-2026-07-26.md`](docs/evidence/E0-tk5-boot-2026-07-26.md), [`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`](docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md), [`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`](docs/evidence/E-L-s390x-port-qemu-2026-07-25.md) |

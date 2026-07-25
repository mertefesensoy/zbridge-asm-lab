---
rung:          E-L (Track L / QEMU inner loop, ADR 0001 §5)
date:          2026-07-25
machine:       QEMU
guest_os:      n/a — user-mode emulation, no guest OS (host kernel is Linux 6.6.114.1-microsoft-standard-WSL2, Ubuntu 26.04 LTS under WSL2)
architecture:  z/Architecture (64-bit), big-endian
emulator:      qemu-s390x 10.2.1 (Debian 1:10.2.1+ds-1ubuntu3.1), user-mode
host:          Windows 11 Home 10.0.26200, Intel Core i7-13650HX (14C/20T), 15.6 GB RAM
speaks_to:     U1
hypothesis:    H002
verdict:       PASS
---

# E-L · The five `_s390x.s` bodies pass their full test suites under QEMU

**What this is not.** This is not a z/OS result, not a hardware result, and contains
no performance measurement. It speaks to **U1 only** — does our Go assembly emit
correct s390x, and does the Go ABI hold on a big-endian 64-bit target. It says
nothing about the WTO parameter list (U2) or about `GOOS=zos`, `Malloc31`, or AMODE
switching (U3).

## Toolchain, pinned

| Component | Version |
|---|---|
| Go (cross-compiler, host windows/amd64) | go1.26.3 |
| Target | `GOOS=linux GOARCH=s390x CGO_ENABLED=0` |
| QEMU user-mode | qemu-s390x 10.2.1 (Debian 1:10.2.1+ds-1ubuntu3.1) |
| GNU binutils (disassembly check) | s390x-linux-gnu-objdump 2.46 |
| WSL2 distro | Ubuntu 26.04 LTS ("resolute") |
| Hercules (installed, not used for this rung) | 3.13 |
| qemu-system-s390x (installed, not used for this rung) | 10.2.1 |

Test binaries were cross-compiled on the Windows host with `go test -c`, producing
statically linked big-endian ELF objects, then executed under `qemu-s390x`:

```
ELF 64-bit MSB executable, IBM S/390, version 1 (SYSV), statically linked
```

`MSB` is the field that matters: the binaries are big-endian, which is the property
amd64 testing could never have exercised.

## Result: 5 of 5 modules PASS

```
### qemu: qemu-s390x version 10.2.1 (Debian 1:10.2.1+ds-1ubuntu3.1)
### host: Linux 6.6.114.1-microsoft-standard-WSL2 x86_64

===================== ebcdic =====================
=== RUN   TestAtoETRMatchesLoop
--- PASS: TestAtoETRMatchesLoop (0.00s)
=== RUN   TestEtoATRMatchesLoop
--- PASS: TestEtoATRMatchesLoop (0.00s)
=== RUN   TestTRCoversEveryTableEntry
--- PASS: TestTRCoversEveryTableEntry (0.00s)
=== RUN   TestTRDoesNotOverrun
--- PASS: TestTRDoesNotOverrun (0.00s)
=== RUN   TestAtoEKnownFixture
--- PASS: TestAtoEKnownFixture (0.00s)
=== RUN   TestEtoAKnownFixture
--- PASS: TestEtoAKnownFixture (0.00s)
=== RUN   TestRoundTripAll256
--- PASS: TestRoundTripAll256 (0.00s)
=== RUN   TestEmpty
--- PASS: TestEmpty (0.00s)
=== RUN   TestSingleByte
--- PASS: TestSingleByte (0.00s)
PASS
---- ebcdic exit=0 ----
===================== strmanip =====================
=== RUN   TestStrLen
--- PASS: TestStrLen (0.00s)
=== RUN   TestWrapKnownFixture
--- PASS: TestWrapKnownFixture (0.00s)
=== RUN   TestWrapEmpty
--- PASS: TestWrapEmpty (0.00s)
=== RUN   TestWrapHighByte
--- PASS: TestWrapHighByte (0.00s)
=== RUN   TestWrapMatchesReference
--- PASS: TestWrapMatchesReference (0.00s)
PASS
---- strmanip exit=0 ----
===================== regs =====================
=== RUN   TestGetLinkRegisterNonZero
--- PASS: TestGetLinkRegisterNonZero (0.00s)
=== RUN   TestLinkRegisterIsARealReturnAddress
--- PASS: TestLinkRegisterIsARealReturnAddress (0.00s)
=== RUN   TestGRegStableWithinGoroutine
--- PASS: TestGRegStableWithinGoroutine (0.00s)
=== RUN   TestGRegDiffersBetweenGoroutines
--- PASS: TestGRegDiffersBetweenGoroutines (0.00s)
=== RUN   TestGetSPIsNotPseudoSP
--- PASS: TestGetSPIsNotPseudoSP (0.00s)
=== RUN   TestGetSPNonZero
--- PASS: TestGetSPNonZero (0.00s)
=== RUN   TestFrameSizeShiftsSP
--- PASS: TestFrameSizeShiftsSP (0.00s)
=== RUN   TestStackGrowsDownward
--- PASS: TestStackGrowsDownward (0.00s)
PASS
---- regs exit=0 ----
===================== bytecmp =====================
=== RUN   TestEqualFixtures
--- PASS: TestEqualFixtures (0.00s)
=== RUN   TestCompareFixtures
--- PASS: TestCompareFixtures (0.00s)
=== RUN   TestEqualAgainstStdlib
--- PASS: TestEqualAgainstStdlib (0.02s)
=== RUN   TestCompareAgainstStdlib
--- PASS: TestCompareAgainstStdlib (0.02s)
PASS
---- bytecmp exit=0 ----
===================== syscall-linux =====================
=== RUN   TestGetpidMatchesOS
--- PASS: TestGetpidMatchesOS (0.00s)
=== RUN   TestWriteToPipe
--- PASS: TestWriteToPipe (0.00s)
=== RUN   TestWriteBadFDReturnsErrno
--- PASS: TestWriteBadFDReturnsErrno (0.00s)
PASS
---- syscall-linux exit=0 ----
OVERALL=0
```

## Cross-compile gate (H002 claim C4), run before execution

`GOOS=linux GOARCH=s390x` on the Windows host, per module:

```
===== ebcdic =====        build OK / vet OK
===== strmanip =====      build OK / vet OK
===== regs =====          build OK / vet OK
===== bytecmp =====       build OK / vet OK
===== syscall-linux ===== build OK / vet OK
```

`go vet` verifies the `$frame-args` contract per architecture, so this is the
mechanical check that argument offsets were re-derived correctly for the s390x
declarations — including `regs`, whose Go API is no longer identical across
architectures.

Regression check on the development host (windows/amd64), all six modules including
`add`: `go vet` clean and `go test` PASS. The port changed no amd64 behaviour.

## The independent check on the hand-encoded TR

`ebcdic_s390x.s` contains six literal bytes, because Go's s390x assembler has no `TR`
mnemonic (see the finding section below). Go's assembler emitted those bytes without
understanding them, so they were handed to GNU binutils — a disassembler that does
know `TR` — and asked what they mean.

```
000000000016ae80 <trbody>:
  16ae80:	ec 58 00 20 00 7c 	cgije	%r5,0,16aec0 <trbody+0x40>
  16ae86:	a7 5f 01 00       	cghi	%r5,256
  16ae8a:	a7 44 00 12       	jl	16aeae <trbody+0x2e>
  16ae8e:	d2 ff 20 00 40 00 	mvc	0(256,%r2),0(%r4)
  16ae94:	dc ff 20 00 30 00 	tr	0(256,%r2),0(%r3)
  16ae9a:	41 20 21 00       	la	%r2,256(%r2)
  16ae9e:	41 40 41 00       	la	%r4,256(%r4)
  16aea2:	a7 5b ff 00       	aghi	%r5,-256
  16aea6:	ec 57 ff f0 00 7c 	cgij	%r5,0,7,16ae86 <trbody+0x6>
  16aeac:	07 fe             	br	%r14
  16aeae:	ec 65 ff ff 00 d9 	aghik	%r6,%r5,-1
  16aeb4:	c6 60 00 00 00 0e 	exrl	%r6,16aed0 <ebcdic_exrl_mvc>
  16aeba:	c6 60 00 00 00 13 	exrl	%r6,16aee0 <ebcdic_exrl_tr>
  16aec0:	07 fe             	br	%r14

000000000016aed0 <ebcdic_exrl_mvc>:
  16aed0:	d2 00 20 00 40 00 	mvc	0(1,%r2),0(%r4)
  16aed6:	e3 00 00 00 00 24 	stg	%r0,0
  16aedc:	07 fe             	br	%r14

000000000016aee0 <ebcdic_exrl_tr>:
  16aee0:	dc 00 20 00 30 00 	tr	0(1,%r2),0(%r3)
  16aee6:	e3 00 00 00 00 24 	stg	%r0,0
  16aeec:	07 fe             	br	%r14
```

Three independent things agree:

1. **objdump decodes `dc ff 20 00 30 00` as `tr 0(256,%r2),0(%r3)`** — exactly the
   instruction `ebcdic_s390x.s` documents in its header, with the length field
   `0xFF` meaning 256 and the base registers R2 (buffer) and R3 (table).
2. **The EXRL target decodes as `tr 0(1,%r2),0(%r3)`**, length field `0x00`, ready
   for `EXRL` to OR `len-1` into it — and `exrl %r6,...<ebcdic_exrl_tr>` is emitted
   pointing at it.
3. **QEMU executed it and produced the right bytes**, which the differential test
   (`TestAtoETRMatchesLoop`) checked against a byte-at-a-time loop implementation at
   lengths 0, 1, 2, 15, 16, 255, 256, 257, 511, 512, 513 and 4096, and which
   `TestTRCoversEveryTableEntry` checked against the Go table for all 256 input
   values.

Static encoding, disassembly, and execution therefore corroborate each other. That is
the strongest statement available without hardware.

## Findings recorded by this rung

### F1 · Go's s390x assembler cannot emit `TR`, `TRT`, `EX`, or `SVC`

Verified against the local toolchain source, `cmd/internal/obj/s390x/anames.go`
(go1.26.3): 729 mnemonics, and the translate family is entirely absent, as are `EX`
and `SVC`. Present and used by this port: `MVC`, `CLC`, `XC`, `MVCIN`, `EXRL`,
`SYSCALL`, `MOVH`, `LA`, `CMPBxx`, `CMPUBLT`, `NEG`.

**Consequence for the repo:** three `_s390x.s` stub headers described a port path
through instructions the assembler does not know. `TR` is now hand-encoded and
disassembly-verified. `SVC` is not needed on Linux — Go's `SYSCALL` assembles to
`SVC 0` — but **it will be needed for the z/OS endgame, where WTO is `SVC 35`**, and
that instruction will have to be hand-encoded exactly as `TR` was here.

The precedent is already in the Go distribution:
`golang.org/x/sys/unix/asm_zos_s390x.s` encodes `SVC 08` as `BYTE $0x0A; BYTE $0x08`.
`SVC 35` is therefore `BYTE $0x0A; BYTE $0x23`.

### F2 · `TR` translates in place, so `AtoE(dst, src)` costs MVC + TR, not TR alone

The roadmap's framing — "the lookup loop collapses into a single hardware
instruction" — is very nearly right but not exactly right for a two-buffer API.
`TR`'s first operand is both source and destination. With distinct `dst` and `src`
the sequence is `MVC` then `TR`: two SS-format instructions per 256-byte block. The
per-byte loop is genuinely gone; the claim "one instruction" is not literally true
and should be stated as "two block instructions" in the thesis.

### F3 · Upstream Go cannot target z/OS

`go tool dist list` (go1.26.3) offers `linux/s390x` and no `zos/s390x`. The string
`"zos"` appears in `internal/syslist` — so `_zos.go` filename constraints parse — but
`zos/s390x` is absent from `internal/platform`'s supported list. The z/OS files under
`GOROOT/src/cmd/vendor/golang.org/x/sys/unix/` are vendored library sources, not
toolchain support.

**Consequence:** IBM's Go fork is a hard dependency of the endgame, and that belongs
in the risk register. This is a U3 fact and no emulator changes it.

### F4 · `regs` cannot have an architecture-identical API

s390x has no frame-pointer register. Go assigns R13 to `g`, R14 to the link register,
R15 to the stack pointer, and nothing to a frame pointer
(`cmd/internal/obj/s390x/a.out.go`: `REGG = REG_R13`, `REGSP = REG_R15`). `GetBP` is
therefore amd64-only, and s390x declares `GetLinkRegister` (R14) and `GetGReg` (R13)
instead. `TestGRegDiffersBetweenGoroutines` — distinct goroutines must read distinct
values from R13 — is the test that proves the register really is `g` rather than
merely a stable number.

This is a deliberate, owner-approved deviation from H002's phrasing that each s390x
implementation "passes the identical test suite its `_amd64.s` counterpart passes."
See the H002 status note.

## Reproduction

```powershell
# 1. Cross-compile the test binaries (Windows host)
$env:GOOS='linux'; $env:GOARCH='s390x'; $env:CGO_ENABLED='0'
foreach ($m in 'ebcdic','strmanip','regs','bytecmp','syscall-linux') {
  Push-Location $m; go test -c -o "$out\$m.test" ./...; Pop-Location
}
```

```bash
# 2. Execute under QEMU (WSL2 Ubuntu)
for t in ebcdic strmanip regs bytecmp syscall-linux; do
  qemu-s390x "./$t.test" -test.v
done

# 3. Disassembly check on the hand-encoded TR
s390x-linux-gnu-objdump -d ebcdic.test | grep -A 14 '<trbody>:'
```

## What is still open after this rung

- **H002 claim C3 (emulator independence) is untested.** Only QEMU has run. Hercules
  3.13 is installed but no Hercules-hosted Linux s390x guest exists yet, so QEMU and
  Hercules have not been compared. C3 stays open.
- **No performance number was taken, deliberately.** QEMU implements `TR` as a
  software loop; timing it would measure QEMU. The roadmap's amd64-vs-s390x `ebcdic`
  benchmark table remains deferred to real hardware (ADR 0001, H002).
- **U2 and U3 are untouched by this rung.**

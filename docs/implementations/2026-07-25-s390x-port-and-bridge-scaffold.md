# 2026-07-25 · Phase 1b s390x port, emulation environment, and the production bridge module

**Author:** Claude Code session
**Scope:** three deliverables in one session — an executable s390x environment, the five
`_s390x.s` bodies with evidence, and the production `zbridge` module scaffold.

---

## 1. Problem / motivation

Three things were blocked, for three different reasons.

1. **Phase 1b (the s390x port) had never been executed.** All five `_s390x.s` files were
   `UNDEF` stubs. The stubs' headers described the intended implementation, and nobody had
   ever tried to assemble what they described.
2. **No s390x code had ever run.** ADR 0001 §5 named QEMU as the Phase 1b inner loop, but
   no emulator was installed on the machine. H002 was pre-registered with zero evidence.
3. **The production bridge module was blocked on owner scope** (goal-prompt §5 boundary 1)
   and had been since 2026-07-05.

The owner lifted the third block in this session and chose the widest scope offered — a
multi-service enterprise surface rather than a WTO-only module.

## 2. What changed

### New files

| File | Purpose |
|---|---|
| `docs/decisions/0002-zos-under-hercules-permitted-by-ibm-backing.md` | Supersedes ADR 0001 §1 on the owner's determination that IBM backing permits z/OS under Hercules. Records the premise as OWNER-ASSERTED with an unfilled citation slot |
| `docs/decisions/0003-production-bridge-module-architecture.md` | Module layout, API surface, error model, build-tag strategy, and the E3 seam |
| `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` | Full QEMU run output, cross-build gate, TR disassembly, four findings |
| `ebcdic/ebcdic_s390x.go`, `ebcdic/ebcdic_diff_s390x_test.go` | The byte-loop reference implementation and the differential test that validates hand-encoded TR |
| `regs/regs_amd64.go`, `regs/regs_s390x.go`, `regs/regs_amd64_test.go`, `regs/regs_s390x_test.go` | The per-architecture API split |
| `syscall-linux/syscall_linux_s390x.go` | s390x syscall numbers and declarations |
| `zbridge/**` | The production module: 15 files across 7 packages |

### Modified files

| File | Change |
|---|---|
| `ebcdic/ebcdic_s390x.s` | `UNDEF` → MVC + hand-encoded TR + EXRL, plus a byte-loop reference path |
| `strmanip/strmanip_s390x.s` | `UNDEF` → `MOVH` for the big-endian header, MVC via EXRL for the copy |
| `regs/regs_s390x.s` | `UNDEF` → R15/R14/R13 reads; no `GetBP` |
| `bytecmp/bytecmp_s390x.s` | `UNDEF` → CLC in 256-byte blocks with an EXRL tail, modelled on `internal/bytealg` |
| `syscall-linux/syscall_linux_s390x.s` | `UNDEF` → the s390x Linux syscall ABI |
| `regs/regs.go`, `regs/regs_test.go` | `GetBP` moved out; package doc explains why the API is not architecture-identical |
| `syscall-linux/syscall_linux.go` → `syscall_linux_amd64.go` | Renamed for symmetry with the new s390x file |
| `syscall-linux/*_test.go`, `doc.go` | Build tags widened to `linux && (amd64 \|\| s390x)` |
| `docs/hypotheses/002-s390x-port-equivalence.md` | Resolution section appended; pre-registration text preserved unchanged |

## 3. Implementation approach

### 3.1 The finding that reshaped the port

Before writing any assembly, the plan was checked against the local toolchain source
rather than against the stub headers. `cmd/internal/obj/s390x/anames.go` (go1.26.3) lists
729 mnemonics, and **`TR`, `TRT`, `TRE`, `EX` and `SVC` are all absent.**

Three of the five stubs described a port path through instructions Go's assembler cannot
emit. That is a documentation defect that had survived since Phase 1, and it would have
been discovered the hard way on borrowed mainframe time.

`MVC`, `CLC`, `XC`, `MVCIN`, `EXRL`, `SYSCALL`, `MOVH` are present, so four of the five
modules port directly. Only `ebcdic` needed a new technique.

### 3.2 Hand-encoding TR

The owner chose "TR primary + a Go-loop oracle". `TR` is emitted as literal bytes, which
is attested practice inside the Go distribution: `x/sys/unix/asm_zos_s390x.s` encodes
`SVC 08` as `BYTE $0x0A; BYTE $0x08`, and `internal/bytealg/indexbyte_s390x.s` encodes
`SRST` as `WORD $0xB25E0082`.

TR is SS-a format, six bytes: opcode `0xDC`, a length byte holding length-1, then two
base/displacement halfwords. With the buffer in R2 and the table in R3 at displacement 0:

```
TR 0(256,R2),0(R3)   ->   DC FF 20 00 30 00
TR 0(1,R2),0(R3)     ->   DC 00 20 00 30 00
```

Two properties of TR shaped the code:

- **TR translates in place.** `AtoE(dst, src)` has two buffers, so each block is `MVC`
  (copy) then `TR` (translate). The roadmap's "one instruction replaces the loop" is
  therefore two block instructions, not one. The per-byte loop is genuinely gone; the
  headline sentence needs the correction.
- **TR encodes its length in the instruction text.** Runtime lengths go through `EXRL`,
  which ORs the low 8 bits of a register into byte 1 of the target instruction — which for
  SS-a format is exactly the length field. The register holds len-1.

The `EXRL` targets are `NOSPLIT|NOFRAME` deliberately: without `NOFRAME` the assembler
emits a prologue and `EXRL` would execute the prologue instead of the intended instruction.

### 3.3 Verifying bytes nobody assembled

Hand-encoded machine code gets no assembler check, so three independent checks replace it:

1. **Disassembly.** GNU `s390x-linux-gnu-objdump` 2.46 — which does know `TR` — decodes
   the emitted bytes as `tr 0(256,%r2),0(%r3)` and `tr 0(1,%r2),0(%r3)`, and shows
   `exrl %r6,<ebcdic_exrl_tr>` pointing at the second.
2. **A differential test.** `AtoELoop`/`EtoALoop` are byte-at-a-time s390x implementations
   of the same translation. `TestAtoETRMatchesLoop` asserts byte-identical output at
   lengths 0, 1, 2, 15, 16, 255, 256, 257, 511, 512, 513, 4096 — chosen to straddle the
   256-byte block boundary in both directions.
3. **A table test and an overrun test.** `TestTRCoversEveryTableEntry` checks all 256
   inputs against the Go table directly; `TestTRDoesNotOverrun` pads the destination on
   both sides and asserts the padding is untouched.

### 3.4 The `regs` API split

s390x has no frame-pointer register. Go assigns R13 to `g`, R14 to the link register, R15
to the stack pointer, and nothing to a frame pointer (`a.out.go`: `REGG = REG_R13`,
`REGSP = REG_R15`).

The owner chose to split the API rather than return R13 from a function named `GetBP`.
`GetSP` and `GetFramedSP` stay shared; `GetBP` becomes amd64-only; s390x gains
`GetLinkRegister` (R14) and `GetGReg` (R13).

The s390x tests are stronger than the amd64 test they replace, and deliberately so:

- `TestLinkRegisterIsARealReturnAddress` calls from two `//go:noinline` call sites and
  requires the two values to differ. A constant would pass a non-zero check; only a real
  return address passes this one.
- `TestGRegDiffersBetweenGoroutines` runs eight goroutines and requires eight distinct
  values. That is what proves R13 holds `g` rather than merely a stable number.

This makes the exercise more relevant to the endgame, not less: R14 is the register MVS
linkage returns through, and R13 is the register a supervisor-call routine must not
clobber.

### 3.5 The bridge module's governing rule

ADR 0003 §4: **every exported function whose behaviour depends on an unretired unknown
returns a typed error naming that unknown — on every platform, including z/OS — until the
corresponding rung passes.**

That is what makes a wide API honest. `console.EncodeWPL` has a complete signature,
performs its validation, performs the EBCDIC conversion, computes the route and descriptor
masks — and then returns `ErrLayoutUnverified` rather than assembling bytes, because the
WPL byte layout has no primary citation on any system (research brief 003 outstanding).

`TestValidationHappensBeforeTheLayoutError` is what keeps that from rotting: bad input must
produce a validation error, never the layout error. Without it the validation code would be
unreachable in practice and would decay unnoticed until the layout landed.

### 3.6 The E3 seam

`EncodeWPL` is pure Go and runs on any host. That is not incidental factoring — it is the
shape rung E3 requires. E3 verifies the Go-side parameter-list construction without Go ever
running on a mainframe: a laptop program emits the bytes, they are embedded in an MVS
assembler program as `DC X'...'` constants, and a real `SVC 35` is handed them.

`EncodeWPL` + `FormatDC` is the entire Go side of that rung. `FormatDC` is implemented and
tested today, because it needs no verified layout.

## 4. Mathematical / bit-level details

**Route and descriptor masks.** Route codes are 1-based and map to a 16-bit mask with code
1 as the most significant bit: code *n* sets bit 2^(16-n). So route 1 is `0x8000`, route 2
is `0x4000`, route 16 is `0x0001`, and routes {1, 11} give `0x8020`. Codes outside 1..16
are ignored rather than wrapping. Descriptor codes use the identical convention.

**The EXRL length patch.** `EXRL R,target` executes the instruction at `target` after ORing
the low 8 bits of general register R into bits 8–15 of that instruction. For SS-a format
(`MVC`, `CLC`, `TR`) bits 8–15 are the length field, which encodes *length − 1*. The
assembled target therefore carries a length field of `0x00` (meaning one byte) and the
driving register holds `len − 1`. Because the operation is OR rather than replace, the
target's length field must be zero or the result is a bitwise mixture rather than the
intended length. The special case "R = R0 means no modification" refers to the register
*number*, not its value, so R6 carrying zero correctly yields a one-byte operation.

**Big-endian halfword store.** amd64 builds the 2-byte length header by hand:
`SHRQ $8` then two `MOVB` stores, high byte first. s390x stores the low 16 bits of a
register with `MOVH` (STH), and because the architecture is big-endian the high byte lands
at the lower address by construction. Four instructions become one, and lengths above
65535 truncate identically on both, preserving the documented caller contract.

## 5. Design decisions

| Decision | Alternatives considered | Why this one |
|---|---|---|
| Hand-encode TR **and** keep a byte-loop reference | TR only; byte loop only | Owner's choice. TR alone gives no check on the six bytes; the loop alone abandons the roadmap's headline result. Together they cross-validate |
| Split the `regs` API | Map `GetBP` → R13; leave `GetBP` as `UNDEF` on s390x | Owner's choice. Mapping would have preserved H002's wording by mislabelling a register in a teaching repo. The split was decided *before* any code was written, which is what keeps it a scope decision rather than a rule loosened to reach green |
| Typed errors in the bridge, not `UNDEF` stubs | Mirror the lab's `UNDEF` convention | A library must compile and be importable on a laptop so its pure-Go parts can be tested. Same principle (fail loudly), different mechanism, because the consumer differs. ADR 0003 §4 |
| `HasCode bool` in the error type | A bare `Code int32` | GC28-0683-2 p.210 documents that a single-line WTO issues no return code at all. A bare int32 makes "nothing returned" indistinguishable from "returned 0" |
| WSL2 + qemu-user + Hercules | Docker Desktop; cross-build gate only | Owner's choice; also the cheapest inner loop. Cross-compile on Windows, execute under `qemu-s390x` — no Go install needed inside WSL |

## 6. Verification

All commands were **run**, not reasoned about. Raw output is in
`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`.

```powershell
# amd64 regression, all six lab modules
foreach ($m in 'add','ebcdic','strmanip','regs','bytecmp','syscall-linux') {
  Push-Location $m; go vet ./...; go test ./...; Pop-Location
}
# -> vet clean, all PASS
```

```powershell
# s390x cross-build gate (H002 C4), five lab modules + the bridge module
$env:GOOS='linux'; $env:GOARCH='s390x'
go build ./... ; go vet ./...
# -> build OK / vet OK for all six
```

```bash
# s390x execution under QEMU 10.2.1 (WSL2 Ubuntu 26.04)
qemu-s390x ./ebcdic.test -test.v          # 9/9 PASS
qemu-s390x ./strmanip.test -test.v        # 5/5 PASS
qemu-s390x ./regs.test -test.v            # 8/8 PASS
qemu-s390x ./bytecmp.test -test.v         # 4/4 PASS
qemu-s390x ./syscall-linux.test -test.v   # 3/3 PASS
qemu-s390x ./zbridge-root.test -test.v    # 5/5 PASS
qemu-s390x ./zbridge-codepage.test        # 9/9 PASS
qemu-s390x ./zbridge-console.test         # 8/8 PASS

# the hand-encoded TR, disassembled by a tool that knows the instruction
s390x-linux-gnu-objdump -d ebcdic.test | grep -A 14 '<trbody>:'
# -> dc ff 20 00 30 00    tr    0(256,%r2),0(%r3)
```

**No performance number was taken, deliberately.** QEMU implements `TR` as a software loop.

## 7. What was NOT done, and why

- **H002 claim C3 (emulator independence) is untested.** Hercules 3.13 is installed but no
  Hercules-hosted Linux s390x guest exists, so QEMU has no counterparty. C3 stays open.
- **Rung E0 (TK5) was not run.** It needs the owner at the console. It is the owner's
  chosen next Hercules milestone.
- **No WPL layout was written.** Brief 003 is outstanding and the boundary in
  `memory/MEMORY.md` ("do not draft the parameter list against the uncited `n + 4`
  reading") was respected. This is the single largest gap in the bridge module and it is
  gated, typed, and tested rather than filled in.
- **`dataset` and `subsys` have interfaces only.** ADR 0003 declares the families and
  explicitly declines to design their operations before Phase 2's go-recordio annotation.
- **The `zos && s390x` files are compiled by nothing.** Upstream Go cannot target z/OS.
  IBM's fork is a hard dependency and is not in hand.
- **ADR 0002's citation slot is empty.** The premise that IBM backing permits z/OS under
  Hercules is recorded as OWNER-ASSERTED. Until a citable instrument is filed, nothing
  published may state it.

## 8. Related docs

- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — §1 superseded by ADR 0002
- `docs/decisions/0002-zos-under-hercules-permitted-by-ibm-backing.md`
- `docs/decisions/0003-production-bridge-module-architecture.md`
- `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`
- `docs/hypotheses/002-s390x-port-equivalence.md` — resolved E-PORT-CLEAN on C1/C2/C4
- `docs/research-briefs/003-wto-wpl-layout-source-and-return-code-contract.md` — the
  outstanding item that unblocks `console.EncodeWPL`

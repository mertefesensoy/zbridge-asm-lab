# ebcdic · table-driven EBCDIC conversion, and the instruction that collapses the loop

Exercise 2 of zbridge-asm-lab. Converts between ISO-8859-1 (treated as ASCII-compatible)
and EBCDIC IBM-1047, the code page z/OS UNIX uses.

This is the exercise the roadmap singles out, because **every MVS parameter list carries
its text in EBCDIC** — including the WTO parameter list this project exists to build.

## ❯ What it proves

Table-driven byte translation through a 256-byte lookup table, in Plan 9 assembly, on two
architectures with very different answers to the same problem.

| Function | Contract |
|---|---|
| `AtoE(dst, src []byte)` | translates `len(src)` bytes of ISO-8859-1 into EBCDIC IBM-1047 |
| `EtoA(dst, src []byte)` | the inverse |

`dst` must have at least `len(src)` capacity. No validation is performed; bytes outside
the printable range map per the table. Side effect: writes `len(src)` bytes to `dst`.

## ❯ Mapping to s390x

**Implemented and tested as of 2026-07-25** — see
[`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`](../docs/evidence/E-L-s390x-port-qemu-2026-07-25.md).
All nine tests pass on `linux/s390x` under QEMU 10.2.1.

### The comparison

This is the architectural claim, and it needs no hardware to make:

| | amd64 | s390x |
|---|---|---|
| Instructions per unit of work | **7 per byte** | **2 per 256 bytes** |
| The sequence | `MOVBQZX` (load) · `MOVB` (table index) · `MOVB` (store) · `INCQ` ×2 · `DECQ` · `JNZ` | `MVC` (copy block) · `TR` (translate in place) |
| Real encoding | — | `d2 ff 20 00 40 00` · `dc ff 20 00 30 00` |
| Runtime-length tail | same loop, fewer iterations | `EXRL` patches the length byte of each |

Disassembled from the linked binary with `s390x-linux-gnu-objdump`:

```
16ae8e:  d2 ff 20 00 40 00    mvc  0(256,%r2),0(%r4)
16ae94:  dc ff 20 00 30 00    tr   0(256,%r2),0(%r3)
```

### Two corrections worth making before anyone makes them for you

**1. `TR` translates in place, so it is `MVC` *then* `TR`.** `TR`'s first operand is both
source and destination. With a two-buffer API — `AtoE(dst, src)` — the block must be copied
first. The per-byte loop is genuinely gone, but "one instruction replaces the whole loop"
is not literally true. Two block instructions per 256 bytes is the honest claim.

**2. Go's s390x assembler has no `TR` mnemonic.** `cmd/internal/obj/s390x/anames.go`
(go1.26.3) lists 729 mnemonics and the entire translate family is absent, as are `EX` and
`SVC`. So `TR` is emitted as literal bytes.

That is not a hack invented here — it is attested inside the Go distribution:
`golang.org/x/sys/unix/asm_zos_s390x.s` encodes `SVC 08` as `BYTE $0x0A; BYTE $0x08`, and
`internal/bytealg/indexbyte_s390x.s` encodes `SRST` as `WORD $0xB25E0082`.

**And the consequence that reaches past this exercise:** since there is no `SVC` mnemonic
either, **`SVC 35` — the project's endgame instruction — will have to be hand-encoded the
same way** (`BYTE $0x0A; BYTE $0x23`). This exercise turned out to be an unintentional
rehearsal of the exact technique the thesis needs.

### How six unchecked bytes are verified

Hand-encoded machine code gets no assembler check, so three independent checks replace it:

1. **Disassembly.** GNU `objdump` for s390x — which does know `TR` — decodes the bytes as
   `tr 0(256,%r2),0(%r3)`, exactly the instruction the source comments claim.
2. **A differential test.** `AtoELoop`/`EtoALoop` (declared in `ebcdic_s390x.go`) are
   byte-at-a-time s390x implementations of the same translation.
   `TestAtoETRMatchesLoop` asserts byte-identical output at lengths 0, 1, 2, 15, 16, 255,
   256, 257, 511, 512, 513 and 4096 — chosen to straddle the 256-byte block boundary in
   both directions.
3. **A table test and an overrun test.** `TestTRCoversEveryTableEntry` checks all 256 input
   values against the Go table directly; `TestTRDoesNotOverrun` pads the destination on
   both sides and asserts the padding is untouched.

## ❯ Honest benchmark note

`go test -bench=.` compares the assembly against a pure-Go reference. **On amd64 the Go
reference is competitive or faster**, and that is expected: the assembly here is a
deliberately simple byte-at-a-time loop that also pays ABI0 call overhead, while Go's
compiler is free to optimise the reference. Naive assembly is not automatically faster, and
saying so is part of the exercise.

**No s390x timing number is published, and none should be.** QEMU and Hercules implement
`TR` as a software loop, so timing it measures the emulator, not an IBM Z machine. This was
pre-registered as an explicit non-claim in
[H002](../docs/hypotheses/002-s390x-port-equivalence.md) *before* any code ran.

Per [ADR 0004](../docs/decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §4 the
timing table moves to **Phase 3c**, where hardware benchmarking is already planned. The
instruction-count comparison above ships in its place — it is the architectural claim the
roadmap actually wanted, and unlike a wall-clock number it is true independent of what
machine you measure on.

## ❯ Provenance and licence

`tables.go` derives from [`ibmruntimes/go-recordio`](https://github.com/ibmruntimes/go-recordio)
(BSD-3-Clause), and the mapping was verified against the ICU file
`ibm-1047_P100-1995.ucm` from unicode-org/icu. Attribution lives in
[`LICENSES.md`](LICENSES.md) and **must survive any refactor.**

A polished copy of this package ships in the production module as
[`zbridge/codepage`](../zbridge/codepage), where it is the one package that is finished and
useful on its own.

## ❯ Build and test

```powershell
go vet ./...
go test -v ./...
go test -bench=. -benchmem -run='^$' ./...
```

```powershell
# s390x: cross-compile on any host, then execute under QEMU
$env:GOOS='linux'; $env:GOARCH='s390x'; $env:CGO_ENABLED='0'
go vet ./...
go test -c -o ebcdic.test ./...
```

```bash
qemu-s390x ./ebcdic.test -test.v
s390x-linux-gnu-objdump -d ebcdic.test | grep -A 14 '<trbody>:'
```

The architecture-specific assembly sits behind filename build constraints
(`ebcdic_amd64.s`, `ebcdic_s390x.s`), so the toolchain selects the right file per target
`GOARCH` with no change to `ebcdic.go`.

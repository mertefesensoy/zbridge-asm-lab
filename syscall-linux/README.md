# syscall-linux · the syscall trap path

> Exercise 6 of `zbridge-asm-lab`. Issues Linux amd64 `SYSCALL` from Go assembly, with parameters in registers and explicit return-value evaluation.

## The One Idea

This exercise is the first direct trap into an operating system from Go assembly. The earlier exercises stayed inside the process: arithmetic, table translation, string and slice headers, registers, and byte loops. Here the assembly loads a service number and arguments into the Linux amd64 syscall registers, executes `SYSCALL`, and interprets the value returned in `AX`.

That is the same conceptual primitive the z/OS path needs for `SVC 35`: register setup, supervisor call, return-code evaluation. The opcode and ABI differ; the shape of the problem is the same.

## Functions

- `Getpid() (pid uintptr, errno uintptr)` issues Linux `getpid(2)` (`SYS_getpid = 39`). It has no arguments, so it is the minimal trap example.
- `Write(fd uintptr, p []byte) (n uintptr, errno uintptr)` issues Linux `write(2)` (`SYS_write = 1`). It proves parameter placement: `AX` gets the syscall number, `DI` gets `fd`, `SI` gets the slice data pointer, and `DX` gets `len(p)`.

Both functions treat values returned in `AX` from `-4095` through `-1` as `errno`, matching the shape used by Go's syscall assembly. On error, `n` or `pid` is `^uintptr(0)` and `errno` is positive.

## Linux amd64 Register Map

- `AX`: syscall number on entry, return value on exit
- `DI`: argument 1
- `SI`: argument 2
- `DX`: argument 3
- `R10`: argument 4
- `R8`: argument 5
- `R9`: argument 6

`SYSCALL` clobbers `CX` and `R11`, which is why Linux uses `R10` rather than `CX` for the fourth argument.

## Mapping to s390x and z/OS

On z/OS, WTO is not a Linux syscall. It is an MVS service reached with `SVC 35`. Parameters are placed in registers such as `R0` and `R1`, the `SVC` instruction traps to the supervisor, and the return code is read from a register afterward.

**Implemented and tested as of 2026-07-25** — `syscall_linux_s390x.s` is a real implementation and all three tests pass on linux/s390x under QEMU 10.2.1 (`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`).

Having both targets is more instructive than either alone:

| | amd64 | s390x |
|---|---|---|
| Service number | `AX` | `R1` |
| Arguments | `DI`, `SI`, `DX`, `R10`, `R8`, `R9` | `R2`–`R7` |
| Trap | `SYSCALL` | `SYSCALL`, which assembles to **`SVC 0`** |
| Result | `AX` | `R2` — the *first argument* register |
| `write` | 1 | **4** |
| `getpid` | 39 | **20** |

The numbers differ because a system-call table is a per-architecture contract, not a per-OS one; s390x inherits the numbering from the 31-bit s390 port. Verified against Go's own `runtime/sys_linux_s390x.s`.

Two things carry forward to the endgame. First, on s390x the result comes back in the register that carried argument one, so anything needing that argument afterwards must save it. Second — and this is the point of the exercise — **`SVC` appears here, on Linux.** Linux takes the service selector from `R1` and issues `SVC 0`; MVS encodes the selector in the instruction itself, so WTO is literally `SVC 35`. Note that Go's s390x assembler has no `SVC` mnemonic at all, so `SVC 35` will have to be hand-encoded as `BYTE $0x0A; BYTE $0x23`, exactly as `TR` is in the `ebcdic` exercise.

## Build and Test

On Linux amd64, run:

```powershell
go vet ./...
go test -v ./...
go test -bench=. -benchmem -run='^$' ./...
```

On non-Linux hosts, `go test` reports a skipped runtime test. From this Windows workstation, the useful verification is a Linux amd64 cross-compile:

```powershell
$env:GOOS='linux'; $env:GOARCH='amd64'; go vet ./...
$env:GOOS='linux'; $env:GOARCH='amd64'; go test -c -o C:\tmp\syscall-linux.test .
go tool objdump -s 'syscalllinux\.(Getpid|Write)' C:\tmp\syscall-linux.test
```

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

The `syscall_linux_s390x.s` file is a documented stub for Phase 1b. It exists to keep the roadmap explicit: this exercise proves the trap shape on Linux amd64 first, then the project swaps in the s390x/z/OS register convention and `SVC` instruction later.

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

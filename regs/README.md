# regs · hardware registers and stack frame layout

> Exercise 4 of `zbridge-asm-lab`. Reads hardware registers (`SP`, `BP`) in Plan 9 assembly and makes stack-frame layout observable.

## ❯ The one idea: there are two SPs

Go assembly has two stack pointers, and they are not the same thing:

- Bare `SP`, with no symbol, as in `MOVQ SP, AX`, is the hardware stack pointer register. It holds a real address.
- `name+0(SP)`, with a symbol and an offset, is the pseudo stack pointer: a virtual register the assembler resolves to a slot in the current function's frame. Its value is unrelated to the hardware `SP`.

Conflating the two is the most common Go-assembly mistake. The `add`, `ebcdic`, and `strmanip` exercises all addressed their arguments through the pseudo registers (`FP`); this exercise reads the hardware ones directly.

## ❯ Functions

- `GetSP() uintptr` returns the hardware stack pointer.
- `GetBP() uintptr` returns the hardware base pointer (`BP`), which Go maintains as the frame pointer on amd64.
- `GetFramedSP() uintptr` is `GetSP` declared with a 256-byte frame. Its `SP` is lower than `GetSP`'s by at least the frame size, which makes the cost of a declared frame visible.

## ❯ What the tests show

- `TestFrameSizeShiftsSP`: `GetSP()` and `GetFramedSP()` are called from the same frame, so the only difference between them is the 256-byte frame; the framed `SP` comes back at least 256 bytes lower. A declared frame size moves the hardware stack pointer.
- `TestStackGrowsDownward`: reading `SP` at recursion depth 8 versus depth 0 shows the deeper call has a lower `SP`. The stack grows toward lower addresses on amd64.

## ❯ Mapping to s390x

The register file is the headline difference. amd64 has named registers; s390x has sixteen general registers `R0` through `R15` and sixteen floating registers `F0` through `F15`, with fixed roles for the high ones:

- `R15` is the stack pointer in the s390x ABI, so `GetSP` becomes a read of `R15` (`MOVD R15, ...`).
- `R14` is the return address (link register); a `BASR` or `BRASL` call leaves the return point there, where amd64 puts it on the stack.
- `R13` is the save-area / base register in standard MVS linkage. There is no single frame-pointer register like amd64's `BP`; the save-area chain plays that role.

These are the same registers the WTO path touches: parameters go in `R0` and `R1`, `SVC 35` is issued, and the return code is read from `R15`. The `regs_s390x.s` file is a documented stub that fails the build if targeted, to be filled in during Phase 1b.

## ❯ Build and test

Run `go test -v ./...` and `go test -bench=. -benchmem -run='^$' ./...` from inside this directory. The benchmark measures the overhead of a single Go-to-assembly call, around 1 to 2 ns; it is a baseline, not something to optimize.

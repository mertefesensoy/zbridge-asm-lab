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

These are the same registers the WTO path touches: parameters go in `R0` and `R1`, and `SVC 35` is issued. **Implemented and tested as of 2026-07-25** (`docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`).

Two corrections came out of actually writing it.

**The API is not architecture-identical, and it should not be.** s390x has *no frame-pointer register*: Go assigns R13 to the goroutine pointer `g`, R14 to the link register, R15 to the stack pointer, and nothing to a frame pointer (`cmd/internal/obj/s390x/a.out.go`: `REGG = REG_R13`, `REGSP = REG_R15`). So `GetBP` is amd64-only, and s390x declares `GetLinkRegister` (R14) and `GetGReg` (R13) instead. Returning R13 from a function named `GetBP` would have kept one API at the cost of mislabelling a register in a teaching repository.

**The pseudo-register lesson gets sharper, not softer.** On amd64 the hardware read is `MOVQ SP, AX` — bare `SP`. On s390x, bare `SP` in Go assembly is *only* the pseudo stack pointer; the hardware register must be named `R15`. The two things amd64 spells with the same two letters have different names here.

The s390x tests are also stronger than the amd64 test they replace. `TestLinkRegisterIsARealReturnAddress` calls from two `//go:noinline` sites and requires the values to differ — a constant would pass a non-zero check but not this one. `TestGRegDiffersBetweenGoroutines` runs eight goroutines and requires eight distinct values, which is what proves R13 really holds `g`.

## ❯ Build and test

Run `go test -v ./...` and `go test -bench=. -benchmem -run='^$' ./...` from inside this directory. The benchmark measures the overhead of a single Go-to-assembly call, around 1 to 2 ns; it is a baseline, not something to optimize.

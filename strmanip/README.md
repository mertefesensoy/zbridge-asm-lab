# strmanip · string headers and the WTO parameter list

> Exercise 3 of `zbridge-asm-lab`. Reads the Go string header in Plan 9 assembly and builds the length-prefixed buffer shape that the WTO parameter list uses.

## ❯ What this exercise proves

- Reading the Go string header directly in assembly. A `string` is a two-word header (data pointer, length); a slice is three words (data, length, capacity). The argument offsets in the `.s` file differ because of this, and that difference is the new skill over the `ebcdic` exercise.
- Building a structured output buffer: a 2-byte big-endian length field followed by the payload. This is the shape of the WTO parameter list (a length field, then the EBCDIC message text), so the exercise doubles as a rehearsal for Phase 3b.

## ❯ Functions

- `StrLen(s string) int` reads the length word of the string header and returns it. It is equivalent to `len(s)`; it exists to isolate string-header access as its own step.
- `WrapLengthPrefixed(dst []byte, msg string) int` writes a 2-byte big-endian length header into `dst[0:2]`, copies the bytes of `msg` into `dst[2:]`, and returns `len(msg)+2`. `dst` must have capacity for at least `len(msg)+2` bytes; `msg` must be at most 65535 bytes.

## ❯ Why big-endian

z/Architecture is big-endian, so the 2-byte length is stored high byte first. On amd64, which is little-endian, the 16-bit value is split by hand: `SHRQ $8` for the high byte, then the low byte. The `TestWrapHighByte` case (length 300, which is `0x012C`, stored as `01 2C`) is the one that breaks if the split is wrong.

## ❯ Mapping to s390x

Two amd64 details collapse on s390x, the same way the `ebcdic` lookup loop collapses into `TR`:

- The big-endian length. `STH` (Store Halfword, Go asm `MOVH`) writes both bytes in big-endian order in one instruction, so the manual high and low byte split disappears.
- The byte copy. `MVC` (Move Characters) copies up to 256 bytes in one instruction, which covers every single-line WTO message, so the copy loop disappears too.

String-header access is the same shape; only the load mnemonic changes (`MOVD`). The `strmanip_s390x.s` file is a documented stub that fails the build if targeted, to be filled in during Phase 1b.

## ❯ Benchmark note

`go test -bench=.` compares the assembly against a pure-Go reference (`encoding/binary` big-endian put plus `copy`). On amd64 the Go reference is faster: `copy` lowers to an optimized, vectorized `runtime.memmove`, while this exercise uses a deliberately simple byte-at-a-time loop and pays the ABI0 call overhead. That is the honest and expected outcome on a platform with a mature optimizer. The architectural payoff is on s390x, where `MVC` does the copy in hardware. Naive assembly is not automatically faster, and saying so is part of the exercise.

## ❯ Mapping to WTO (Phase 3b)

The real WTO parameter list is richer than this buffer: the length field is followed by message-control flags before the text, and the length value follows IBM's convention. This exercise rehearses the length-prefix mechanics and the string-to-buffer construction. Phase 3b adds the flags field and the exact WTO header, and feeds the text through the `ebcdic` `AtoE` conversion first.

## ❯ Build and test

Run `go test -v ./...` and `go test -bench=. -benchmem -run='^$' ./...` from inside this directory. The architecture-specific assembly sits behind filename build constraints (`strmanip_amd64.s` now, `strmanip_s390x.s` in Phase 1b), so the toolchain selects the right file per target `GOARCH` with no change to `strmanip.go`.

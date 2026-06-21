# bytecmp · byte comparison and condition codes

> Exercise 5 of `zbridge-asm-lab`. Compares two byte sequences in Plan 9 assembly: memory loads from two pointers, a counted loop, and branching on the condition codes a byte compare sets.

## ❯ Functions

- `Equal(a, b []byte) bool` reports whether `a` and `b` are the same length and the same bytes. Matches `bytes.Equal`.
- `Compare(a, b []byte) int` returns `0`, `-1`, or `+1` comparing `a` and `b` lexicographically; bytes are compared as unsigned, and a shorter slice that is a prefix of the other sorts first. Matches `bytes.Compare`.

Matching the standard library means `bytes.Equal` and `bytes.Compare` serve as both the test oracle and the benchmark baseline.

## ❯ What this exercise proves

- Memory instructions: loading bytes from two independent pointers, including the indexed form `(SI)(R8*1)`.
- Loop control: a counted scan over `min(len(a), len(b))` bytes, with a length tiebreak when the common prefix matches.
- Condition codes: each byte compare sets the flags, and the code branches on them (`JLT`, `JGT`, `JNE`). The bytes are zero-extended into `0..255`, so the signed `JLT`/`JGT` give the same ordering as an unsigned byte compare.

## ❯ Mapping to s390x

Both functions collapse onto one instruction family, the same way the `ebcdic` loop collapses into `TR`: `CLC`, Compare Logical Characters. `CLC` compares two storage operands byte by byte and sets the condition code, `0` equal, `1` first operand low, `2` first operand high. So:

- `Equal` becomes a length check, then a single `CLC`; equal iff the condition code is `0`.
- `Compare` becomes a `CLC` over the common length, then a branch on the condition code for the ordering, with the same length tiebreak as the amd64 `bylen` path.

`CLC` carries its length in the instruction (up to 256 bytes); a length known only at run time uses `EX` or `EXRL` to patch the length byte, or a loop issuing `CLC` in 256-byte strides. The `bytecmp_s390x.s` file is a documented stub for Phase 1b.

## ❯ Benchmark note

`go test -bench=.` compares each function against its standard-library twin on a 256-byte worst-case input, two buffers that differ only in the last byte, so the whole length is scanned. On amd64 the standard library is far faster, roughly 15x to 25x, because `bytes.Equal` and `bytes.Compare` compare a machine word (or a SIMD vector) at a time, while this exercise compares one byte per iteration. That is the honest and expected result. The byte-at-a-time loop is the worst case on amd64, and `CLC` is exactly where it stops being one: the hardware does the whole scan in a single instruction. Naive assembly is not automatically fast, and a byte loop is the clearest example of why.

## ❯ Build and test

Run `go test -v ./...` and `go test -bench=. -benchmem -run='^$' ./...` from inside this directory. The architecture-specific assembly sits behind filename build constraints (`bytecmp_amd64.s` now, `bytecmp_s390x.s` in Phase 1b), so the toolchain selects the right file per target `GOARCH` with no change to `bytecmp.go`.

# zbridge-asm-lab

Lab notebook for the Go assembly to z/OS services bridge project.
Each subdirectory is one self-contained exercise on the path toward
calling MVS macros from Go assembly without C, following the pattern
established by `ibmruntimes/go-recordio`. Endgame:
`WTO(message string) error` via `SVC 35` on z/OS — no cgo.

**Phase 1 status: complete.** All exercises pass `go vet` and `go test`
on amd64 (`syscall-linux` runtime tests require Linux; it cross-compiles
clean from other hosts).

## Exercises

- `add/` · toolchain checkpoint: prove Plan 9 assembly compiles, links,
  tests, and benchmarks on amd64.
- `ebcdic/` · table-driven byte translation between ISO-8859-1 and
  EBCDIC IBM-1047, the conversion primitive every MVS macro call
  needs at its parameter list boundary. Includes an s390x stub
  documenting the TR-instruction collapse.
- `strmanip/` · Go string headers in assembly and the length-prefixed
  buffer (2-byte big-endian length + text) that rehearses the WTO
  parameter list shape. s390x stub documents the STH/MVC collapse.
- `regs/` · hardware registers vs. pseudo-registers: bare `SP` against
  `name+0(SP)`, `BP` frame chains, and observable stack-frame layout.
  s390x stub documents the R13/R14/R15 MVS linkage conventions.
- `bytecmp/` · byte-sequence comparison with condition-code branching,
  oracle-tested against `bytes.Equal`/`bytes.Compare`. s390x stub
  documents the CLC collapse.
- `syscall-linux/` · the trap: Linux `SYSCALL` from Go assembly with
  parameters in registers and return-value evaluation — the dress
  rehearsal for `SVC 35`.

## Docs

- [Interactive module explorer](docs/interactive/zbridge-module-explorer.html) —
  open in a browser: annotated assembly for every exercise, live demos
  (including the real IBM-1047 translation table), and the WTO pipeline map.
- [Mainframe baseline strategy](docs/mainframe-baseline-strategy.md) —
  the operating plan for when z/OS access arrives: Day-0 checklist,
  T0→T3 test ladder, what gets skipped and what never does.
- [Codex handover](docs/codex-handover.md) — self-contained project
  state and guardrails for AI-assisted continuation.
- [Implementation docs](docs/implementations/) — one doc per meaningful
  change (documentation-first workflow).

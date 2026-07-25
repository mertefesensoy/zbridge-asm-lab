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

## Emulation program (from 2026-07-25)

z/OS cannot legally run under Hercules, so the project does not try. Instead
Hercules is used as a **two-track laboratory** that retires specific unknowns
before mainframe access:

- **Track M** — MVS 3.8j via TK5. Real `SVC 35`, real operator console. The
  only available oracle for *"is our WTO parameter list byte-correct?"*
- **Track L** — Linux s390x, time-boxed. Real z/Architecture execution of Go
  assembly. QEMU s390x carries the fast inner loop for the Phase 1b port.

The payoff is rung **E3**: Go cannot run on MVS 3.8j, but the *bytes* can cross.
A parameter list built by Go on the laptop, fired at a genuine `SVC 35`, retires
four of the six steps in the WTO endgame before any mainframe time is spent.

Nothing produced under emulation is presented as a z/OS result, and no
performance number is taken from an emulator.

See [ADR 0001](docs/decisions/0001-emulation-strategy-hercules-two-track.md).

## Documentation

**[`zbridge-asm-roadmap.pdf`](zbridge-asm-roadmap.pdf)** (repo root, 8 pages) is the
project mandate — phase definitions, the WTO endgame rationale, the risk register,
and the open questions for the mentor. **[`docs/goal-prompt.md`](docs/goal-prompt.md)**
is the operating doctrine derived from it.

| Directory | Contents |
|---|---|
| `docs/decisions/` | Architecture Decision Records — numbered, with explicit scope limits and reopen conditions |
| `docs/hypotheses/` | Pre-registered falsifiable claims with decision rules fixed before evidence |
| `docs/evidence/` | Captured rung outputs with mandatory provenance headers |
| `docs/runbooks/` | Executable setup and operation procedures |
| `docs/research-briefs/` | Scoped research requests (the Gemini interface) |
| `docs/implementations/` | Per-change implementation docs |
| `research/` | Research returns, stored verbatim with sources |

Start with [ADR 0001](docs/decisions/0001-emulation-strategy-hercules-two-track.md)
and [the baseline strategy](docs/mainframe-baseline-strategy.md).
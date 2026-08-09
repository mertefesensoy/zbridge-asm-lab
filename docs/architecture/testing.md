# How the tests work

**Read [`README.md`](README.md) first**, and ideally
[`evidence-ladder.md`](evidence-ladder.md) — several of the categories below only make
sense in light of the E0→E3 rungs.

This repository uses the word "test" for seven genuinely different things, and
conflating them is easy to do and misleading once done. This document walks through
each one: what it actually checks, what it can't check, and a real example from the
codebase. Verified counts (§8) are from a fresh run during this session, not recalled
from an older document.

---

## 0. The seven kinds, at a glance

| # | Kind | Runs via | Answers |
|---|---|---|---|
| 1 | Unit tests | `go test` | Does this one function do what it claims, in isolation? |
| 2 | Oracle tests | `go test` | Does our output match bytes a real IBM macro produced? |
| 3 | Differential / parity tests | `go test` (s390x only) | Do the hand-encoded assembly path and a plain-Go reference agree? |
| 4 | Integration tests | `go test` (Linux only) | Does this code produce the same result as the real operating system? |
| 5 | The cross-compile gate | `go vet` / `go build` | Does this compile correctly for a target that can't be executed here? |
| 6 | Benchmarks | `go test -bench` | How fast — and where is that number allowed to be quoted? |
| 7 | Evidence rungs | manual/scripted, not `go test` at all | Did a real (emulated) mainframe actually accept this? |

---

## 1. Unit tests — the ordinary kind

Pure logic, no OS interaction, no assembly-specific concerns: does one function behave
correctly across its input space? Most of this repository's tests are this kind.

Example, from `zbridge/errors_test.go` — the test that encodes the module's single
most important documented finding (see
[`zbridge-module.md`](zbridge-module.md) §6):

```go
func TestHasCodeDistinguishesNoReturnCodeFromZero(t *testing.T) {
    noCode := &Error{Op: "console.WTO", Service: "SVC 35", Err: ErrServiceFailed}
    zeroCode := &Error{Op: "console.WTO", Service: "SVC 35", Code: 0, HasCode: true, Err: ErrServiceFailed}
    // ... asserts the two render differently, so the distinction is visible in logs,
    // not just in the type system ...
}
```

Nothing here touches an emulator, a build tag, or an external oracle — it's checking
that a Go type behaves the way its own doc comment says it does. `console/console_test.go`
is the same kind: message validation, the platform-boundary error, and the route/descriptor
bit-mask arithmetic are all pure functions of their inputs, so they're tested as pure
functions.

**Why this category matters even in a project this evidence-heavy:** not everything
needs a mainframe, and pretending it does would be its own kind of dishonesty. The
bit-mask arithmetic in `console/options.go` is exactly this — see
[`zbridge-module.md`](zbridge-module.md) §3.5 for why *what* the mask contains was
never actually in doubt, only *where in the parameter list* it goes.

---

## 2. Oracle tests — bytes checked against a real machine's output, not a spec

This is the strongest, and most unusual, test in the repository:
`zbridge/console/wpl_oracle_test.go`.

**What makes it different from an ordinary unit test:** an ordinary test's expected
value comes from reasoning about what the code *should* do. This test's expected value
is not reasoning — it's a literal transcription of bytes a **real IBM `WTO` macro**
produced when assembled on MVS 3.8j (rung E1, see
[`evidence-ladder.md`](evidence-ladder.md) §3):

```go
// macroOracle is the exact parameter list IBM's macro generated for
// WTO 'ZBRIDGE TEST HELLO',MF=L — 18 characters of text, length field 0x0016.
const macroOracle = "0016" + "0000" +
    "E9C2D9C9C4C7C540" + "E3C5E2E340C8C5D3" + "D3D6"

func TestEncodeWPLMatchesIBMMacro(t *testing.T) {
    want := mustDecode(t, macroOracle)
    got, err := EncodeWPL("ZBRIDGE TEST HELLO")
    // ... assert got == want, byte for byte ...
}
```

If `EncodeWPL` is ever changed in a way that would produce a parameter list a real
`SVC 35` implementation would reject, **this test fails on a laptop, immediately, with
no emulator involved** — because rung E3 already proved these exact bytes are what a
real system accepts, and this test is what keeps that proof attached to the code
instead of living only in a historical evidence file. This is the practical payoff of
the design choice described in [`zbridge-module.md`](zbridge-module.md) §3.2: separating
byte *construction* from byte *transmission* is what makes a laptop-runnable regression
test for a mainframe-verified fact possible at all.

A second test in the same file, `TestLengthFieldIsTextPlusFour`, generalizes the check
across every valid message length rather than just the two lengths the macro happened
to be run at — so a regression at, say, 50 characters would be caught even though the
macro was only ever assembled at 18 and 38.

**The oracle-test pattern is the direct, executable consequence of hypothesis
[H001](../hypotheses/001-mvs38j-svc35-wto-oracle.md)'s methodology**: Line 2 of that
hypothesis calls the MVS 3.8j assembler listing "ground truth" precisely because it's
what a real system actually did, not a claim about what it should do — and once that
ground truth existed, encoding it as a Go test was the natural next step.

---

## 3. Differential / parity tests — does the hand-encoded assembly agree with a plain-Go reference?

These exist specifically because of a recurring fact in this project: **Go's s390x
assembler has no mnemonic for several instructions this project needs** (`TR`, and
eventually `SVC` — see [`zbridge-module.md`](zbridge-module.md) §2 and
[`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md) §6). When an instruction is
hand-encoded as raw bytes rather than written as a mnemonic, "does this compile" proves
nothing about "does this do what I think it does" — so every hand-encoded path gets a
second, independent implementation in plain Go, and a test that the two agree across
values chosen specifically to hit the implementation's edge cases.

Example, from `ebcdic/ebcdic_diff_s390x_test.go` (and its near-identical twin,
`zbridge/codepage/codepage_diff_s390x_test.go`):

```go
// diffLengths covers the boundaries that matter to the s390x implementation:
// the empty case, the single-byte EXRL case, either side of the 256-byte block
// stride, and a multi-block buffer with a partial tail.
var diffLengths = []int{0, 1, 2, 15, 16, 255, 256, 257, 511, 512, 513, 4096}

func TestAtoETRMatchesLoop(t *testing.T) {
    for _, n := range diffLengths {
        // ... AtoE (hand-encoded TR) vs AtoELoop (plain Go) must agree, byte for byte ...
    }
}
```

The lengths aren't arbitrary — 255/256/257 specifically straddle the 256-byte block
`TR` operates on in one shot, which is exactly where an off-by-one in the block/tail
arithmetic would surface. A companion test, `TestTRCoversEveryTableEntry`, checks
something slightly stronger: not that the assembly agrees with the *loop*, but that it
agrees with the underlying *translation table* directly, for all 256 possible input
bytes at once.

**This is the executable form of hypothesis
[H002](../hypotheses/002-s390x-port-equivalence.md)'s claim C2** ("instruction-collapse
fidelity" — the single-instruction s390x form must produce byte-identical output to the
loop it replaces). The same *shape* of test — real implementation vs. a stdlib or
reference implementation on the same inputs — also appears without a special build tag
in `bytecmp` (`TestEqualAgainstStdlib`, `TestCompareAgainstStdlib`, checked against
Go's own `bytes.Equal`/`bytes.Compare`) and in `strmanip`
(`TestWrapMatchesReference`, checked against `encoding/binary`). Same idea, whichever
side of the comparison happens to be "the assembly" varies by module.

---

## 4. Integration tests — code checked against the real operating system

Everything above runs entirely inside the Go process under test. `syscall-linux` is
different: its tests submit real system calls to a real kernel and check the result
against what the operating system itself reports, which is what makes it this
repository's closest thing to a conventional integration test.

```go
func TestGetpidMatchesOS(t *testing.T) {
    pid, errno := Getpid()               // our hand-written syscall wrapper
    if want := uintptr(os.Getpid()); pid != want {  // the standard library, asking the same kernel
        t.Fatalf("Getpid = %d, want %d", pid, want)
    }
}

func TestWriteToPipe(t *testing.T) {
    r, w, _ := os.Pipe()                 // a REAL kernel pipe, not a mock
    Write(w.Fd(), []byte("syscall-linux\n"))
    // ... read the real bytes back out of the real pipe and compare ...
}
```

`TestWriteBadFDReturnsErrno` deliberately submits an invalid file descriptor and checks
that the kernel's real error path is surfaced correctly — the point isn't just "does
the happy path work," it's "does our raw syscall wrapper handle what the kernel
actually does on failure."

This module is gated to `//go:build linux && (amd64 || s390x)`; on any other platform,
a separate file (`syscall_nonlinux_test.go`) supplies a single test that calls
`t.Skip(...)` with an explicit reason, rather than the package silently having zero
tests and looking untested. **Skipped-with-a-reason and silently-absent look identical
in a test count; they are not the same thing**, and this repository is careful to keep
that distinction visible.

**What this category does *not* reach:** `syscall-linux`'s s390x form validates the
Linux `SVC` trap shape — U1 territory. It says nothing about `SVC 35` or WTO, which is
U2's job and a different oracle entirely (MVS 3.8j, not Linux). H002's own "threats to
validity" section names this explicitly, precisely so a passing `syscall-linux` on
s390x never gets casually read as evidence about the WTO path.

---

## 5. The cross-compilation gate — not a test, but a mandatory check before one

```bash
GOOS=linux GOARCH=s390x go vet ./...
GOOS=linux GOARCH=s390x go build ./...
```

This has no assertions and executes no code — it's a **build-time** check, and it's
still one of the most valuable gates in the repository, for a specific reason:
`go vet` mechanically verifies the `$frame-args` contract (the argument-size and
frame-size annotations every Plan 9 assembly function must declare — see
[`README.md`](README.md)'s note on the bug class this repo exists to prevent). A
signature change that gets the s390x argument offsets wrong is caught **here**, before
anything is ever run under an emulator, because `go vet` checks the contract
mechanically rather than trusting the comment above the function.

`add/` is the one module that's *expected* to fail this gate — it declares a bodyless
function served only by an amd64 assembly file and was never meant to be ported. A
repo-wide cross-build failing on `add/` specifically is not a regression; it's the
gate working correctly on a module that was never a candidate for it. The five modules
that must pass clean are `ebcdic`, `strmanip`, `regs`, `bytecmp`, `syscall-linux`, plus
the `zbridge` module itself.

---

## 6. Benchmarks, and the rule that governs every number they produce

```go
func BenchmarkAtoEAsm(b *testing.B) { /* the hand-encoded TR path */ }
func BenchmarkAtoEGo(b *testing.B)  { /* atoeRef — a plain Go loop, same inputs */ }
```

Benchmarks exist in this repository, and they run cleanly with `go test -bench=.`. What
they must **never** do is produce a number presented as this project's actual
performance claim, unless the machine underneath them is real. `docs/goal-prompt.md`
§4.3 states this as one of the doctrine's non-negotiable rules, and
[H002](../hypotheses/002-s390x-port-equivalence.md) pre-registered the reason *before*
any code ran: **QEMU and Hercules both implement `TR` as a software loop**, so timing
it under either one measures the emulator's interpreter, not an IBM Z CPU. A number from
an emulator could show a speedup, a slowdown, or nothing at all, and none of the three
would say anything true about real hardware.

**What ships in its place, honestly, until real hardware is available:** an
instruction-count and encoding comparison — not a timing number at all —
in `ebcdic/README.md`:

| | amd64 | s390x |
|---|---|---|
| Per unit of work | 7 instructions **per byte** | 2 instructions **per 256 bytes** |
| Real encoding | — | `d2 ff 20 00 40 00` (`mvc`) · `dc ff 20 00 30 00` (`tr`) |

This is the architectural claim the roadmap actually wanted, and it needs no hardware
to be true. The timing table itself is deferred to Phase 3c, on real hardware, per
[ADR 0004](../decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) §4.

---

## 7. Evidence rungs — when the "test" is a real (emulated) machine, not `go test`

The E0→E3 rungs described in full in [`evidence-ladder.md`](evidence-ladder.md) are
tests in every meaningful sense — they have a clear pass/fail gate, defined in advance,
and a result — but they do not run through `go test`, have no assertions in Go source,
and cannot be re-run by typing one command. Rung E3's gate, for instance, is: *does a
real `SVC 35` on MVS 3.8j accept the exact bytes `EncodeWPL` produced, and does the
message appear on the console?* The "test runner" is a JCL job submitted over a TCP
socket to an emulated mainframe; the "assertion" is grep-ing a captured console log for
the expected text; the result is written up as a Markdown file in `docs/evidence/` with
a mandatory provenance header, not as a green checkmark in a CI run.

**Why this category has to exist, and can't be collapsed into "integration test":** an
integration test (§4) checks code against a system the test itself can also *construct*
a known-good comparison from (the real Linux kernel is right there, and `os.Getpid()`
can ask it the same question). An evidence rung checks code against a system — MVS
3.8j's `SVC 35`, standing in for z/OS's — where **no independent way to compute the
right answer exists** other than asking the system itself. That's a fundamentally
different evidentiary situation, which is why these results get the heavier apparatus
(provenance headers, hypothesis pre-registration, an evidence file per rung) that an
ordinary `go test` result doesn't need.

`docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` and
`docs/hypotheses/002-s390x-port-equivalence.md` are the pre-registered decision rules
that say, in advance, what result would count as each rung passing — written before any
evidence existed, specifically so a rung's outcome can't be quietly reinterpreted after
the fact to look better than it was.

---

## 8. Verified counts (this session, 2026-08-09)

Run directly, not recalled from an older document — a clean methodology (top-level test
functions only, matching `^--- (PASS|FAIL)`, excluding subtests) was used specifically
so this count is reproducible by anyone reading this file:

| Module | Top-level tests | Result |
|---|---|---|
| `zbridge` (all packages) | 27 | all pass |
| `ebcdic` | 5 | all pass |
| `strmanip` | 5 | all pass |
| `regs` | 4 | all pass |
| `bytecmp` | 4 | all pass |
| `syscall-linux` (windows/amd64 — build-tag skipped) | 0 | n/a by design, see §4 |
| **Lab total (Windows-buildable)** | **18** | all pass |

`go vet` is clean on every module above. `GOOS=linux GOARCH=s390x go build` is clean on
all six (`ebcdic`, `strmanip`, `regs`, `bytecmp`, `syscall-linux`, `zbridge`); `add/`
correctly fails, as expected (§5). The 29-test `linux/s390x` figure quoted elsewhere in
this repository is a *different* number measuring a different thing — it's the count of
tests actually **executed** under QEMU emulation (§3/§7), not the count of Windows-buildable
top-level test functions in this table; the two need not match and don't.

**A small, deliberate note on precision:** an earlier repository document recorded "19"
for the lab total; this session's clean re-run gets 18. That's a one-test difference
accumulated over roughly ten days of normal development, not a discrepancy anyone needs
to chase down — but it's recorded here rather than silently overwritten, in keeping with
this project's own rule that corrections are visible, not quiet.

---

## 9. How to run any of this yourself

Full, dependency-by-dependency instructions — including the WSL2/QEMU/Hercules setup
required for §3's s390x runs and §7's evidence rungs — are in `RUN.md` at the
repository root. The short version, from each module's own directory:

```powershell
go vet ./...
go test ./...
go test -bench=. -benchmem -run='^$' ./...
```

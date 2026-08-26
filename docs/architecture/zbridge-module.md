# The `zbridge/` module, package by package

**Read [`README.md`](README.md) first** if you haven't — this document assumes you
already know what WTO, `SVC 35`, and the U1/U2/U3 unknowns are.

This document explains the actual Go module the project ships: what each package does,
what it depends on, and — this is the part that surprises first-time readers — **why
several packages are deliberately, honestly unfinished**, and how that's a designed
property of the codebase rather than a gap someone forgot to fill in.

---

## 1. Two codebases in one repository, and why they don't touch

The repository root holds six small teaching modules (`add/`, `ebcdic/`, `strmanip/`,
`regs/`, `bytecmp/`, `syscall-linux/`), each with its **own, independent `go.mod`**.
`zbridge/` is a **seventh, separate Go module**, also with its own `go.mod`, living in
its own subdirectory.

This isn't an accident of growth — it's a standing project rule: **per-exercise
`go.mod`, no cross-imports.** The lab modules exist to rehearse one concept each in
isolation (what does a Go string header actually look like in memory? what does the
hardware stack pointer look like vs. the pseudo-register named `SP`?), and letting them
import each other — or letting the production module depend on lab code — would let a
teaching shortcut quietly become a load-bearing dependency. `zbridge/` is allowed to be
*informed* by what the lab modules taught (it reuses the same hand-encoding technique
`ebcdic` proved out, for instance) without literally depending on their code.

**Module identity**, from
[ADR 0003](../decisions/0003-production-bridge-module-architecture.md) §2:

| Property | Value |
|---|---|
| Module path | `github.com/mertefesensoy/zbridge` |
| License | BSD-3-Clause (matches `ibmruntimes/go-recordio`, the project's architectural blueprint) |
| cgo | Never — see [`README.md`](README.md) §4 |
| Minimum Go | 1.26 |

---

## 2. `codepage` — the one finished package

```
zbridge/codepage/
    codepage.go     AtoE, EtoA — the public API
    codepage_s390x.go
    tables.go       BSD-3-Clause-derived translation tables
```

Converts between UTF-8/ISO-8859-1 and **EBCDIC IBM-1047**, the character encoding z/OS's
Unix System Services uses (see [`README.md`](README.md) §2.5 if "why does a mainframe
need a different alphabet" wasn't already clear).

```go
func AtoE(dst, src []byte)   // ASCII-ish -> EBCDIC
func EtoA(dst, src []byte)   // EBCDIC -> ASCII-ish
```

**Why this is the only package in the module that's actually done:** every other
package in `zbridge/` is blocked on at least one of U1/U2/U3. `codepage` isn't, because
character conversion depends on a documented lookup table, not on an undocumented
mainframe parameter-list format. Its correctness question is "does our assembly
implement this table correctly on s390x" — pure U1 — and U1 was retired on 2026-07-25
(see [`evidence-ladder.md`](evidence-ladder.md)).

**How it's implemented**, in Plan 9 assembly on both supported architectures:

- **amd64:** a straightforward byte-at-a-time lookup loop.
- **s390x:** `MVC` (block copy) then `TR` (Translate — one instruction translates up to
  256 bytes through a table in a single shot), with `EXRL` handling whatever length
  isn't a clean multiple of 256.

The catch, and it's worth internalizing early because the same shape reappears for
`SVC 35`: **Go's s390x assembler has no `TR` mnemonic at all.** `cmd/internal/obj/s390x/anames.go`
lists 729 instruction names in go1.26.3, and the entire translate family is absent. So
`TR` is emitted as **literal instruction bytes** (`BYTE $0xDC; ...`) rather than written
as an assembly mnemonic, and its correctness is proven three independent ways rather
than trusted on sight: GNU `objdump` disassembles the hand-written bytes back into `tr
0(256,%r2),0(%r3)` (confirming the *encoding* is right), a differential test runs the
`TR` path against a plain byte-loop reference implementation at eleven different buffer
lengths (confirming the *behavior* matches), and a third test walks all 256 possible
input bytes through the table (confirming *completeness*). Full detail and the actual
disassembly output: `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`.

One more correction worth knowing, because an earlier version of the roadmap got it
subtly wrong: `TR` translates **in place** (its first operand is both source and
destination), so a two-buffer API like `AtoE(dst, src)` costs **`MVC` then `TR`** — two
block instructions per 256 bytes, not the single instruction "TR" might suggest in
isolation. The per-byte loop is genuinely gone on s390x; "one instruction" was never
quite literally true.

**Provenance:** the tables are derived from `ibmruntimes/go-recordio` (BSD-3-Clause) and
cross-checked against ICU's official `ibm-1047_P100-1995.ucm` mapping file. Attribution
lives in `codepage/LICENSES.md` and — per the project's attribution rule — must survive any
future refactor.

---

## 3. `console` — the thesis endgame

```
zbridge/console/
    doc.go
    wpl.go              EncodeWPL, FormatDC — the parameter-list encoder (pure Go)
    wpl_oracle_test.go  bytes checked against a REAL IBM macro's output
    options.go          RouteCode, DescriptorCode, WithRoute, WithDescriptor
    console.go          WTO, WTOR — the public functions
    console_zos.go       //go:build zos && s390x — the real SVC 35 call
    console_stub.go      //go:build !(zos && s390x) — every other platform
```

This is the package the entire project exists to produce. Its full byte-level mechanism
— exactly how a Go string becomes a line on an operator's screen — has its own document:
**[`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md)**. This section covers the
package's *shape*, not its byte layout.

### 3.1 The public contract

```go
func WTO(msg string, opts ...Option) error
```

- **Input:** `msg`, 1 to `MaxTextLen` (124) bytes, no control characters.
- **Side effect (on z/OS, once U3 is retired):** one line appears on the operator
  console(s) selected by the route codes.
- **Returns:** `nil` on success, otherwise a `*zbridge.Error` (§6 below).

`WTOR` (WTO **with Reply** — wait for the operator to type something back) is declared
alongside it because [ADR 0003](../decisions/0003-production-bridge-module-architecture.md)
commits to declaring the whole console family's surface, but it's strictly harder than
`WTO` — it needs a reply buffer reachable by the operating system, which is U3 territory
— so today it always returns an error explaining exactly that, rather than pretending to
work.

### 3.2 `EncodeWPL` — pure Go, testable anywhere, and the seam the whole design pivots on

```go
func EncodeWPL(msg string, opts ...Option) ([]byte, error)
```

This one function is arguably the most carefully-placed piece of code in the module.
It takes a message and turns it into the exact byte sequence `SVC 35` expects — and it
is **pure Go**, with no assembly, no build tags, and no platform dependency. It runs
identically on a Windows laptop, in CI, or on z/OS itself.

That's deliberate, not incidental. [ADR 0001](../decisions/0001-emulation-strategy-hercules-two-track.md)
§6 defines the project's "E3" proof rung as: *a Go program, running anywhere, emits the
exact bytes it intends to hand to `SVC 35`; those bytes get embedded into a mainframe
assembler program; a real `SVC 35` on a real (if ancestor) system is handed them.* For
that rung to be possible at all, the byte-construction logic has to be **separable**
from the byte-*sending* logic — which is exactly what splitting `EncodeWPL` (construct)
from `issueWTO` (send, §3.3) buys. `console/wpl.go`'s own header comment states this
plainly: *"the module is designed so that E3's artifact is a first-class exported
capability rather than a one-off script."* And it worked: rung E3 (see
[`evidence-ladder.md`](evidence-ladder.md)) is this function's real output, hex-dumped
into an MVS assembler program, accepted by a genuine `SVC 35`.

### 3.3 The platform split — `console_zos.go` vs. `console_stub.go`

Go lets you compile different files depending on the build target, using a `//go:build`
constraint at the top of the file. `console` uses this to draw a hard line between "the
real thing" and "everywhere else":

```go
// console_zos.go
//go:build zos && s390x

// console_stub.go
//go:build !(zos && s390x)
```

Both files implement the same unexported function, `issueWTO(wpl []byte) error` — the
step that actually hands bytes to `SVC 35`. Only one of them is ever compiled into any
given build. Notice the constraint is `zos && s390x` **together**, never `s390x` alone:
`linux/s390x` is a real, useful, frequently-built target in this project (it's where U1
was retired, under QEMU), and it must never be mistaken for z/OS just because the CPU
architecture matches.

- **`console_stub.go`** compiles on every platform that isn't z/OS-on-s390x (which today
  is *every* platform, since z/OS-on-s390x can't be built at all — see §3.4) and returns
  a typed "wrong machine" error. This is what lets the module be `go build`-able and
  `go test`-able on an ordinary laptop.
- **`console_zos.go`** contains the shape of the real call — but see §3.4 for why "the
  shape" is all it can be right now.

### 3.4 The file nobody can compile yet, and why it exists anyway

`console_zos.go` is compiled by **nothing today**. Not because of a bug — because
upstream Go's toolchain has no `zos/s390x` build target at all. This was verified
directly against the toolchain source (not assumed): `go tool dist list` on go1.26.3
lists `linux/s390x` and no `zos/s390x`; the string `"zos"` does appear in
`internal/syslist` (so the `//go:build zos` constraint at least *parses*), but
`internal/platform`'s list of supported targets doesn't include it. IBM maintains its
own fork of Go that *can* target z/OS, and obtaining and building that fork is now this
project's critical external dependency (see `docs/roadmap-2026-27.md` §1).

So why write a file nothing can compile? Because **the shape of the call is knowable
now, and writing it down in code — including exactly which parts are still missing — is
more honest and more useful than leaving it as a paragraph of prose.** The file's own
comments enumerate precisely what's still needed beyond a working toolchain: the WPL
byte layout (retired — see [`evidence-ladder.md`](evidence-ladder.md)), below-the-bar
storage for the parameter list (open — §5 below), and the addressing-mode switching
around the call (open — Phase 2 of the roadmap, not yet started). The function's body
today is a typed error plus a comment showing the intended real path — never a guess
dressed up as an implementation.

### 3.5 Route and descriptor codes — what's verified vs. what isn't

`options.go` defines `RouteCode` (which console groups see the message) and
`DescriptorCode` (what *kind* of message it is — a status update, an immediate action
request, and so on), plus `WithRoute(...)` / `WithDescriptor(...)` functional options.
The bit-mask arithmetic for turning a list of codes into a 16-bit mask
(`routeMask`/`descriptorMask`) is fully implemented and tested on every platform,
because *what* the mask should contain is documented and stable — that part never
depended on an unretired unknown.

**But calling `WTO` with either option currently returns an error, not a result.** The
reason is precise, not vague caution: the macro-expansion evidence that revealed the
WPL byte layout (see [`wpl-svc35-mechanism.md`](wpl-svc35-mechanism.md)) *also* revealed
that adding routing/descriptor codes appends extra bytes after the message text and
flips one bit in the flags field to signal their presence — but research brief 003
could not find a citable, bit-by-bit table for the other fifteen flag bits, so nobody
can currently say with a straight face what the *other* bits mean. `EncodeWPL` refuses
routed calls outright (`ErrLayoutUnverified`) rather than emitting a parameter list that
*might* be missing something the caller asked for. **Refusing beats guessing, and
refusing beats silently dropping the request** — an operator-facing message going to
the wrong console group, or losing its urgency flag, is a worse failure than a
Go error return.

---

## 4. `internal/svc` — the raw supervisor-call primitive

```
zbridge/internal/svc/
    svc.go        Supported const; extensive doc comment on the encoding + linkage
    svc_zos.go    //go:build zos && s390x — Call35
    svc_stub.go   //go:build !(zos && s390x)
```

This package is `internal` — unreachable from outside the module — on purpose. A
supervisor call issued with a malformed parameter list is a good way to disturb a
system other people depend on, so nothing outside the module gets to attempt one
directly; every caller goes through `console`, which validates first.

```go
func Call35(wpl *byte) (rc int32, hasCode bool)
```

Two things stand out about this signature. First, **`SVC 35` itself has no Go mnemonic**
— exactly the same situation as `TR` in §2 — so the real assembly body (not yet
written; see §3.4) will hand-encode it as `BYTE $0x0A; BYTE $0x23`, following the exact
precedent already inside the Go distribution itself
(`golang.org/x/sys/unix/asm_zos_s390x.s` encodes `SVC 08` the same way). Second, the
**two return values, not one**: `hasCode` exists because whether the service actually
puts anything meaningful in the return register is *itself* an open, system-dependent
question — see §6 below. "Nothing was returned" has to be a state the type system can
express, not something silently indistinguishable from a return code of zero.

---

## 5. `internal/storage` and `internal/linkage` — the parts that need real z/OS

```go
// internal/storage
func Malloc31(n int) (buf []byte, free func(), err error)
```

A parameter list handed to a z/OS service has to live somewhere the service can
actually address — specifically, in this case, **"below the bar"**: below the 2 GB
address boundary, in 31-bit addressing mode (AMODE 31). An ordinary Go slice makes no
such promise; Go's garbage collector can move heap objects around, and nothing stops
the runtime from placing a slice's backing array above the bar. Handing `&slice[0]`
straight to a mainframe service would be wrong — and wrong *silently*, which is worse
than wrong loudly, because it might work by coincidence in testing and fail
unpredictably later.

`Malloc31` is where this module borrows exactly one thing from IBM's Language
Environment (LE) runtime — the pattern `ibmruntimes/go-recordio` already established.
It's the project's **one LE touchpoint**, and it's the precise, literal meaning of "no
LE dependency beyond `Malloc31`" in the project's mission statement. Today it
unconditionally returns "unsupported," because below-the-bar allocation is z/OS
behavior through and through — QEMU has no notion of it, and MVS 3.8j's equivalent
concept (the 16 MB line, not the 2 GB bar) is close enough to be informative but not
close enough to test against. This is squarely **U3** — nothing emulates it.

`internal/linkage` doesn't implement anything yet either — it *documents* the single
most dangerous fact in the whole codebase, in a form both prose and future assembly code
can point at:

| Register | What standard MVS linkage expects | What Go's s390x ABI already uses it for |
|---|---|---|
| R1 | address of the parameter list | general purpose |
| **R13** | address of a 72-byte save area | **the goroutine pointer, `g`** |
| R14 | return address | the link register (these two agree) |
| **R15** | entry point in, return code out | **the stack pointer** |

R13 and R15 are used for **completely different purposes** by MVS calling convention
and by Go's own s390x ABI (`cmd/internal/obj/s390x/a.out.go`: `REGG = REG_R13`, `REGSP =
REG_R15` — not folklore, read directly from the toolchain source). Any Go assembly
routine that sets up MVS linkage to call a service **must save both registers, do the
mainframe-style call, and restore both before returning to Go code** — clobber R13 and
the Go runtime loses track of which goroutine is running; clobber R15 and there is no
stack. This is exactly the failure mode the `regs/` lab exercise (see
[`README.md`](README.md) §6) exists to make directly observable from a Go test, rather
than something you find out about from a crash in production.

---

## 6. The error model — why so much of this module "fails on purpose," and how

Every operation in `zbridge/` that can't run yet returns a single structured error type
rather than panicking, returning a zero value, or silently no-op'ing:

```go
type Error struct {
    Op      string  // "console.WTO"
    Service string  // "SVC 35"
    Code    int32   // return code, ONLY meaningful if HasCode is true
    HasCode bool     // did the platform actually supply a return code?
    Reason  uint32
    Unknown string  // "U1" / "U2" / "U3a" / "U3b" — which unknown blocks this, if any
    Err     error   // a wrapped sentinel, for errors.Is
}
```

Two design choices here are worth explaining, because both come directly from evidence
rather than convention:

**Why `HasCode` exists at all.** It would be natural to assume a mainframe service
always returns some kind of status code, the way almost every OS API does. That
assumption turned out to be *wrong* for this exact service, on the one system this
project can actually test against: reading the primary IBM manual directly
(`GC28-0683-2`, p.210) established that a single-line WTO on MVS 3.8j issues **no
return code whatsoever** — register R1 comes back holding a message ID number instead,
and the documented 00–14 return-code table only applies to a different, multi-line form
of WTO (MLWTO) this project doesn't use. Whether z/OS itself differs was an open
question for a long time (research brief 003's Q4); it no longer is — see the note in
§7 below. An error type that assumed a return code always exists would have quietly
baked in an assumption the project's own evidence contradicted. `HasCode` makes "the
service gave us nothing to read" a real, representable state — not a bug, not a zero
that looks like success.

**Why there's a named `Unknown` field distinguishing two kinds of failure.** A caller
of this library needs to be able to tell "this is broken" apart from "this is honestly
not implemented yet because nobody has verified it." Two sentinel errors carry that
distinction:

- `ErrUnsupportedPlatform` — *wrong machine.* You're not running on z/OS/s390x at all.
- `ErrLayoutUnverified` — *right machine, but this project refuses to guess.* The code
  path is complete; what's missing is a citable primary source for a byte layout.

This mirrors, deliberately, the lab modules' `UNDEF`-stub convention
(a standing project rule) — **fail loudly and specifically, never quietly** — but
through a different mechanism, because the audience is different. An `UNDEF` stub is
correct for a *teaching exercise*, where making the build fail is the whole lesson.
It's wrong for a *library*, which has to compile and be testable on an ordinary
developer's laptop so its finished parts (like `codepage`, or `EncodeWPL`'s pure-Go
logic) are actually usable. Same principle, different implementation, because the
consumer changed. See
[ADR 0003](../decisions/0003-production-bridge-module-architecture.md) §4 for the
formal statement of this rule — "no exported function whose behavior depends on an
unretired unknown may return anything other than a typed error naming that unknown,
on every platform, until the corresponding rung passes."

---

## 7. `subsys` and `dataset` — declared, deliberately not designed

```go
// subsys — subsystem requests via IEFSSREQ
func Do(req Request) (Response, error)

// dataset — record-oriented I/O
type RecordReader interface { ReadRecord() ([]byte, error); ... }
type RecordWriter interface { WriteRecord(rec []byte) error; ... }
```

These two packages exist because the owner's decision on module scope
([ADR 0003](../decisions/0003-production-bridge-module-architecture.md), lifting an
earlier scope restriction) chose the full "multi-service enterprise surface" rather
than a WTO-only module — `console`, `dataset`, and `subsys` are the three service
families that implies. But **declaring the family and designing its operations are two
different commitments**, and only the first one has actually been made. Both packages
fix only what's structurally certain (a subsystem request has a function code and a
parameter block and returns a code; a data set is read and written in whole records,
never as an undifferentiated byte stream, because z/OS record formats — fixed, blocked,
variable — don't map onto `io.Reader` without losing information) and defer everything
else. `subsys` in particular is deliberately not designed ahead of time because it's the
exact surface `ibmruntimes/go-recordio` already implements — this project's own
architectural blueprint — and the right next step is reading that code closely (Phase 2
of the roadmap, not started as of this writing) rather than inventing an API against a
guess about a codebase that's sitting there, readable, right now.

---

## 8. Quick reference: what actually works today

| Package | Status | Depends on |
|---|---|---|
| `codepage` | ✅ Working, tested, both architectures | Nothing unretired |
| `console.EncodeWPL` | ✅ Working for the minimal (unrouted) form | U2 — **retired** |
| `console.WTO` (off z/OS) | Returns `ErrUnsupportedPlatform` | — |
| `console.WTO` (on z/OS, once buildable) | Returns `ErrUnsupportedPlatform` naming U3a | U3 — open |
| `console.WithRoute` / `WithDescriptor` | Returns `ErrLayoutUnverified` | Brief 003 Q3 — open |
| `internal/svc.Call35` | Declared, no assembly body yet | U3 (writing it before storage/AMODE is settled risks a live-system fault) |
| `internal/storage.Malloc31` | Returns `ErrUnsupportedPlatform` | U3a — open |
| `subsys`, `dataset` | Interfaces only | Not yet scoped by the owner |

Next: **[`evidence-ladder.md`](evidence-ladder.md)** — exactly how U1 and U2 got
retired, rung by rung, with what was actually run and what it actually proved.

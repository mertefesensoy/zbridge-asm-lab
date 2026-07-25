# ADR 0003 · Production bridge module: layout, API surface, and the evidence-driven seam

- Status: **Accepted**
- Date: 2026-07-25
- Decided by: **Mert Efe Şensoy (owner)**, 2026-07-25 — chose "unblock fully:
  multi-service enterprise surface", lifting goal-prompt §5 autonomy boundary 1
- Builds on: `docs/codex-handover.md` §3 (the WTO call-path decomposition) and §4
  (conventions); ADR 0001 §7 (provenance); ADR 0002 (z/OS guest access);
  `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` (U1 retired, and findings F1/F3)

---

## 1. Context

Autonomy boundary 1 has stood since 2026-07-05: *"the owner will provide the scope
for the production Go↔IBM Z bridge module. Until then, do not scaffold it."* The
owner lifted it on 2026-07-25 and chose the widest of the three offered scopes: a
multi-service enterprise surface, not a WTO-only module.

That is the right call for the thesis and the wrong call to execute naively. A
multi-service API designed before any rung has run would be an API designed against
assumptions. This ADR resolves that tension with a single rule, stated in §4.

## 2. Decision: module identity

| Property | Value | Why |
|---|---|---|
| Module path | `github.com/mertefesensoy/zbridge` | Separate from the lab. The lab is a notebook; this is a library |
| Location | `zbridge/` in this repository, own `go.mod` | Consistent with the per-exercise convention (handover §4.1). Splitting to its own repository is a publication decision, and publication is autonomy boundary 3 |
| Licence | BSD-3-Clause | Matches `ibmruntimes/go-recordio`, from which `codepage`'s tables derive. Attribution obligations travel with the code (handover §4.7) |
| cgo | **Never** | The no-cgo constraint is the thesis |
| Minimum Go | 1.26 | Matches the lab modules |

## 3. Decision: package layout

```
zbridge/
  zbridge.go            façade: Platform(), Supported(), version
  errors.go             typed error model (§5)
  doc.go

  codepage/             EBCDIC ⇄ ASCII. Public, standalone-useful.
                        BSD-3-Clause-derived tables + LICENSES.md.
                        Promoted from the ebcdic exercise, TR path included.

  console/              WTO, WTOR. The thesis endgame.
      wpl.go            parameter-list encoder — pure Go, testable anywhere
      console_zos.go    the SVC 35 call path
      console_stub.go   every other platform

  dataset/              record I/O (QSAM/BSAM shape). Declared, minimally scaffolded.
  subsys/               IEFSSREQ subsystem requests, the go-recordio surface.
                        Declared, minimally scaffolded.

  internal/svc/         raw supervisor-call primitives (assembly)
  internal/linkage/     MVS save-area / register conventions / AMODE helpers
  internal/storage/     Malloc31, below-the-bar allocation (U3a, LE touchpoint)
```

**`console`, `dataset` and `subsys` are the three service families.** Declaring all
three now is what "enterprise surface" means; implementing all three now is not, and
§4 governs which parts get bodies.

## 4. The governing rule: no API commits ahead of its evidence

**Every exported function whose behaviour depends on an unretired unknown returns a
typed error naming the unknown, and does so on every platform including z/OS, until
the corresponding rung passes.**

This is the mechanism that lets a wide API be declared honestly.

| Surface | Depends on | Status today | Ships as |
|---|---|---|---|
| `codepage.AtoE` / `EtoA` | U1 | **Retired** (E-L evidence) | Working code, both architectures |
| `console.EncodeWPL` | U2 — the WPL byte layout | **Open.** No primary citation on either system; research brief 003 outstanding | Complete API, returns `ErrLayoutUnverified` |
| `console.WTO` | U2 + U3a + U3b | Open | Complete API, returns `ErrLayoutUnverified` before it can reach the platform stub |
| `internal/storage.Malloc31` | U3a | Open | Declared, returns `ErrUnsupportedPlatform` |
| `dataset.*`, `subsys.*` | not yet scoped | Open | Interfaces only |

The rule's value is that the failure is *typed and specific*. A caller gets
`ErrLayoutUnverified` with a message pointing at brief 003 — not a panic, not a wrong
answer, and not a silently-plausible parameter list.

This is the production-module analogue of the lab's `UNDEF` convention (CLAUDE.md
hard rule 1), and the difference is deliberate: `UNDEF` is correct for a teaching
exercise where a build failure is the lesson, and wrong for a library, which must
compile and be importable on a developer's laptop. **Same principle — fail loudly,
never silently — different mechanism because the consumer is different.**

## 5. Decision: the error model

```go
type Error struct {
    Op       string   // "console.WTO"
    Service  string   // "SVC 35"
    Code     int32    // return code, if the platform supplied one
    HasCode  bool     // whether Code means anything at all
    Reason   uint32
    Unknown  string   // "U2" — which unknown blocks this, if any
    Err      error    // wrapped sentinel, for errors.Is
}
```

Sentinels: `ErrUnsupportedPlatform`, `ErrLayoutUnverified`, `ErrNotAuthorized`,
`ErrMessageTooLong`, `ErrServiceFailed`.

**`HasCode` is not defensive padding; it is a finding made structural.** H001's C4
finding, read directly from GC28-0683-2 p.210, is that a single-line WTO on the
ancestor system **issues no return code at all** — R1 comes back holding the message
identification number, and the documented 00–14 table is MLWTO-only. Whether z/OS
behaves the same way is exactly what is in doubt (it puts ADR 0001 §6's "E3 retires
Phase 3b step 6" claim in question, and brief 003 Q4 is the resolving question).

An error type that assumes a return code exists would encode an assumption this
project has documentary evidence against. `HasCode` makes "the service returned
nothing to map" a representable state instead of a bug.

## 6. Decision: the E3 seam

The parameter-list encoder is **pure Go, exported, and testable on any host**:

```go
func EncodeWPL(msg string, opts ...Option) ([]byte, error)
```

This is not an implementation detail that happens to be separable. It is the shape
rung E3 requires. ADR 0001 §6 defines E3 as: a Go program emits the parameter list it
intends to build, that byte sequence is embedded in an MVS assembler program as
`DC X'...'` constants, and a real `SVC 35` is handed it. **The module is designed so
that E3's artifact is a first-class exported capability rather than a one-off script**
— `EncodeWPL` plus a hex formatter is the whole of the Go side of that rung.

Consequence: the byte layout lives in exactly one file (`console/wpl.go`), is
constructed by code that runs on a Windows laptop, and is verified by an emulator
that never runs Go. Splitting construction from invocation is what makes the
unknown testable before the platform exists.

## 7. Decision: build-tag strategy

```
console_zos.go       //go:build zos && s390x
console_stub.go      //go:build !(zos && s390x)
```

The real path is gated on `zos && s390x` together, never on `s390x` alone —
`linux/s390x` is a real and useful target (it is where U1 was retired) and must not
be mistaken for z/OS.

**The `zos` build tag does not work with upstream Go today**, and that is recorded
rather than assumed: `go tool dist list` (go1.26.3) offers no `zos/s390x`, and
`internal/platform` does not list it, though `"zos"` is in `internal/syslist` so the
filename constraints do parse. IBM's Go fork is a hard dependency (evidence E-L, F3).
Until it is in hand, the `zos && s390x` files are compiled by nothing, and CI
compiles the stub path only. That is a real limitation and it is not hidden.

## 8. Decision: `SVC 35` will be hand-encoded

Go's s390x assembler has no `SVC` mnemonic — 729 mnemonics in
`cmd/internal/obj/s390x/anames.go`, and `SVC`, `TR`, `TRT` and `EX` are all absent
(evidence E-L, F1). `SYSCALL` exists and assembles to `SVC 0`, which is Linux's trap
and not what WTO needs.

`SVC 35` is therefore emitted as literal bytes: `BYTE $0x0A; BYTE $0x23`. The
technique is attested inside the Go distribution — `x/sys/unix/asm_zos_s390x.s`
encodes `SVC 08` as `BYTE $0x0A; BYTE $0x08` — and it was already exercised and
disassembly-verified in this project today for `TR` (evidence E-L). The same
verification is mandatory for `SVC 35` before any claim rests on it.

## 9. What this decision does not claim

- **It does not claim the WPL layout.** Nothing in this ADR states a byte offset.
  `console/wpl.go` will contain the layout when brief 003 or rung E1's assembler
  listing supplies a primary citation, and not before.
- **It does not claim `dataset` or `subsys` scope.** Both are declared as families
  with interfaces. Their operations need their own scoping conversation, and
  go-recordio is the pattern source for `subsys` specifically.
- **It does not claim the module works on z/OS.** Nothing has run on z/OS. Every
  z/OS-side path is unexecuted code.
- **It does not commit to publication.** Releasing this module publicly remains
  autonomy boundary 3 and is the owner's call.
- **It does not supersede the lab.** The six exercises stay as they are. The library
  promotes `codepage` from `ebcdic`; the other five exercises remain teaching
  artifacts and keep their `UNDEF` stubs where they have them.

## 10. What would reopen or reverse this decision

1. **Brief 003 returns a WPL layout that does not fit a `[]byte` encoder** — for
   example if the parameter list requires live pointers to storage the caller must
   allocate below the line. Then `EncodeWPL`'s signature changes and §6's seam needs
   rework. This is the most likely reversal and it is why §4 exists.
2. **IBM's Go fork imposes a different assembly or ABI convention** than upstream Go
   for s390x. Then §7 and §8 are re-derived against the fork, not against upstream.
3. **The owner scopes `dataset` or `subsys` concretely.** Those sections become their
   own ADRs rather than amendments to this one.
4. **A single-line WTO on z/OS is confirmed to issue a real return code.** Then
   `HasCode` in §5 may be simplified — but only on evidence, and the H001 C4 finding
   currently points the other way.

## 11. Links

- `docs/codex-handover.md` §3 — the call-path decomposition this layout implements
- `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` — F1 (no `SVC`/`TR` mnemonics),
  F3 (no `zos/s390x` in upstream Go)
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — the C4 finding behind §5
- `docs/research-briefs/003-wto-wpl-layout-source-and-return-code-contract.md` — the
  outstanding item that unblocks §6
- `docs/decisions/0002-zos-under-hercules-permitted-by-ibm-backing.md` — how the
  z/OS-side paths might first be executed

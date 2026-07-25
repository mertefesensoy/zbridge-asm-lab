# Evidence

Captured outputs from ladder rungs. Raw and dated. This directory is what makes the
project's claims checkable by someone who was not present.

**Filename:** `<rung>-<subject>-<YYYY-MM-DD>.md`
(e.g. `E0-tk5-boot-2026-07-28.md`, `T1-add-zos-2026-09-14.md`)

## `DOC` — documentary verification (provisional, added 2026-07-25)

One file here executes nothing: `DOC-001-wto-wpl-primary-source-2026-07-25.md`, the
page-by-page read of GC28-0683-2 that audited the brief 001 return. It is a
verification of what a **primary source says**, not of what a machine did, so it fits
no rung and sets the machine-specific header fields to `n/a`.

It is filed here anyway because a source check that contradicts a research return has
to be independently citable, and because burying it in a hypothesis would have meant
editing pre-registered text. **The `DOC` prefix is a provisional convention extension
awaiting owner review** — see that file's own "Convention note" for the alternative
that was considered and rejected. If the owner prefers documentary checks live
elsewhere, this is the only file affected.

## Provenance header — mandatory

Every evidence file opens with this block, filled in completely. It is required by
ADR 0001 §7 and exists to make it impossible to mistake an emulated result for a
hardware result.

```markdown
---
rung:          E0 | E1 | E2 | E3 | T0 | T1 | T2 | T3
date:          YYYY-MM-DD
machine:       real hardware | QEMU | Hercules
guest_os:      MVS 3.8j (TK5 Update N) | Linux s390x (distro/version) | z/OS vN.N
architecture:  S/370 (24-bit) | z/Architecture (64-bit)
emulator:      name + exact version, or "n/a — real hardware"
host:          OS, CPU, RAM
speaks_to:     U1 | U2 | U3
hypothesis:    H001 | H002 | ... | none
verdict:       PASS | FAIL | INCONCLUSIVE
---
```

`speaks_to` is the field that does the real work. It forces the question *"which
unknown does this actually retire?"* at capture time rather than at write-up time,
when the temptation to over-read a result is strongest.

## Rules

- **Capture raw output.** Full command lines, full listings, exact error text. Trimmed
  output hides the detail that turns out to matter three weeks later.
- **Record failures.** A failed rung is evidence. Failures are not deleted or
  quietly re-run until green; the failure and its diagnosis both stay.
- **No emulated result may be described as a z/OS result** in any document, ever
  (ADR 0001 §7). The provenance header makes the distinction mechanical.
- **Pin versions.** Emulator version, distro version, Go version, TK5 update level.
  A result that cannot be reproduced against a known configuration is an anecdote.
- **Negative controls belong here too.** Where a hypothesis specifies one (H001
  threat 3), its output is captured like any other rung.

## The ladders

**E-ladder — off-mainframe** (ADR 0001 §6):

| Rung | Gate |
|---|---|
| E0 | TK5 IPLs; console readable; job submits and output returns |
| E1 | `WTO` macro puts a message on the MVS operator console; assembler listing captured |
| E2 | Hand-built parameter list + raw `SVC 35` reaches the console — no macro |
| E3 | Go-produced byte sequence fired at a real `SVC 35` reaches the console |

**T-ladder — real hardware** (`docs/mainframe-baseline-strategy.md` §4):

| Rung | Gate |
|---|---|
| T0 | Pure-Go binary runs on the target |
| T1 | `add` passes on the target |
| T2 | All five exercises pass with real `_s390x.s` bodies |
| T3 | WTO scaffold: message on the z/OS operator console |

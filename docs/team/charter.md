# Team charter · zbridge-asm-lab

**Shape fixed by [ADR 0005](../decisions/0005-team-scale-up-and-academic-year-recalibration.md) §3 and §4.**
This document is operational and may change without an ADR. The *shape* — five workstream
seats, and doctrine enforced at the merge gate — may not.

- **Team:** 5 · lead + 4 contributors, all proficient computer-engineering students
- **Lead:** Mert Efe Şensoy (TED University · IBM Z Student Ambassador · IBM Champion 2026)
- **Mentor:** Jürgen Holtz (IBM) — advisory, one point of contact, **not** in the review loop
- **No onboarding phase.** Contributors join at working level. §9 is the read-first list.

---

## 1. What this project is, in one paragraph

We are building `WTO(message string) error` — a single Go function, written in **pure Go
(Plan 9) assembly for s390x**, that writes a message to the z/OS operator console by issuing
`SVC 35` directly. **No cgo. No C. No Language Environment dependency** beyond one `Malloc31`
touchpoint. `ibmruntimes/go-recordio` proves the architectural pattern works but covers only
IEFSSREQ; **no public Go-assembly implementation of an operator-facing z/OS service exists**,
and that gap is the thesis. Two of the project's three unknowns are already retired without a
mainframe. The third needs entitled z/OS access that does not exist yet.

**The quality of the explanation is a deliverable, not a byproduct.** A correct patch with a
thin rationale is not done.

---

## 2. The five seats

Seats are **standing workstreams**, not a shared queue. The reason is specific: this project's
doctrine is heavy — pre-registered hypotheses, provenance headers, ADR discipline, a hard ban
on uncited claims — and that is learned by owning an area, not by picking up whatever is next.

Everyone reviews everyone. Owning a seat means being accountable for it, not being alone in it.

### W1 · Service semantics & assembly

**Owns:** Phase 2 annotation · the WPL · `SVC 35` encoding · later WTOR.
**Needs to be comfortable with:** reading assembly listings, byte-level data layout,
big-endian thinking. Mainframe experience is *not* assumed — nobody here started with it.

**First ticket:** `PHASE-2` — annotate `ibmruntimes/go-recordio` **`v2/utils/utils.s`**
(154 lines) across its nine routines: `IefssreqX`, `Bpxcall`, `Svc8`, `Svc9`, `Call24`,
`Call31`, `Call64`, `Deref`, `Pc31`.
⚠️ **Read [errata E4](../roadmap-errata.md) first** — the roadmap points at `utils/utils.s`,
which does not exist.

### W2 · Go module & toolchain

**Owns:** the `zbridge/` module · IBM's Go fork · the `GOOS=zos` gate · platform build splits.
**Needs to be comfortable with:** Go build constraints, cross-compilation, the stub/real
platform-split pattern, Go's `internal/` discipline.

**First ticket:** `PHASE-3A` prep — write the `_zos.go` / `_zos_s390x.s` files against the
constraint names that already parse (`"zos"` is in `internal/syslist`, so the constraints are
valid even though the target is not), keeping both `windows/amd64` and `linux/s390x` green.

### W3 · Emulation & lab operations

**Owns:** the TK5/Hercules harness · QEMU · reproducibility · CI.
**Needs to be comfortable with:** shell scripting, WSL2, process lifecycle, being pedantic
about what counts as proof that something stopped cleanly.

**First ticket:** `OPS` — productise `docs/runbooks/mvsjob.sh` and stand up CI (there is no
`.github/` yet) running `go vet`, `go test`, and the `GOOS=linux GOARCH=s390x` cross-build on
every PR.
⚠️ Read the operational-lessons sections of both evidence files before touching the emulator.
Three unclean stops are recorded there, with the rules that came out of them.

### W4 · Evidence, hypotheses & publication

**Owns:** provenance audits · ADRs · hypotheses · the thesis-facing writing.
**Needs to be comfortable with:** technical writing, and the discipline to say "not verified"
in a document that would read better without it.

**First ticket:** `EVIDENCE` — audit everything currently committed against ADR 0001 §7, and
land the untracked documents sitting in the working tree.

### W5 · Lead & integration — *the owner*

**Owns:** the mentor interface · the roadmap · ADR approval · final review · the access and
entitlement pursuit.

**Reserved to this seat and not delegable** (`docs/goal-prompt.md` §5):

- Anything **published publicly** — releases, posts, module publication.
- **Any spend of emulator or mainframe session time** not already authorised.
- **Approval of any ADR**, and any change that would reverse an accepted one.
- **All mentor-facing communication.** Not ceremony — it is what keeps the citation
  discipline attached to everything IBM sees.

---

## 3. How work enters: the ticket

The tracker itself — GitHub Issues, Jira, Linear, whatever the lead picks — is **not fixed
here**. What is fixed is the ticket's shape and the gate it has to pass.

| Type | For | Must state |
|---|---|---|
| `PHASE-n` | Roadmap phase work | Which deliverable, and its gate |
| `RUNG` | An E- or T-ladder rung | Which unknown it retires, and what evidence it will produce |
| `ADR` | A decision needing recording | The fork, and the options rejected |
| `BRIEF` | External research | Numbered questions; goes to Gemini, not to a crawl |
| `EVIDENCE` | Capturing a result | Provenance header fields, filled |
| `FIX` | Bug, build break, doc error | What broke and how it is verified fixed |
| `OPS` | Harness, CI, tooling | What it automates and what it replaces |

**Every ticket answers one question before it is accepted:**

> **Which of U1, U2, U3 does this retire — or does it retire none, and why is it still worth
> doing?**

"None, and here is why" is a perfectly good answer. Being unable to answer at all is the
signal that the ticket is not ready.

---

## 4. Branch and PR conventions

```
<type>/<ticket-id>-<short-slug>
```

`phase2/T14-annotate-call31` · `rung/T22-e4-dry-run` · `fix/T31-jobcard-length`

- **Branch from `main`. Never commit to `main` directly.**
- One ticket, one PR. If a PR needs two summaries, it is two PRs.
- Rebase, don't merge-commit, before review.
- **The roadmap PDF is never edited.** Contradictions become an ADR plus an errata entry.

---

## 5. The pull-request template

Lives at `.github/pull_request_template.md`. Reproduced here so the reasoning is visible:

```markdown
## What and why
<!-- The change, and the problem it solves. "Why" is not optional on this project. -->

## Which unknown does this retire?
- [ ] U1 — Go assembly correctness / ABI on big-endian 64-bit
- [ ] U2 — WTO parameter list, SVC 35 acceptance
- [ ] U3 — GOOS=zos, Malloc31, AMODE 31↔64, USS
- [ ] None — and here is why it is still worth doing:

## Claims check
Every factual claim about IBM, z/OS, MVS or z/Architecture behaviour in this PR is:
- [ ] **cited** — primary source with form number, or Go source path
- [ ] **evidenced** — a file in docs/evidence/
- [ ] **registered** — an open assumption inside a hypothesis
- [ ] no such claims in this PR

There is no fourth category. Hedged guessing is the specific thing this box exists to stop.

## Verification — run, not reasoned about
<!-- Paste actual output. "Should work" fails review. -->
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `GOOS=linux GOARCH=s390x go build ./...`  (n/a for `add/`)
- [ ] not applicable, because:

## If this touches docs/evidence/
- [ ] Provenance header complete: machine, guest OS, architecture, emulator version, `speaks_to`
- [ ] No emulated result is presented as a z/OS result
- [ ] No performance number taken under an emulator

## Documentation
- [ ] `docs/implementations/YYYY-MM-DD-<slug>.md` written, or genuinely not warranted
- [ ] `memory/MEMORY.md` still accurate after this change

## Anything deliberately left undone
<!-- Scope is delivered, not narrowed. If something was left out, say so here. -->
```

---

## 6. Review gates — what has to pass before merge

| The PR touches… | Automated | Human |
|---|---|---|
| any `.s` file | `go vet` hook · `go-asm-reviewer` | 1 reviewer |
| any `_s390x.s` file | above **+ s390x cross-build hook** | 1 reviewer |
| `docs/evidence/**` | `evidence-provenance-auditor` | W4 |
| `docs/decisions/**` | — | **W5 — lead approval, mandatory** |
| `ebcdic/tables.go`, any `LICENSES.md` | confirmation hook | **W5** — BSD-3-Clause attribution must survive |
| `console/wpl.go` `LayoutVerified` | — | **W5** — it is a claim about reality, not a flag |
| an `UNDEF`-bearing stub | confirmation hook | **W5** |
| `zbridge/**` public API | CI | 1 reviewer + W2 |
| anything else | CI | 1 reviewer |

The hooks in `.claude/hooks/` (`asm-gate.ps1`, `convention-guard.ps1`) and the subagents in
`.claude/agents/` already exist and already do this work locally. **CI does not exist yet** —
standing it up is W3's first ticket, and until then these gates are honoured by hand.

### Two automatic rejections

1. **Reaching green by weakening.** Replacing an `UNDEF` stub with a no-op, loosening a test,
   dropping a module from a table so the table looks complete, or optimising the pedagogical
   byte loops to win a benchmark. A failing rung captured honestly is worth more than a passing
   rung that was made to pass.
2. **An emulated result described as a z/OS result**, or any performance number taken under an
   emulator. QEMU and Hercules implement `TR` as a software loop; timing it measures the
   emulator. The project's contribution is credibility about a gap nobody has documented, and
   one overstated claim costs more than the entire emulation programme returns.

---

## 7. Cadence

| When | What |
|---|---|
| **Weekly** | Async written standup per seat: what moved, what is blocked, which unknown |
| **Fortnightly** | Team sync — 45 min. Gate review against `docs/roadmap-2026-27.md` |
| **Monthly** | Written checkpoint to the mentor from W5, **whether or not a meeting is scheduled** |
| **Per phase gate** | The gate is demonstrated, not asserted. Evidence or it did not happen |

The monthly written checkpoint is a rule with a reason: the 2026-07-27 meeting slipped, two
prepared briefings went undelivered, and the project's open questions sat unasked for three
days. Written checkpoints do not depend on calendars.

---

## 8. Definition of done

A ticket is done when — from `docs/goal-prompt.md` §6:

- [ ] Every rung run has evidence in `docs/evidence/` with a complete provenance header
- [ ] Every hypothesis whose evidence now exists has an accurate status
- [ ] Any decision made is an ADR, with its "does not claim" and "would reopen" sections
- [ ] `docs/implementations/YYYY-MM-DD-<slug>.md` exists for the change
- [ ] `memory/MEMORY.md` reflects current state and the next unblocked action
- [ ] **What was *not* done, and why, is stated explicitly**

---

## 9. Read-first, in this order

Roughly two hours. It is the whole context.

1. **`CLAUDE.md`** — the standing rules, and the ones that are easy to break helpfully.
2. **`docs/goal-prompt.md`** — the doctrine. §4 is the part that makes this project different.
3. **`zbridge-asm-roadmap.pdf`** (8 pages) — the mandate. **Then `docs/roadmap-errata.md`**,
   because four of its statements are superseded and one produces a broken parameter list if
   implemented as written.
4. **`docs/roadmap-2026-27.md`** — where we are and what is left.
5. **`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`** — the project's central result.
6. **ADRs 0001, 0003, 0004, 0005.** `0002` is **withdrawn** — read its §10 before citing
   anything in it.

### The five things a competent engineer will otherwise break, helpfully

1. **`UNDEF` stubs are `UNDEF` on purpose.** They must fail the build if targeted. Never
   replace one with a no-op to make something pass. (The production module uses typed errors
   instead, for the same reason by a different mechanism — a library must be importable on a
   laptop.)
2. **Benchmarks are honest.** The naive byte-at-a-time assembly *loses* to pure Go on amd64
   and the READMEs say so. The byte loops are the pedagogy.
3. **Bare `SP` on s390x is only the pseudo-register.** The hardware stack pointer is `R15`.
   Conflating them is the most common Go-assembly bug and `regs/` exists to document it.
   When a signature changes, re-derive offsets: **strings are 2 words, slices are 3 words.**
   Run `go vet` — it checks the `$frame-args` contract.
4. **There is no `TR`, `TRT`, `EX` or `SVC` mnemonic** in Go's s390x assembler — 729 names in
   `cmd/internal/obj/s390x/anames.go`. They are hand-encoded as `BYTE` directives, which is
   attested practice inside the Go distribution itself and inside go-recordio.
5. **`add/` does not cross-compile for s390x and is not supposed to.** A repo-wide s390x build
   fails on it. That is expected, not a regression.

---

## 10. What is not decided here

- **Who fills which seat.** ADR 0005 fixes the shape; the lead assigns the names.
- **The tracker product.** Lead's choice.
- **Whether the team's size holds.** At three seats, W3 folds into W2 and W4 into W5
  ([ADR 0005 §9.1](../decisions/0005-team-scale-up-and-academic-year-recalibration.md)).
- **Anything about the mentor's involvement.** ADR 0005 §7: he is **informed** of the scale-up
  and **nothing is asked of him** about it. His interface is unchanged.

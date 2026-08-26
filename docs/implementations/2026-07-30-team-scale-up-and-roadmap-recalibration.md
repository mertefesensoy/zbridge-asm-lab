# Team scale-up to five, roadmap recalibration to AY 2026–27, and the replacement mentor email

**Date:** 2026-07-30
**Author:** Mert Efe Şensoy
**Status:** shipped — except the mentor email, which is **drafted and unsent** by design

---

## 1. Problem / Motivation

Four things came together and none of them could be addressed alone.

**The project acquired four more people.** It was written, planned, and executed as a solo
thesis project. Every planning artifact in the repository assumes one contributor: the
roadmap's time budget (p.7) says "5-8 hours per week", the doctrine's autonomy boundaries
address "the owner", and the review discipline lived in one person's head plus local notes.
None of that survives four more contributors unchanged.

**The plan's dates were spent.** The roadmap's phase table (p.3) runs on 2026 calendar
windows. Phase 2's window — mid-July to early August — was closing with the phase not
started, and the roadmap's own risk register did not anticipate the dependency that now
gates everything downstream (IBM's Go fork).

**The mentor meeting of 2026-07-27 did not happen.** Two briefings had been prepared for it.
Neither was delivered, so every question in them — including the two that block Phase 3 — had
been sitting unasked. There was no artifact that could be *sent* rather than *presented*.

**The repository's own state documents had gone stale in ways that would mislead.** The standing rules
still said *"Rung E0 is still unrun and is the owner's chosen next milestone"* — four days
after E0 through E3 all passed. `memory/MEMORY.md`'s active plan still listed rung E0 as action
1 and research brief 003 as unrun, and its boundaries section still said
`LayoutVerified` was `false` when it had been flipped to `true` on evidence. A fresh contributor
reading either file would have started on work that was already finished, and might have
"fixed" a verified constant back to its unverified value.

## 2. What Changed

| File | Change |
|---|---|
| `docs/decisions/0005-team-scale-up-and-academic-year-recalibration.md` | **New ADR.** Records the scale-up to five seats, the re-anchoring to AY 2026–27, doctrine moving to the merge gate, and the explicit decision *not* to re-split U3. |
| `docs/roadmap-2026-27.md` | **New.** Part I audits where the project actually is against the roadmap's own phases, with an evidence link per cell; Part II is the phase-by-phase forward plan; Part III the three-block schedule with a named contingency; Part IV the risks. |
| `docs/team/charter.md` | **New.** Five workstream seats with entry state and first ticket, the ticket taxonomy, branch/PR conventions, the review-gate matrix, cadence, definition of done, and the read-first list. |
| `.github/pull_request_template.md` | **New.** The merge gate made concrete — the "which unknown does this retire" question, the claims check, run-not-reasoned verification, and the evidence/assembly checklists. |
| `docs/mentor-briefings/2026-07-30-progress-and-hercules.md` | **New.** The replacement for the undelivered 2026-07-27 pair, written as an email. Carries progress, the Hercules account, the three roadmap corrections, the team note, and six questions. **Marked DRAFT / NOT SENT.** |
| `docs/roadmap-errata.md` | Added scope change **S4** (phase windows and solo time budget superseded) and a fourth entry under "structures the roadmap does not contain" (the team structure and PR gate). |
| project standing rules | "Current state" rewritten — it claimed rung E0 was unrun. Three rows added to the "where things are" table. Errata count corrected from three to four corrections plus four scope changes. |
| `memory/MEMORY.md` | "Active plan" rewritten (it listed finished work as next actions); ADR 0005 added to decisions of record; the `LayoutVerified = false` boundary corrected; three session-log entries added for 07-26, 07-27 and 07-30. |

**No code changed.** Not one line of Go or assembly. This is a planning and documentation
change, and the test runs in §6 exist to establish the baseline the plan describes, not to
verify an edit.

## 3. Implementation Approach

### The load-bearing structure: separate the decision from the plan from the procedure

Three documents, because they change at three different rates and have three different
approval requirements:

| Document | Rate of change | Who may change it |
|---|---|---|
| **ADR 0005** — the decision | Rarely; superseded, never edited | Owner approval (autonomy boundary 4) |
| **`roadmap-2026-27.md`** — the plan | Whenever an entitlement or gate moves | Lead, on evidence |
| **`team/charter.md`** — the procedure | Freely, as the team learns | Anyone, by PR |

Collapsing these into one "project plan" document is the obvious alternative and it is why
plans rot: the parts that should be stable get edited alongside the parts that should not, and
after a semester nobody can say what was decided versus what drifted.

### The audit is evidence-linked per cell, not summarised

Part I of the roadmap gives every status cell a pointer to the file that justifies it. Where no
evidence exists, the cell says "not started" or "blocked" rather than estimating a percentage.
This is `docs/goal-prompt.md` §4.1 applied to project status: a status claim is a claim.

The contract this establishes for future readers: **any cell in that table can be checked
without trusting the table.**

### Doctrine moves from memory to the merge gate

The project's rules — no emulated result presented as a z/OS result, no performance number from
an emulator, no reaching green by weakening, no claim without citation — were enforced by one
person knowing them. The charter maps each rule onto the place it is now checked: an existing
hook, an existing subagent, a PR-template checkbox, or a named seat's mandatory approval.

Nothing about the rules changed. Only where they live.

### The email is written to be sent, and is not sent

The draft is a complete email with a subject line, addressed and signed, so that sending it is
a copy-paste and not a rewrite. It carries an explicit non-email preamble marking it DRAFT, the
context it assumes (the meeting slipped, both prior briefings undelivered), and three things to
check before sending. Sending is autonomy boundary 3.

## 4. Mathematical / Statistical Details

Not applicable — this change is structural. No formula, statistical test, or numeric algorithm
is involved.

The one quantitative claim it makes is a count of passing tests, and §6 shows how it was
obtained.

## 5. Design Decisions

### 5.1 Re-split U3 into toolchain and access? — **Rejected**

The toolchain dependency (IBM's Go fork) and the access dependency (an entitled z/OS system)
are visibly different things: different entitlements, different channels, different failure
modes. Splitting them into U3a/U3b would make the plan easier to describe.

**Rejected because it amends ADR 0001 §3**, which is the structure every evidence file's
`speaks_to:` header points at, and because this project has already spent one ADR withdrawal on
moving faster than the evidence (ADR 0002). ADR 0005 §6 instead records the *ordering fact* —
you cannot run a binary on a system you have no compiler for — which requires no amendment and
is sufficient to sequence the plan. Whether the split should become formal is listed as an
explicit reopen condition.

### 5.2 Seats by workstream, or a shared ticket queue? — **Workstreams**

A shared queue load-balances better and is the default for a five-person team.

**Rejected because this project's doctrine is unusually heavy** and is learned by owning an
area. A contributor who picks up whatever is next will write a plausible, fluent, unverified
sentence into an evidence file, because they have not internalised why the provenance header
exists. Owning W3 for a semester means having personally hit the three unclean-stop failures
recorded in the E0 evidence. That is not transferable by reading a checklist.

The charter mitigates the load-balancing loss by making everyone a reviewer.

### 5.3 An onboarding phase? — **Rejected, on owner ruling**

The obvious plan for four newcomers to a mainframe project is a ramp: work the six existing
exercises, then take real tickets.

The owner ruled against it — the contributors are proficient computer-engineering students and
join at working level. The charter honours that by making the exercises *reference material*
and giving each seat a real first ticket, while keeping one concession to the actual risk: §9's
"five things a competent engineer will otherwise break, helpfully". That list exists because
the failure mode here is not incompetence, it is competent instincts applied to a codebase
where `UNDEF` stubs and losing benchmarks are deliberate.

### 5.4 Re-date the roadmap, or replace it? — **Neither; supersede by pointer**

The roadmap PDF is never edited — it is the mentor-facing mandate and the record of what was
believed in June, and rewriting it destroys the evidence trail the project's credibility rests
on. Regenerating it with corrections inline was rejected for ADR 0004 §6 and the same reasoning
holds here.

So `roadmap-2026-27.md` supersedes only the *timeline*, says so in its first three lines, and
errata S4 records the supersession where a reader of the PDF will find it. ADR 0005 §2 states
first and most emphatically what the recalibration does **not** touch: the endgame, the phase
definitions, and every deliverable.

### 5.5 Should the email ask Jürgen about mentoring a larger team? — **No, on owner ruling**

Scaling from one mentee to a five-person team is a material change to what he agreed to, and
the honest move is arguably to name it and ask.

The owner ruled: **inform, do not ask.** The email states the change in three sentences,
explicitly says nothing changes on his side, and makes no request about it. ADR 0005 §7 records
this so a future contributor does not "helpfully" add the ask back in.

### 5.6 Should Block C have a stated Plan B? — **Yes, and it is the most useful part of the schedule**

Writing a contingency for "no z/OS by February 2027" risks reading as defeatism in a
mentor-facing document.

**Included anyway**, because the alternative is discovering in February that the thesis has no
fallback. The five Plan B deliverables are all real, all independent of z/OS, and one of them —
the E-ladder as transferable methodology — is arguably a better contribution than the WTO
function itself. The project already treats negative results as publishable (goal-prompt §4.3);
this applies that stance to its own schedule.

## 6. Verification

### What was run, 2026-07-30, `windows/amd64`, Go 1.26.3

```powershell
cd zbridge
go vet ./...          # clean
go test ./...         # 27 pass, 0 fail
```

```powershell
# every module that must cross-build
$env:GOOS='linux'; $env:GOARCH='s390x'
foreach ($m in 'ebcdic','strmanip','regs','bytecmp','syscall-linux','zbridge') {
  cd $m; go build ./...   # clean on all six
}
```

Test counts were taken by counting `--- PASS` lines under `go test -count=1 -v ./...`:
**27 in `zbridge`; 19 across the lab modules** (`add` 1, `ebcdic` 5, `strmanip` 5, `regs` 4,
`bytecmp` 4, `syscall-linux` 0 — the last contributes nothing on Windows because it is
Linux-constrained by design).

### What was *not* run, and is stated rather than glossed

- **The 29-test `linux/s390x` figure was not re-run.** It is quoted from
  `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md` and labelled as such in the roadmap.
- **No emulator was started.** Nothing in this change touches TK5, and starting it would spend
  session time for no gate (autonomy boundary 2).
- **The email was not sent.** Autonomy boundary 3.

### How to check the documents themselves

1. Every status cell in `docs/roadmap-2026-27.md` Part I cites a file. Open three at random and
   confirm the cell matches the file.
2. `docs/roadmap-errata.md` S4 should agree with ADR 0005 §5. If they disagree, the errata is
   wrong by construction — the ADR governs.
3. The standing rules' "Current state" and `memory/MEMORY.md` "Current state" should now agree with each
   other and with the evidence files. Before this change they did not.

## 7. Related Docs

- [ADR 0005](../decisions/0005-team-scale-up-and-academic-year-recalibration.md) — the decision this doc implements
- [`docs/roadmap-2026-27.md`](../roadmap-2026-27.md) — the plan
- [`docs/team/charter.md`](../team/charter.md) — the procedure
- [`docs/mentor-briefings/2026-07-30-progress-and-hercules.md`](../mentor-briefings/2026-07-30-progress-and-hercules.md) — the unsent email
- [`docs/roadmap-errata.md`](../roadmap-errata.md) — S4 records the supersession
- [ADR 0001](../decisions/0001-emulation-strategy-hercules-two-track.md) §3 — the U1/U2/U3 decomposition left intact by §5.1
- [ADR 0004](../decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) — the corrections and the cgo closure the email reports
- [`docs/goal-prompt.md`](../goal-prompt.md) §5 — the autonomy boundaries that kept the email unsent
- [`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`](../evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md) — the result the whole plan is built on

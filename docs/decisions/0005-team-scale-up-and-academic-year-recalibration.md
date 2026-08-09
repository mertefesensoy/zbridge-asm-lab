# ADR 0005 · The project scales from one person to five, and the plan re-anchors to academic year 2026–27

- Status: **Accepted**
- Date: 2026-07-30
- Author: Claude (architecture role)
- Decided by: **Mert Efe Şensoy (owner and team lead)**, 2026-07-30, in session
- Supersedes: the phase-window column of the roadmap's phase table (`zbridge-asm-roadmap.pdf`
  p.3) and the single-contributor time budget in its risk register (p.7)
- Does **not** supersede: any phase *definition*, any deliverable, or the endgame
- Evidence: `docs/evidence/E0-tk5-boot-2026-07-26.md`,
  `docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md`,
  `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`

---

## 1. Context

Three things changed at once, and they interact.

**The project acquired four more people.** It was written, planned, and executed as a
solo thesis project with a mentor. It is now a five-person team — the owner as lead plus
four contributors, all proficient computer-engineering students — with Jürgen Holtz (IBM)
continuing as mentor. The owner's ruling is that **no onboarding phase is inserted**;
contributors join at working level and the existing exercise ladder serves as reference
material rather than as a gate.

**The roadmap's timeline has been overtaken.** The PDF's phase table (p.3) runs on 2026
calendar windows that assumed one person at 5–8 hours per week (p.7). Phase 1 landed about
a week after "late June"; Phase 1b about ten days after "mid July"; Phase 2's window
(mid-July to early August) is closing with the phase not started. The roadmap says of
itself that *"the phase definitions are stable, the timeline is provisional"* (p.8), which
is the licence this ADR uses.

**The mentor meeting of 2026-07-27 did not happen.** Two briefings were prepared for it and
neither was delivered. Every question in them is still open, including the two that gate
Phase 3 — z/OS access and IBM's Go fork.

Meanwhile the technical position improved substantially and asymmetrically: **U1 and U2 are
both retired** and **U3 is untouched**. The plan has to be re-cut around that shape, because
it means the remaining work divides cleanly into a large body that needs no entitlement and
a smaller body that cannot start without one.

## 2. Decision: the endgame, the phase definitions, and the deliverables are unchanged

Stated first because it is the most important thing this ADR does *not* do.

`WTO(message string) error`, implemented in pure Go (Plan 9) assembly for s390x, issuing
`SVC 35` on z/OS, **no cgo**, remains the deliverable. Phases 0–4 keep their roadmap
definitions. Every deliverable listed under them still ships.

**This ADR changes who does the work, in what order, and by when. It changes nothing about
what the work is.** A reader comparing the recalibrated plan against the PDF should find
the same destination reached on a different schedule by more hands.

## 3. Decision: five seats, defined by workstream

The team is organised as five standing workstreams, one per person, rather than as a queue
of tickets against a shared backlog. The reason is specific to this project: **its doctrine
is unusually heavy** — pre-registered hypotheses, provenance headers, ADR discipline, the
citation rules in `docs/goal-prompt.md` §4.1 — and that discipline is learned by owning an
area, not by picking up whatever is next.

| Seat | Workstream | Owns |
|---|---|---|
| **W1** | Service semantics & assembly | Phase 2 annotation, WPL, `SVC 35` encoding, later WTOR |
| **W2** | Go module & toolchain | `zbridge/`, the IBM Go fork, the `GOOS=zos` gate |
| **W3** | Emulation & lab operations | TK5/Hercules harness, QEMU, reproducibility, CI |
| **W4** | Evidence, hypotheses & publication | provenance audits, ADRs, the thesis-facing writing |
| **W5** | Lead & integration (owner) | mentor interface, roadmap, ADR approval, final review |

The full charter — entry state, first tickets, review gates, and the contribution workflow —
is `docs/team/charter.md`. It is an operational document and may change without an ADR; this
section fixes only the *shape*.

## 4. Decision: contribution flows through a ticket → PR gate, and doctrine is enforced at it

Solo, the project's doctrine was held in one person's head and in `CLAUDE.md`. That does not
survive four more contributors. The doctrine therefore moves to the merge gate.

Every change enters as a ticket, leaves as a pull request, and a PR does not merge until the
checks that correspond to the project's existing rules pass. Those rules are unchanged — what
changes is that they are now *checked* rather than *remembered*:

| Doctrine rule | Where it is now enforced |
|---|---|
| No emulated result presented as a z/OS result (ADR 0001 §7) | `evidence-provenance-auditor` on any PR touching `docs/evidence/` |
| Pseudo-register and frame-contract discipline | `go-asm-reviewer` + the existing `go vet` hook on any `.s` file |
| No performance number from an emulator | provenance auditor; explicit PR checklist item |
| `UNDEF` stubs fail loudly; no reaching green by weakening | lead review, mandatory on any stub or test change |
| BSD-3-Clause attribution survives refactors | existing confirmation hook on `tables.go` / `LICENSES.md` |
| No claim without citation, evidence, or registration | PR template requires naming which unknown the change retires |

## 5. Decision: phase windows re-anchor to the 2026–27 academic year

The roadmap's calendar windows (p.3) are superseded by the schedule in
`docs/roadmap-2026-27.md`. In summary:

| Block | Window | Content |
|---|---|---|
| **Pre-semester** | 2026-07-30 → late Sept 2026 | Phase 2 in full; team forms; access questions pursued |
| **Fall semester** | Oct 2026 → Jan 2027 | Phase 3a and the remainder of 3b; U3 work as access permits |
| **Spring semester** | Feb 2027 → June 2027 | Phase 3c, Phase 4, publication, thesis write-up |

**The schedule carries one dependency it cannot absorb.** Everything in Phase 3a onward
needs IBM's Go fork, and Phase 3b step 1 and all of 3c additionally need an entitled z/OS
system. Neither has a date. The plan is therefore built so that **no unentitled work is ever
waiting on entitled work** — which is the same structural principle the E-ladder was built
on, applied to the calendar.

## 6. Decision: U3 is **not** re-split by this ADR

ADR 0001 §3 defines U3 as a single unknown. ADR 0002 briefly split it and was withdrawn;
`docs/goal-prompt.md` §2 records that U3 is unsplit again. It is tempting to split it now,
because the toolchain dependency and the system dependency are visibly different things and
resolve through different channels.

**This ADR does not do that.** Amending ADR 0001's unknown decomposition is a decision in its
own right, it touches the structure every evidence file's `speaks_to:` header points at, and
this project has already spent one ADR withdrawal on moving faster than the evidence.

What this ADR does instead is record an **ordering fact inside U3**, which requires no
amendment: *you cannot run a binary on a system you have no compiler for.* The toolchain
question is upstream of the access question in execution order even though it is downstream
in urgency, and the plan sequences them accordingly. Whether that ordering deserves to become
U3a/U3b in the formal decomposition is left open and is listed in §9 as a reopen condition.

## 7. Decision: the mentor relationship is informed, not renegotiated

The owner's ruling is that Jürgen is **told** the project has scaled to a team with Mert as
lead, and that **nothing is asked of him about it** — no request for a changed cadence, no
request to mentor four more people, no new commitment implied. His interface stays exactly
what it was: one point of contact, the same technical questions, the same review offer on the
Phase 2 walkthrough that the roadmap already asked for (p.5).

All mentor-facing communication continues to route through W5. That is not ceremony — it is
the mechanism that keeps `docs/goal-prompt.md` §4.1's citation discipline attached to
everything IBM sees.

## 8. What this decision does not claim

- **It does not claim the schedule will hold.** Two of its three blocks depend on entitlements
  that do not exist yet. §5 claims only that the *unentitled* work is scheduled so as never to
  idle behind them.
- **It does not claim five people is five times the throughput.** Phase 2 parallelises across
  nine routines; Phase 3b step 1 does not parallelise at all, because it is one person on one
  borrowed system. The plan's fall block is deliberately not compressed to reflect headcount.
- **It does not claim the new contributors have mainframe experience**, and the plan does not
  assume it. It assumes competent systems programmers, which is what the owner stated.
- **It does not change any unknown's status.** U1 and U2 were retired by evidence and stay
  retired; U3 is open and this ADR does not narrow it.
- **It does not claim the 2026-07-27 briefings are obsolete.** They were never delivered.
  Their content is carried forward, not replaced.
- **It does not authorise publication.** `docs/goal-prompt.md` §5 boundary 3 is untouched:
  anything public is still the owner's call, and a larger team makes that boundary more
  load-bearing rather than less.

## 9. What would reopen or reverse this decision

1. **The team does not actually form, or forms smaller.** The workstream shape assumes five
   seats. At three, W3 folds into W2 and W4 folds into W5, and the fall block loses Phase 4
   preparation. Revisit rather than silently under-staff.
2. **z/OS access arrives early and generously.** Then the fall block re-cuts around machine
   time, and ADR 0001's own reopen-condition 1 fires alongside this one.
3. **The Go fork turns out to be unobtainable for this project.** That is the failure this
   plan cannot absorb, because §2 refuses the cgo hedge on the reasoning in ADR 0004 §3. It
   would force a genuine scope decision, and it belongs to the owner and the mentor, not here.
4. **The ordering fact in §6 starts doing load-bearing work.** If plans, evidence headers, or
   mentor questions begin routinely distinguishing the toolchain from the system, the
   decomposition should be formally amended by an ADR rather than left implicit.
5. **The PR gate proves to be friction without benefit.** If §4's checks reject nothing real
   over a semester, they are ceremony and should be cut to the ones that caught something.

## 10. Links

- `zbridge-asm-roadmap.pdf` — the mandate; pp.3 and 7 affected by §5
- `docs/roadmap-2026-27.md` — the recalibrated plan this ADR authorises
- `docs/team/charter.md` — the seats, workflow, and review gates fixed in shape by §3 and §4
- `docs/roadmap-errata.md` — entry S4 records the supersession
- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — §3's decomposition, left intact by §6
- `docs/decisions/0004-roadmap-corrections-and-cgo-scope-closure.md` — the cgo closure §8.3 depends on
- `docs/goal-prompt.md` — §5's autonomy boundaries, unchanged by §7
- `docs/mentor-briefings/2026-07-30-progress-and-hercules.md` — the undelivered checkpoint, carried forward

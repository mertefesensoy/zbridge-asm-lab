# Research Brief 003 · Where the WTO parameter list is *actually* documented, and what comes back in R15

- Status: **ANSWERED** — returned 2026-07-26
- Date: 2026-07-25
- Requested by: Claude (architecture role)
- Consumed by: `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — C1, C2 and C4; `docs/decisions/0001-emulation-strategy-hercules-two-track.md` §6 (the Phase 3b retirement table)
- Return: `research/003-wto-wpl-layout-source-and-return-code-contract.md`
- Supersedes the unmet parts of brief 001. **Does not** re-ask what 001 answered well.
- Priority: **highest open item.** Rung E2 still cannot be designed without Q1/Q2,
  and ADR 0001 §6 has a table entry in doubt until Q4 is answered.

## Why this brief exists — read this part carefully

Brief 001 was answered on 2026-07-25 and the return was audited against the primary
source it cited. The audit is
`docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md`. Two things came out of
it, and they set the shape of this brief.

**First: the manual brief 001 pointed at does not contain the answer.**
*OS/VS2 MVS Supervisor Services and Macro Instructions* (GC28-0683-2) was retrieved
and read page by page — pp. 75–77 (`Communicating with the System Operator`) and
pp. 208–213 (`WTO`, `WTO (List Form)`, `WTO (Execute Form)`). **The WTO parameter
list byte layout is not documented in that manual at all.** The `WTO (List Form)`
section describes only macro syntax and refers the reader back to the standard form.
The quotation the return attributed to it does not appear in it.

That was Claude's error as much as the researcher's: brief 001 named the wrong
manual and asked for a verbatim quotation from a book that does not contain the
material. **This brief's first job is to find the right book.**

**Second: a risk nobody had registered surfaced.** GC28-0683-2 p. 210 states, of the
return-code table it prints:

> Note: No return codes are issued by the WTO service routine if the MLWTO feature is
> not used.

and p. 210 also states that on return *"general register 1 contains the
identification number (24 bits and right-justified) assigned to the message."* On
that release, a **single-line** WTO — the exact case this project builds — appears to
issue **no return code at all**.

The project's endgame signature is `WTO(message string) error`. ADR 0001 §6 lists
Phase 3b step 6, *"Read R15, map to a Go error"*, as retired by rung E3. If z/OS
behaves like its ancestor here, there may be nothing to read. If z/OS *diverges* from
its ancestor here, that is a documented incompatibility in the interface this whole
project rests on — and per brief 001's own framing, an explicit divergence is worth
more than every other answer combined.

Either way the answer changes the design. That is why this is priority one.

## Scope note

Everything below concerns the **simple form**: single-line, problem state, no routing
codes, no descriptor codes, no extended WPL. Multi-line WTO, WTOR, connect IDs, and
the `__console2()` C path all remain out of scope, as in brief 001.

## The questions

Answer for **both** systems, separately and explicitly, as before:

- **System A — MVS 3.8** (the TK5 guest). **Note the release**: TK5 runs MVS **3.8j**;
  the manual retrieved so far is Release **3.7**. If a Release 3.8 edition exists,
  prefer it and say which you used.
- **System B — current z/OS** (the deployment target). Name the release.

### Q1 — Which publication documents the WPL byte layout? *(the unblocking question)*

GC28-0683 does not. **Which IBM publication does?** Candidates worth checking, not a
closed list:

- *OS/VS2 System Programming Library: Supervisor* (GC28-0628) — named in
  GC28-0683-2's own preface as the system-programmer companion volume
- The *OS/VS2 MVS Debugging Handbook* series — these carry control-block layouts
- *z/OS MVS Data Areas* (multi-volume) — the modern home of mapped control blocks
- The mapping macro itself: does a **`IHAWPL`** / **`IEZWPL`** (or similarly named)
  DSECT-generating macro exist, on either system? If so, **that macro source is the
  authoritative layout**, and on MVS 3.8j it ships in the distributed macro library
  where it can be listed directly.

State the form number, the section, and the page. If the layout is documented *only*
by a mapping macro and not in prose, **say so explicitly** — that is a legitimate and
important finding, and it would mean the E1 assembler listing is the primary
instrument for MVS rather than a cross-check.

### Q2 — The length field, restated

Once Q1 identifies the right source: at offset 0 there is a halfword length. **Which
bytes does it count** — its own 2, the flags halfword, the text, trailing code areas?
Quote the source verbatim, per system, with form number and page.

The working assumption in the repo is **text length + 4** (i.e. header-inclusive).
It is currently supported only by a vendor documentation page, which the brief
protocol classifies as corroboration and never as a primary citation. Confirm or
refute it against a primary source.

### Q3 — The MCS flags halfword, bit by bit *(brief 001's Q3, unanswered)*

Enumerate the bits at offset 2 on each system: name, position, meaning, and whether a
plain problem-state single-line message requires it set or clear.

The question that actually matters, restated because the previous return addressed it
only narratively: **is there any bit that is "reserved, must be zero" on MVS 3.8j and
has a defined, load-bearing meaning on current z/OS?** The extended-WPL selector is
the obvious candidate. A single such bit is an H001 C1 failure and would change the
E-ladder's value. An enumerated bit table from a mapping macro or data-areas volume
settles it; prose about `PLISTVER` does not.

### Q4 — The return-code contract *(the new one, and the reason for urgency)*

For a **single-line, non-MLWTO** WTO, on each system:

1. **Is a return code issued at all?** MVS Release 3.7 documentation says no — confirm
   for MVS 3.8j, and answer independently for z/OS.
2. If one is issued, **in which register**, and what is the **complete documented list
   of values and meanings**? Is 0 the only success value?
3. **What does R1 contain on return?** MVS Release 3.7 says the 24-bit message
   identification number. Does z/OS still do this?
4. Is there a **documented divergence** between the two on any of the above? An
   explicit IBM statement either way is the single most valuable item in this brief.

### Q5 — Out-of-range and error behaviour *(brief 001's Q7, unanswered)*

What is documented to happen if the length field is out of range, or the parameter
list sits above the applicable addressing boundary (16 MB line on MVS, 2 GB bar in
AMODE 31 on z/OS)? Abend code, return code, or undefined? Cite it.

The return for brief 001 attributed an abend sentence to GC28-0683; that sentence is
not in that manual. If such a statement exists, **find where it actually lives.**

### Q6 — The documented macro expansion *(brief 001's Q5, still wanted)*

Does any IBM publication print the actual generated expansion of `WTO 'text',MF=L`?
If yes, reproduce it verbatim with its citation. This is directly comparable to the
listing rung E1 will capture from IFOX00, which makes it the highest-value
cross-check available before machine time.

A reconstruction is **not** what is wanted here. If no published expansion exists,
state that plainly — the answer "IBM never printed it" is itself a finding and would
explain why brief 001 could not be satisfied.

## Required sources

**Primary — these are what count:** IBM publications with **form number, section, and
page**, or IBM-distributed macro source. For MVS 3.8j, the distributed macro library
(reachable through TK5, or through the CBT Tape archive) counts as primary for a
mapping macro; cite the member name.

Bitsavers, the Internet Archive, and CBT Tape are legitimate **access routes** — cite
the publication by form number and note the route.

**Corroboration only, never the primary citation:** Micro Focus / Broadcom / BMC
vendor documentation, Jay Moseley's Hercules materials, SimoTime, forums, blogs.
Several of these are written by people who know this material deeply and they are
genuinely useful for *locating* a source — but the previous return shows exactly what
happens when corroboration is presented with a primary-source label.

## Protocol reminders — these were the failure points last time

1. **A citation must support the specific sentence it is attached to.** If a fact
   comes from a vendor page and the form number comes from elsewhere, they are two
   different sources and must be reported as two different sources.
2. **Do not reconstruct.** An `INFERRED` reconstruction offered where documented
   material was requested reads as an answer and is not one. If it cannot be found,
   the correct output is "not found."
3. **Quote, do not paraphrase, anything presented in quotation marks.**
4. **Report page numbers.** "GC28-0683" is not checkable; "GC28-0683-2 p. 210" is —
   and page-level precision is what let the previous return be audited at all.

## Acceptance criteria

1. **Q1 answered with a specific publication or macro name, per system.** If the
   layout is undocumented in prose on a system, that is stated explicitly.
2. Q2, Q3, Q4 answered per system, each with form number **and page**, or an explicit
   "not found."
3. Q4 in particular carries an unambiguous yes/no on *"is a return code issued for a
   single-line WTO"* for **each** system.
4. Every claim tagged `FOUND` / `INFERRED` with confidence.
5. An updated **side-by-side divergence table** covering layout, length semantics,
   MCS flag bits, return-code contract, and R1-on-return.
6. Anything not found is listed in a short explicit "could not establish" section.

## How the result will be used

- **Q1/Q2** unblock rung E2's hand-built parameter list, which is currently blocked on
  having any citable layout at all.
- **Q3** closes or fails H001 C1.
- **Q4** determines whether ADR 0001 §6's Phase 3b retirement table is correct.
  If a single-line WTO issues no return code on the ancestor system, step 6
  (*"Read R15, map to a Go error"*) is **not** retirable by rung E3 and the ADR needs
  an owner-approved amendment. Per `docs/goal-prompt.md` §5 boundary 4, that
  amendment is the owner's call, not Claude's.
- **Q6** becomes the cross-check against the E1 listing.

If Q4 shows the two systems genuinely diverge on the return-code contract, that is a
publishable finding in its own right: the ancestor interface and its descendant
disagree about how a program learns whether its message was accepted, and no one has
written that down for a Go audience.

# Hypothesis 001 · MVS 3.8j's `SVC 35` is a valid semantic oracle for the z/OS WTO parameter list, such that a byte sequence verified on TK5 transfers to z/OS as a correct simple-form WPL

- Status: **IN PROGRESS** (updated 2026-07-25). Line 1 partially returned and audited
  against primary source. Lines 2 and 3 not run. **No sub-claim is resolved.**
- Date: 2026-07-25
- Author: Claude (hypothesis and architecture role)
- Ladder rung: E2 (documentary + TK5 empirical), closed at T3 (z/OS empirical)
- Builds on: ADR 0001 (`docs/decisions/0001-emulation-strategy-hercules-two-track.md`),
  which adopts Track M explicitly *without* assuming this hypothesis; the roadmap's
  Phase 3b step list; `docs/codex-handover.md` §3 (the WTO call-path decomposition)

---

## Evidence log — Line 1 (added 2026-07-25)

**Everything below the horizontal rule that follows this section is the original
pre-registration text, preserved unchanged**, per `docs/hypotheses/README.md`. This
section records what arrived; it does not edit what was predicted.

### What arrived

The brief 001 return (`research/001-wto-parameter-list-authoritative-layout.md`)
landed 2026-07-25. It answers all seven questions and tags every claim. It was then
audited against the primary source it cites — GC28-0683-2 read directly, pages
75–77 and 208–213 — and the audit is captured in
`docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md`.

**The audit found the citation does not support the central claim.** The manual does
not document the WTO parameter list byte layout anywhere, and the quotation the
return attributes to it does not appear in it.

### Sub-claim status after Line 1

| | Sub-claim | Status | Basis |
|---|---|---|---|
| **C1** | Field layout identical (2-byte length, 2-byte MCS flags, EBCDIC text) | **OPEN — shape uncited** | The shape is consistent across secondary sources and one vendor doc, but no primary citation was obtained for *either* system. The MCSFLAG bit-level question the brief's Q3 was written to answer was **not answered** — the return gave an `INFERRED` narrative that does not enumerate the bits. The C1 killer (a bit that is zero-and-reserved on MVS but load-bearing on z/OS) is therefore **not excluded**. |
| **C2** | Length-field semantics identical | **OPEN — reverted to registered assumption** | text+4 is corroborated by a vendor doc but has **no primary citation on either system**. This is the exact question the brief existed to settle, and it is not settled. Only the E1 listing can settle the MVS side. |
| **C3** | R1 in, `SVC 35`, R15 out | **MVS half CONFIRMED for R1-in; R15-out CONTRADICTED** | GC28-0683-2 p.212 confirms R1 carries the parameter list address. But p.210 states *"No return codes are issued by the WTO service routine if the MLWTO feature is not used"*, and R1 returns the message identification number. |
| **C4** | R15 = 0 means accepted, on both | **MVS half CONTRADICTED (documentary, Release 3.7)** | Per the above: a simple-form WTO on MVS Release 3.7 issues **no return code at all**. There is nothing in R15 to compare against zero. |

### The C3/C4 finding, and why it matters more than C2

The brief was commissioned to settle C2. The audit instead surfaced a problem in C3/C4
that nobody had registered as a risk:

> On MVS Release 3.7, a **single-line** WTO issues no return code. The documented
> return-code table (00/04/08/0C/10/14) applies **only to MLWTO**, which this
> hypothesis explicitly scopes out.

If this holds for MVS 3.8j at rung E1/E2, then the E-ladder cannot exercise the
`R15 → Go error` mapping at all, because on this system there is no R15 value to map.
That bears directly on ADR 0001 §6, which lists Phase 3b step 6 (*"Read R15, map to a
Go error"*) as ✅ retired by rung E3. **That table entry is in doubt.** Per
`docs/goal-prompt.md` §5 boundary 4, this is not being rewritten unilaterally — the
evidence is brought, the decision stays the owner's.

### A defect in the pre-registered decision rule, recorded not repaired

The decision rule below names outcomes for C2 failing (`E-ORACLE-PARTIAL`) and for C1
or C3 failing (`E-ORACLE-WEAK`). **It does not say what happens if C4 alone fails.**
That is a genuine gap in the pre-registration, and it is being recorded rather than
patched, because silently editing a decision rule after seeing evidence is exactly
what pre-registration exists to prevent.

Proposed resolution, for owner sign-off *before* E1 runs: a C4-only failure caused by
the *absence* of a return code on the ancestor system is an `E-ORACLE-PARTIAL`
outcome, not `WEAK` — the WPL construction still transfers, only the status-handling
half does not. It should be adopted, if adopted, by amending this hypothesis with a
dated note before evidence is collected, never after.

### What Line 1 did not deliver

Against the brief's own acceptance criteria: no form-numbered citation for Q1, Q3, or
Q4; no form number or release for the z/OS side of Q2; Q3's bit map not produced;
Q4's "is 0 the only success value" not answered; Q7's out-of-bounds behaviour not
answered; the documented macro expansion for Q5 not found (the return supplied an
`INFERRED` reconstruction instead). Brief 003 exists to close these.

**Line 1 is therefore recorded as PARTIALLY RETURNED, not complete.**

---

## Evidence log — Brief 003 (added 2026-07-26)

Brief 003 (`research/003-wto-wpl-layout-source-and-return-code-contract.md`) landed 2026-07-26 and supersedes the missing pieces of Brief 001.

### What arrived

- **C1/C2 (WPL Layout):** The byte layout for MVS 3.8j is definitively **not documented in prose** and exists only in the `IEZWPL` mapping macro. The "text length + 4" rule is inferred from reading the macro. On z/OS, the layout is documented in *z/OS MVS Data Areas*.
- **C3 (Return Code Contract):** A **major documented divergence** exists. MVS 3.8j issues **no return code** in R15 for a single-line WTO. z/OS **does** issue a return code in R15.
- **Missing Information:** The explicit bit-by-bit table for MCS flags (Q3) was not found, nor was a published `WTO` macro expansion (Q6).

### Sub-claim status after Brief 003

| | Sub-claim | Status | Basis |
|---|---|---|---|
| **C1** | Field layout identical | **OPEN** | Still requires the E1 assembler listing on MVS 3.8j to establish the baseline mapping macro shape. |
| **C2** | Length-field semantics identical | **OPEN** | Still lacks a prose citation on MVS 3.8j. Relies on the E1 listing. |
| **C3** | R1 in, `SVC 35`, R15 out | **CONTRADICTED (MVS side)** | Brief 003 confirmed that MVS 3.8j does not issue a return code in R15 for single-line WTO. |
| **C4** | R15 = 0 means accepted, on both | **CONTRADICTED (MVS side)** | No return code means there is nothing to compare to 0. |

### Conclusion on Brief 003

The contradiction on C3 and C4 validates the "E-ORACLE-PARTIAL" fallback proposed after Line 1. A single-line WTO on MVS 3.8j issues no return code, so the status-handling half of the WPL construction does not transfer directly to z/OS. E1 is now strictly required to observe the macro expansion and answer C1/C2, as IBM prose documentation does not exist for the MVS 3.8j layout.

---

ADR 0001's entire return on investment rests on one unproven claim: that work done
against MVS 3.8j's `SVC 35` tells us something true about z/OS's `SVC 35`. If that
claim holds, four of the six Phase 3b steps are retired before mainframe access
exists. If it fails, Track M is an interesting history lesson and nothing more.

That is far too much weight to rest on an assumption, so it is registered as a
hypothesis with a decision rule fixed *before* any evidence is gathered.

There is a specific reason for caution rather than confidence. During the ADR's
research pass, the *shape* of the WTO parameter list was easy to confirm from
multiple secondary sources — a halfword length, a halfword of MCS flags, then EBCDIC
message text. But the **semantics of the length field could not be pinned down from
any secondary source**: whether it counts the 4-byte header, and whether it counts
trailing routing/descriptor code areas, was either unstated or stated
inconsistently. That is precisely the kind of detail that produces an S0C4 or a
silently truncated console message, and precisely the kind of detail that a student
project would otherwise discover at rung T3, on borrowed time, in front of an
operator.

The prior is genuinely favourable — `SVC 35` is among the most stable interfaces in
computing history, and z/OS's WTO is the lineal descendant of this exact code — but
"probably compatible" is not a thesis-grade claim, and the divergences below are real
enough that guessing is not acceptable.

## What is being claimed, precisely

The hypothesis decomposes into four independent sub-claims. They are listed
separately because they can fail separately, and the decision rule depends on
*which* fails.

- **C1 · Field layout.** The simple-form (single-line, no routing/descriptor codes)
  WTO parameter list has the same field order and widths on MVS 3.8j and z/OS:
  a 2-byte length, a 2-byte MCS flag field, then EBCDIC message text.
- **C2 · Length-field semantics.** The 2-byte length counts the same bytes on both
  systems (the open question: header-inclusive or text-only; and whether trailing
  code areas are counted).
- **C3 · Invocation contract.** On both systems, R1 holds the parameter-list address,
  the message is issued by `SVC 35`, and R15 holds the return code on return.
- **C4 · Success semantics.** R15 = 0 indicates the message was accepted, on both.

**The hypothesis is the conjunction C1 ∧ C2 ∧ C3 ∧ C4.**

## Known and suspected divergences (registered before testing)

These are stipulated as *differences that exist*, not as threats to the hypothesis.
The hypothesis concerns the simple-form WPL only, and these sit outside it. Recording
them now prevents post-hoc rationalisation later.

| Divergence | MVS 3.8j | z/OS | Bears on the hypothesis? |
|---|---|---|---|
| Addressing mode | 24-bit; WPL below the 16 MB line | AMODE 31; WPL below the 2 GB bar | No — same category, different bound. Allocation is U3, out of scope here. |
| Architecture | S/370 | z/Architecture | No — the WPL is data, not instructions. |
| Extended WPL (WPX) | Does not exist | Exists; `CONSID=` sets an MCSFLAG bit selecting it | No — simple form only. **But**: if a bit we set as zero on MVS 3.8j is *defined* on z/OS, that is a C1 failure and must be caught. |
| Multi-line WTO (MLWTO) | Limited | Full connect-ID model | No — out of scope. |
| Assembler | IFOX00 (Assembler XF), S/370 only, no HLASM extensions | HLASM | No — affects how we *write* the test, not what the WPL is. |
| Console prefixing | Authorization-dependent | Authorization-dependent (blank / `@` / `*` / `+`) | No — cosmetic, and expected. |
| Message length limit | Shorter maximum | Longer maximum | Only if the E2/E3 test message approaches either limit. **Test messages stay under 40 bytes** to keep this out of play. |

## The instrument

Three independent lines of evidence, two of which can run before z/OS access.

### Line 1 — Documentary (Gemini, brief 001)

Compare the WTO macro expansion as documented for MVS 3.8 (*OS/VS2 MVS Supervisor
Services and Macro Instructions*) against current z/OS (*MVS Programming: Assembler
Services Reference*). Required output is a field-by-field table with byte offsets and
an explicit statement of what the length field counts on each system, each citation
traced to an IBM manual with its form number — not to a forum post or a blog.

### Line 2 — Empirical on MVS 3.8j (rungs E1 → E2)

This line is stronger than the documentary line for the MVS side, because **the
assembler listing is ground truth**. IFOX00 prints macro expansions, so:

1. **E1.** Assemble and run a program using the `WTO` macro. Capture the full
   assembler listing. The expansion shows the exact bytes IBM's own macro generates —
   the authoritative MVS 3.8j layout, obtained without trusting any manual.
2. **E2.** Hand-build a byte-identical parameter list with `DC` constants, load R1,
   issue a raw `SVC 35`, no macro anywhere. Confirm the console output is identical.

E1 establishes the layout; E2 proves we can reproduce it from first principles. E2 is
the rung that actually retires U2.

### Line 3 — Empirical on z/OS (rung T3, after access)

Fire the byte sequence validated at E3 on real z/OS and confirm the console message.
**This line cannot run before access, so the hypothesis stays open until T3.**

## Pre-registered decision rule

Fixed before evidence. The outcome selects among three differently-priced futures.

**E-ORACLE-STRONG** — C1 ∧ C2 ∧ C3 ∧ C4 all hold.
→ E2/E3 evidence transfers to z/OS as design evidence. Rung T3 is re-scoped from
"construct and debug a parameter list" to "port a verified byte sequence and handle
the U3 wrapper (Malloc31, AMODE)." The Go-side WPL builder is written once, target-
independent. This is the outcome ADR 0001 is betting on.

**E-ORACLE-PARTIAL** — C1 ∧ C3 ∧ C4 hold, C2 differs (or a single field width or
flag-bit meaning differs).
→ Still a substantial win, and arguably a *more* interesting finding. The Go-side
builder becomes parameterised by target, the delta is documented in a table, and E3
still validates the construction machinery. T3 keeps a WPL-layout debugging budget.

**E-ORACLE-WEAK** — C1 or C3 fails.
→ Track M's scope collapses to operational learning (JCL, console, MVS linkage,
mainframe fluency). The E-ladder stops at E1; E2 and E3 are cancelled. ADR 0001 is
revisited under its own reopen clause 2. **The divergence is written up and
published anyway** — "the ancestor interface is not a valid oracle, here is
precisely where it diverges" is a genuine contribution that nobody has documented.

**Tie-break and honesty clause.** If Line 1 (documentary) and Line 2 (empirical)
disagree, **Line 2 wins for the MVS side** — the assembler listing is what the system
actually does. A documentary/empirical disagreement is itself recorded as a finding,
because it means a published manual is wrong or ambiguous, which is worth knowing.

## Threats to validity

1. **Asymmetric evidence.** Until T3, this hypothesis is *empirical on the MVS side
   and documentary-only on the z/OS side*. A documentary error on the z/OS side would
   propagate undetected into E3. This is the single largest threat and it cannot be
   eliminated before access — only mitigated, by requiring IBM-manual citations with
   form numbers in brief 001 and by treating E-ORACLE-STRONG as *provisional* until
   T3 confirms it. **No published artifact may claim the oracle holds until T3.**
2. **The macro may not exercise the simple form.** IFOX00's `WTO` expansion may
   default to including routing/descriptor code areas. If so, E1's listing shows a
   richer WPL than the simple form, and E2 must be built against the simple form
   deliberately rather than by copying the expansion. Check the listing, do not assume.
3. **Success is confounded with tolerance.** A supervisor that ignores a malformed
   length field would let an incorrect WPL "pass" E2. Mitigation: a deliberate
   negative control — submit a WPL with a knowingly wrong length and confirm the
   console output is wrong or the call fails. **If the negative control also
   succeeds, E2 proves nothing** and the rung is void until a discriminating test is
   designed.
4. **EBCDIC code page.** MVS 3.8j-era EBCDIC is not identical to IBM-1047 in the
   national-use positions. Test messages use **uppercase A–Z, digits, and space
   only**, which are invariant across the relevant code pages, so E2/E3 test the WPL
   and not the code page. Code-page fidelity is `ebcdic/`'s job and is already
   ICU-verified.

## What would falsify

- The assembler listing at E1 shows a field order or width incompatible with the
  documented z/OS simple-form WPL → C1 fails.
- The length field demonstrably counts different bytes on the two systems → C2 fails.
- MVS 3.8j uses a different register for the parameter list address, or the return
  code does not arrive in R15 → C3 fails.
- The negative control (threat 3) succeeds → the instrument is invalid; no claim
  either way until redesigned.

## Links

- `docs/decisions/0001-emulation-strategy-hercules-two-track.md` — the ADR this
  hypothesis is the load-bearing assumption of
- `docs/research-briefs/001-wto-parameter-list-authoritative-layout.md` — Line 1
- `docs/runbooks/tk5-hercules-setup.md` — the environment Lines 2 runs in
- `docs/evidence/` — E1/E2/E3 captures land here with provenance headers
- `ebcdic/` — the code-page work this hypothesis deliberately holds constant

---
rung:          DOC (documentary — see "Convention note" below)
date:          2026-07-25
machine:       n/a — documentary verification, no code executed
guest_os:      n/a
architecture:  n/a
emulator:      n/a
host:          n/a
speaks_to:     U2 (documentary half only — retires nothing empirically)
hypothesis:    H001 (Line 1)
verdict:       INCONCLUSIVE for C2 · CONTRADICTS the brief-001 return on two points
---

# DOC-001 · What GC28-0683-2 actually says about WTO

**Primary source read directly, page by page, on 2026-07-25.** This file exists
because the return for research brief 001 asserted a verbatim quotation from this
manual, and the brief's own acceptance criteria make that quotation the load-bearing
answer. It was checked. It does not appear where it was said to appear.

## Convention note — the `DOC` rung prefix

`docs/evidence/README.md` defines the rung namespace as `E0`–`E3` and `T0`–`T3`, all
of which are machine-executed. This file is a documentary verification and executes
nothing, so it fits none of them. Rather than mislabel it as a rung or leave a
substantial verification uncaptured, it is filed with a `DOC` prefix and the
machine-specific header fields set to `n/a`.

**This is a convention extension and is flagged for owner review, not assumed.** The
alternative considered was to bury the verification inside H001's Line 1 section; it
was rejected because a 25-page source check that contradicts a research return needs
to be independently citable, and H001 must preserve its pre-registration text.

## Source identification

| Field | Value |
|---|---|
| Title | *OS/VS2 MVS Supervisor Services and Macro Instructions* |
| Form number | **GC28-0683-2** |
| Release | Release 3.7 |
| Base date | April 1978 (stated implied date 3 April 1978) |
| Revision | Revised 30 June 1978 by TNL **GN28-2914** |
| Access route | bitsavers (`bitsavers.trailing-edge.com/pdf/ibm/370/OS_VS2/Release_3.7_1976/`) |
| Pages read | Front matter iii–x; 74–78; 87–90; 193–202; **208–213** |

**Release mismatch, stated plainly:** TK5 runs **MVS 3.8j**. This manual is
**Release 3.7**. It is the closest edition retrieved, not the exact one. A Release 3.8
edition may exist (bitsavers carries a `Release_3.8_1978` directory for other manuals
in the family). Every finding below is therefore a Release 3.7 finding, and the E1
assembler listing remains the tie-break for MVS per H001's honesty clause.

## Finding 1 — the manual does not document the parameter list byte layout

**The WTO parameter list byte layout does not appear in this manual.** It is absent
from all three places it could plausibly be:

- **pp. 208–210, `WTO — Write to Operator`** — documents macro *syntax*, `ROUTCDE`,
  `DESC`, routing and descriptor code meanings, and return codes. No field offsets.
- **p. 211, `WTO (List Form)`** — the section that generates the parameter list. It
  says only: *"The list form of the WTO macro instruction is used to construct a
  control program parameter list."* and *"The parameters are explained under the
  standard form of the WTO macro instruction."* **No byte layout, no offsets, no
  length-field description.**
- **pp. 75–76, `Communicating with the System Operator`** — narrative, routing and
  descriptor codes, sample macro invocations (Figure 43). No layout.

### Consequence for the brief-001 return

The return's Q2 answer — tagged `[FOUND] (High Confidence)` — attributes this
quotation to GC28-0683:

> "The length field (the first two bytes) contains the length of the message text
> plus 4."

**That sentence does not appear on any page of this manual that concerns WTO.** The
manual does not discuss the length field at all. The citation does not support the
claim it is attached to.

This is not proof the *semantics* are wrong — the text+4 reading is corroborated
elsewhere (see Finding 4) — but it means **H001 C2 has no primary-source citation and
remains open.** Per `docs/goal-prompt.md` §4.1 the claim is neither cited nor
evidenced; it reverts to registered-assumption status.

## Finding 2 — maximum message length is 124, not 126

The return states: *"Maximum length for a single-line WTO text is 126 characters."*

The manual, p. 208 and again p. 211, states:

> `msg`: Up to 124 characters.

**124.** Contradicted by the primary source. The return also attributes an abend
quotation ("If the length field contains a value less than 5, or greater than the
maximum allowed for the message type, the system issues an abend") to this manual;
that sentence likewise does not appear in the WTO section.

Practical impact is low — H001 threat 4 already caps test messages under 40 bytes —
but it is a direct factual contradiction in a claim tagged `[FOUND]`, which bears on
how much weight the rest of the return can carry.

## Finding 3 — the return-code contract is materially different, and this is the important one

The return's Q4 states, tagged `[FOUND] (High Confidence)`:

> Return code: Register 15 (R15). Return code 0: Processing completed successfully.
> **Both systems share this contract.**

The manual, p. 210, states three things that together contradict this for MVS:

> When control is returned, general register 1 contains the identification number
> (24 bits and right-justified) assigned to the message.

> Return codes from execution of a WTO using the **multiple-line feature** are as
> follows: [table: 00, 04, 08, 0C, 10, 14]

> **Note: No return codes are issued by the WTO service routine if the MLWTO feature
> is not used.**

Corroborated at p. 77 (`Message Deletion`):

> The control program assigns a message identification number to each WTO and WTOR
> message and returns the message identification number in register 1.

**On MVS Release 3.7, a single-line (simple-form) WTO issues no return code at all.**
R1 comes back holding the message identification number, not a status. The 00–14
return-code table applies only to MLWTO, which H001 explicitly scopes out.

## Finding 4 — what *is* confirmed

Not everything in the return failed. From primary source:

| Claim | Status | Where |
|---|---|---|
| R1 carries the parameter list address into WTO | **CONFIRMED** | p. 212, Execute Form Example 2: *"Write a message with a pre-built parameter list pointed to by register 1."* → `WTO MF=(E,(1))` |
| Message text is EBCDIC, non-printable characters replaced | **CONFIRMED** | p. 208: *"only standard printable EBCDIC characters are passed to the display devices. All other characters are replaced by blanks."* |
| Message assembled as a variable-length record | **CONFIRMED** | p. 208 |
| Console prefixing is blank / `@` / `*` / `+` (ADR 0001 evidence item 7) | **CONFIRMED** | p. 75 |
| Simple form needs no ROUTCDE/DESC | **CONFIRMED** | p. 210: with both omitted and not MLWTO, the `OLDWTOR` sysgen routing code is assigned; *"if the OLDWTOR parameter is omitted, no routing code is assigned."* |

Separately, an independent **corroborating** source (Micro Focus Enterprise Developer
WTO documentation, retrieved via search 2026-07-25, vendor doc — *corroboration only,
never primary per the brief protocol*) independently describes the MF=L expansion as:
first halfword = message length plus 4, second halfword = `DC H'0'`, then text. This
agrees with the return's *semantics* while leaving its *citation* unsupported.

## What this file does not claim

- It does not claim the text+4 semantics are wrong. It claims they are **uncited**.
- It does not claim MVS 3.8j behaves as Release 3.7 does. Release mismatch is stated.
- It says **nothing about z/OS.** The z/OS half of every H001 sub-claim remains
  documentary-only and now uncited; brief 003 exists to close that.
- It does not resolve C2. Only the E1 assembler listing can, for the MVS side.

## Links

- `research/001-wto-parameter-list-authoritative-layout.md` — the return audited here
- `docs/research-briefs/001-wto-parameter-list-authoritative-layout.md` — the brief
- `docs/research-briefs/003-wto-wpl-layout-source-and-return-code-contract.md` — the gap-closer
- `docs/hypotheses/001-mvs38j-svc35-wto-oracle.md` — Line 1 consumer

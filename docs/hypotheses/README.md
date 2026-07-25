# Hypotheses

Pre-registered, falsifiable claims with decision rules fixed **before** evidence is
collected. Format ported from the CASSANDRA project.

**Filename:** `NNN-slug.md`, three digits, sequential.

## Why pre-registration

This project's conclusions are going into a mentored thesis, and its subject matter
makes it unusually easy to fool yourself. Emulators are tolerant, ancestor interfaces
look compatible until they are not, and a console message appearing is a seductive
success signal that can mask a malformed parameter list. Deciding what would count as
success *before* running the experiment is the cheapest available defence.

The rule that carries the most weight: **a hypothesis is a bet, not a plan.** If it
falsifies, that is a result — and on this project, several of the possible negative
results are more publishable than the positive ones.

**Required sections:**

| Section | Purpose |
|---|---|
| Header block | Status · Date · Author · Ladder rung · Builds on |
| **Why this, and why now** | The pressure that makes this worth testing. |
| **What is being claimed, precisely** | Decomposed into independently falsifiable sub-claims (C1, C2, …), because *which* one fails changes what happens next. |
| **What is explicitly NOT being claimed** | Include whenever a nearby, tempting, stronger claim exists. |
| **The instrument** | How the claim gets tested, including independent evidence lines. |
| **Pre-registered decision rule** | Named outcomes, each with its consequence, written before evidence. |
| **Threats to validity** | Honest list, with mitigations. Negative controls belong here. |
| **What would falsify** | Concrete, observable outcomes. |

**Status values:** `PRE-REGISTERED` · `IN PROGRESS` · `RESOLVED, <outcome>` ·
`VOID (instrument invalid)`

On resolution, the status block is updated with the outcome, the date, the evidence
files, and a pointer to the ADR that closes it. **The original pre-registration text
is preserved unchanged below the updated status**, so the record shows what was
predicted, not what was rationalised.

## Index

| H | Claim | Status | Unknown |
|---|---|---|---|
| [001](001-mvs38j-svc35-wto-oracle.md) | MVS 3.8j `SVC 35` is a valid oracle for the z/OS WTO parameter list | **IN PROGRESS** — Line 1 partially returned; C3/C4 contradicted on the MVS side by primary source | U2 |
| [002](002-s390x-port-equivalence.md) | The `_s390x.s` bodies are behaviourally equivalent to amd64; emulation cannot measure performance | PRE-REGISTERED | U1 |

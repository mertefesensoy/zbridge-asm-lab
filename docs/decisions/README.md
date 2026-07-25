# Architecture Decision Records

Numbered, immutable-once-accepted records of decisions that shape the project.
Format ported from the CASSANDRA project, where it did its job well.

**Filename:** `NNNN-slug.md`, four digits, sequential, never reused.

**Required sections:**

| Section | Purpose |
|---|---|
| Header block | Status · Date · Author · Resolves · Builds on |
| **Context** | What situation forced a decision. Written so a reader who was not there understands the pressure. |
| **Evidence** | The verified facts the decision rests on, each traceable to a source. Facts only — no reasoning yet. |
| **Decision** | What was decided, numbered, unambiguous. |
| **Scope and what this decision does not claim** | The boundary. This section is what makes the record honest and is never omitted. |
| **What would reopen or reverse this decision** | Pre-stated falsification conditions, so revisiting is a normal event rather than an admission. |
| **Links** | Related ADRs, hypotheses, evidence, external sources. |

**Status values:** `Proposed` · `Accepted` · `Superseded by NNNN` · `Reversed`

**Rules.** An accepted ADR is not edited to change its meaning — it is superseded by
a new one. Corrections of fact are fine and are marked inline with a date. The
"does not claim" section is mandatory: on this project the boundary between what an
emulator proves and what only real hardware proves is the difference between a
credible thesis and an overstated one.

## Index

| ADR | Title | Status |
|---|---|---|
| [0001](0001-emulation-strategy-hercules-two-track.md) | Hercules is a two-track semantic laboratory, not a z/OS substitute | Accepted |

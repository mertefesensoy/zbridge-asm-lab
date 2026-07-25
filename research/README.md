# Research returns

Gemini's answers to the briefs in `docs/research-briefs/`, stored **verbatim** with
their sources.

**Filename:** matches the brief it answers — `NNN-slug.md`.

## Why verbatim

The return is the raw material; the synthesis happens in ADRs and hypotheses, where
it is attributed and can be checked against this file. Editing a return in place
destroys the ability to audit how a conclusion was reached — which is precisely what
a thesis reader is entitled to do.

If a return is wrong, that is noted **in the consuming document**, with the reason.
The return itself stays as it arrived.

## Reading a return

Returns carry `FOUND` / `INFERRED` tags and confidence ratings per the brief protocol
(`docs/research-briefs/README.md`). When synthesising:

- **`INFERRED` claims do not become `FOUND` claims by being copied into an ADR.**
  Carry the tag through, or verify independently and say so.
- Preserved contradictions between sources stay visible in the ADR. On this project
  source disagreement is often the finding, not noise to be resolved away.
- Corroborating sources (forums, blogs, vendor docs) are never promoted to primary
  citations during synthesis.

## Index

| File | Brief | Received | Audit |
|---|---|---|---|
| [001-wto-parameter-list-authoritative-layout.md](001-wto-parameter-list-authoritative-layout.md) | 001 | 2026-07-25 | **Central citation unsupported.** Audited against GC28-0683-2 directly — see `docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md`. Two `[FOUND]` claims contradicted by the primary source. Read the audit before citing anything from this return. |
| [002-prior-art-census-go-assembly-mvs-services.md](002-prior-art-census-go-assembly-mvs-services.md) | 002 | 2026-07-25 | Search surface documented; both negative existence claims UPHELD. Q2/Q5 thin (no abandoned-work sweep, no dates for `go-recordio`). Usable as a thesis appendix after a currency re-run. |

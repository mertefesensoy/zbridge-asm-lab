# Research Brief 002 · Prior-art census: does a public Go-assembly WTO implementation exist?

- Status: **OPEN**
- Date: 2026-07-25
- Consumed by: the roadmap's Thesis section; ADR 0001; any published artifact
- Priority: high — this validates the project's novelty claim

## Why this brief exists

The roadmap's thesis rests on a **negative existence claim**:

> "Research confirms no public Go-assembly WTO implementation exists. Python has it
> via Ambitus's pyzkiln. Java has it via JZOS. Go does not."

Negative existence claims are the most fragile kind of claim a thesis can make. They
cannot be proven, only *failed to be disproven* by a search whose thoroughness is
documented. And they decay: the claim was true when the roadmap was written, and a
repository published since then would invalidate it silently.

This is also the one research task in the project that genuinely requires exhaustive
external search rather than reasoning — which makes it the natural shape for Gemini's
role.

The purpose is **not** to confirm the comfortable answer. If a public implementation
exists, that is a finding worth far more than a defended novelty claim: it becomes
prior art to study, cite, and build on, and the project's contribution reframes from
"first" to "documented, reviewed, and reusable" — which was arguably the stronger
contribution anyway.

## The questions

### Q1 — Direct hit

Does any public repository, package, article, or conference talk contain a **WTO
implementation callable from Go without cgo** — that is, Go assembly (Plan 9 syntax,
`_s390x.s`) issuing `SVC 35` or otherwise reaching the operator console?

Search at minimum: GitHub, GitLab, pkg.go.dev, IBM's `ibmruntimes` organisation,
the Go module proxy index, SHARE and GSE presentation archives, IBM Developer,
Medium/dev.to, and the IBM Z Open Community.

Useful search vectors: `SVC 35` with `golang`; `WTO` with `go assembly`; `GOARCH=s390x`
with `.s` files; `GOOS=zos` repositories generally.

### Q2 — Adjacent Go-on-z/OS assembly work

Independent of WTO specifically: which public Go codebases contain **s390x or z/OS Go
assembly** that calls MVS services?

`ibmruntimes/go-recordio` is known (IEFSSREQ, BSD-3-Clause) and is this project's
architectural blueprint. **What else?** Anything doing Name/Token, ENQ/DEQ, STORAGE,
console services, or any other SVC or PC-routine call from Go assembly. Include
partial, abandoned, and experimental work — an abandoned attempt is prior art and
frequently the most informative kind.

### Q3 — Cross-language comparators

Confirm and characterise the implementations the roadmap names, and find any others:

- **Ambitus `pyzkiln`** (Python) — confirm it implements WTO; how does it reach the
  console (direct SVC, C shim, or LE service)? License?
- **JZOS** (Java) — same questions.
- Any Rust, Node.js, or other-language implementation of z/OS console services
  without going through C.

For each: **does it issue `SVC 35` directly, or route through a C runtime function
such as `__console2()`?** This distinction is the entire technical point of the
project — a language binding that calls a C shim is not the same achievement as one
that issues the supervisor call itself, and the census must not blur them.

### Q4 — The published-walkthrough gap

Separately from code: does a **published, instruction-by-instruction annotation of
`ibmruntimes/go-recordio`'s `utils/utils.s`** exist anywhere?

Roadmap Phase 2's deliverable is exactly that document, on the stated basis that
nobody has published one. Same fragility as Q1, same need to re-verify with a date.

### Q5 — Currency

For every item found, record **last commit or publication date**. A dormant
repository and an actively maintained one mean very different things for both the
novelty claim and for whether the work is usable.

## Required sources

Primary here is the artifact itself — the repository, the package page, the slide
deck. Cite the URL and the date accessed. For each hit, state the **license**, since
this project is BSD-3-Clause to match go-recordio and license compatibility governs
what can be reused.

## Acceptance criteria

1. A **census table**: name · language · what it implements · direct SVC or C shim ·
   license · last activity · URL.
2. An explicit verdict on the roadmap's claim, in one of three forms:
   - **UPHELD** — no public Go-assembly WTO implementation found, *with the search
     surface documented* (which sources, which query terms, what date). The
     documented surface is what makes the claim defensible; without it the verdict is
     worthless.
   - **REFUTED** — one or more found. Full detail on each. This is a good outcome,
     not a failure; report it plainly and without hedging.
   - **AMBIGUOUS** — something exists that partially qualifies (a C shim, an
     incomplete or abandoned attempt, an internal-only artifact). Characterise
     precisely where it falls short of the claim.
3. Same verdict structure for Q4 (the go-recordio walkthrough gap).
4. Every claim tagged `FOUND` / `INFERRED` with a confidence rating.
5. **A re-check date.** This claim decays; the census states when it should be run
   again — recommend before any public release of the module or the thesis.

## How the result will be used

The return lands at `research/002-prior-art-census-go-assembly-mvs-services.md`.

- **UPHELD** → the census table and its documented search surface become an appendix
  of the thesis. A novelty claim backed by a documented search is defensible; one
  backed by "research confirms" is not.
- **REFUTED** → this is genuinely good news. ADR 0001 and the roadmap's thesis
  section are revised, the found implementation becomes prior art to study and cite,
  and the project's contribution reframes toward the annotated walkthrough and the
  reviewed, reusable module — the parts nobody had done regardless.
- **AMBIGUOUS** → the thesis states the boundary precisely instead of claiming a
  clean first.

The one outcome that must not happen is a comfortable answer accepted without a
documented search behind it.

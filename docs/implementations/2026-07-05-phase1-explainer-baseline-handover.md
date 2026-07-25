# Phase 1 wrap-up: interactive explainer, mainframe baseline strategy, Codex handover

**Date:** 2026-07-05
**Status:** shipped

---

## 1. Problem / Motivation

Phase 1 of the roadmap (six Go-assembly exercises on amd64) is complete, but three
gaps blocked the next phase:

1. **No consolidated explanation** of what each module does and how — the knowledge
   lived in six separate READMEs and in the assembly comments. The owner requested an
   interactive artifact suitable for review, demos, and mentor meetings.
2. **z/OS access may arrive ahead of schedule**, out of order with the roadmap's
   Phase 1b (LinuxONE) → Phase 2 (annotation) → Phase 3 (z/OS) sequence. There was no
   documented plan for spending scarce mainframe time well.
3. **Continuation by another agent (Codex)** was planned, with no self-contained
   handover of project state, conventions, and guardrails.

## 2. What Changed

| File | Change |
|---|---|
| `docs/interactive/zbridge-module-explorer.html` | New. Single self-contained interactive page: per-module annotated assembly (click-a-line), live demos (real IBM-1047 AtoE table, WTO length-prefix builder, byte-compare stepper, syscall register choreography, SP/frame visualizer), WTO pipeline map, phase status. Zero dependencies; works offline. |
| `docs/mainframe-baseline-strategy.md` | New. Operating plan for early z/OS access: how the phase order bends, Day-0 checklist, T0→T3 test ladder with port order, laptop-only prep list, access-path table, risk register delta, and a "bridge baseline pending scope" section. |
| `docs/codex-handover.md` | New. Self-contained handover: project identity, verified state, architecture summary, eight conventions, verification commands, ordered next steps (first one blocked on owner-provided scope), guardrails, external references. |
| `README.md` | Updated. Was stale (listed 2 of 6 exercises); now catalogs all six plus the docs section. The full catalog is itself a Phase 1 roadmap deliverable. |
| `docs/implementations/_TEMPLATE.md` | New. Implementation-doc template per the global documentation-first convention. |
| `memory/MEMORY.md` | New. Project memory index for future sessions. |

No Go or assembly source files were modified.

## 3. Implementation Approach

**Explorer:** one HTML file, no external assets, so it can be double-clicked, mailed,
or committed without a build step. The annotated-code viewers are generated from a
single JS data structure (`MODULES`: array of `[codeLine, annotation]` pairs holding
the *verbatim* repo assembly), so code and commentary stay adjacent and greppable.
Demos re-implement each module's algorithm in JavaScript rather than pretending to run
Go: the EBCDIC demo embeds the exact 256-byte `atoeTable` from `ebcdic/tables.go` and
derives `EtoA` as its inverse permutation; the strmanip demo reproduces the
big-endian split including the length-300 (`0x012C`) test case; the bytecmp stepper
executes one `CMPQ` per click with the same `bylen` tiebreak as the assembly.

**Strategy doc:** written against the actual roadmap PDF (extracted and re-read this
session), organized around one principle — mainframe time is the scarcest resource —
and made operational via a gated test ladder (T0 pure Go → T1 `add` → T2 five ports →
T3 WTO) with per-rung evidence requirements.

**Handover:** structured for a cold-start reader: state before conventions,
conventions before tasks, tasks ordered with the blocked one (owner's bridge scope)
first so Codex does not scaffold the production module prematurely.

## 4. Mathematical / Statistical Details

Only one piece of math is load-bearing: the EBCDIC demo derives the reverse table as
the inverse permutation of the forward table (`ETOA[ATOE[i]] = i`). This is valid
because IBM-1047 ↔ ISO-8859-1 is a bijection on 0–255 (both tables are permutations;
the package's round-trip tests assert the same property). The demo also surfaces the
big-endian split used by `strmanip`: for a 16-bit length *L*, the header bytes are
`hi = L >> 8`, `lo = L & 0xFF`, stored `hi` first — z/Architecture byte order.

## 5. Design Decisions

- **Single HTML file vs. a small site or notebook:** a file survives being emailed to
  a mentor, opened from OneDrive, or committed to GitHub with zero infrastructure.
- **Demos in JS re-implementation vs. WebAssembly-compiled Go:** WASM would be
  "really running" the code but adds a build pipeline and hides the algorithm; the
  point of the page is explanation, and the JS mirrors the assembly line-for-line.
- **Absorbing Phase 1b into 3a rather than deleting it:** the LinuxONE path is kept
  as an explicit fallback; the strategy bends the order without burning the bridge.
- **Handover as a repo doc rather than a chat artifact:** it must survive the session
  and travel with the code.

## 6. Verification

```powershell
# 1. Repo state backing the docs' claims (run 2026-07-05, all green):
#    add, ebcdic, strmanip, regs, bytecmp:  go vet ./... && go test ./...   -> ok
#    syscall-linux:  GOOS=linux GOARCH=amd64 go vet ./... && go build ./... -> ok

# 2. Explorer: open docs/interactive/zbridge-module-explorer.html in a browser.
#    - Click assembly lines in each module: annotation panel updates.
#    - EBCDIC demo: type "HELLO WTO" -> bytes C8 C5 D3 D3 D6 40 E6 E3 D6
#      (verify 'H'=0x48 -> 0xC8 and space=0x20 -> 0x40 against ebcdic/tables.go).
#    - strmanip demo: press the length-300 button -> header bytes 01 2C.
#    - bytecmp demo: "WTO ROUTE" vs "WTO REPLY" -> steps to i=5, result +1
#      (O(0x4F) > E(0x45)).

# 3. Links: README docs section and the explorer's "road ahead" links resolve to
#    the two new markdown docs.
```

## 7. Related Docs

- [Interactive module explorer](../interactive/zbridge-module-explorer.html)
- [Mainframe baseline strategy](../mainframe-baseline-strategy.md)
- [Codex handover](../codex-handover.md)
- Roadmap: `zbridge-asm-roadmap.pdf` (owner's desktop, outside the repo)
- Per-module READMEs: `strmanip/`, `regs/`, `bytecmp/`, `syscall-linux/`, `ebcdic/LICENSES.md`

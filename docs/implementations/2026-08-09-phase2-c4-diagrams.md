# Phase 2: C4 model diagrams, all four levels

**Date:** 2026-08-09
**Author:** Claude Sonnet 5, in session with Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

Phase 1 (`docs/architecture/*.md`) explains the system in prose. The owner separately
asked for the same system redrawn as [C4 model](https://c4model.com) diagrams — Context,
Container, Component, and Code — for readers who build a mental model faster from a
picture than from text, and specifically asked that the diagrams be built with the
native visualization widget and exported as SVG into the repository rather than
described only in words. This is Phase 2 of the day-plan in
[GitHub issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1).

## 2. What Changed

| File | Change |
|---|---|
| `docs/architecture/c4/level1-context.svg` | New. System Context: zbridge, its two human actors, and the three external systems it depends on (two available now, one open). |
| `docs/architecture/c4/level2-container.svg` | New. Container: the four buildable units inside the repository and what each currently talks to. |
| `docs/architecture/c4/level3-component.svg` | New. Component: the zbridge library's eight Go packages, color-coded by why each is or isn't finished. |
| `docs/architecture/c4/level4-code.svg` | New. Code: the `console` package's actual call chain from `WTO()` down to the platform-gated `issueWTO` branch. |
| `docs/architecture/c4/README.md` | New. Walkthrough explaining what C4 is and how to read each of the four diagrams, cross-linked to the Phase 1 prose documents. |

No source or test files changed.

## 3. Implementation Approach

**Each diagram was built twice, deliberately.** First with the `mcp__visualize__show_widget`
tool (SVG mode) so it rendered inline in this session, using that tool's class-based
theming system (`c-teal`, `c-coral`, `class="box"`, etc.) which auto-adapts to light and
dark mode inside the chat host. Second, the *same* diagram was written to a standalone
`.svg` file in `docs/architecture/c4/`, with every themed class expanded to explicit
hex colors drawn from the same named color ramp (e.g. `c-teal`'s 50/600/900 stops
becoming literal `fill="#E1F5EE" stroke="#0F6E56"` / text `fill="#04342C"`) — a
standalone SVG file, viewed on GitHub or in an editor, has no access to the chat host's
CSS variables, so it needs real values to render at all, let alone correctly.

**Geometry was verified against the actual rendered DOM, not eyeballed.** Screenshot
capture was unavailable in this session's browser tooling, so each saved `.svg` file was
opened directly (`file://`) and inspected with `element.getBBox()` calls: every
`<rect>`'s and `<text>`'s actual on-screen bounding box was pulled and checked
programmatically for (a) any two text elements overlapping, (b) any element extending
past the safe viewBox area, and (c) the true rendered width of text against its
containing box — catching the exact class of bug (cramped labels, arrows slashing
through boxes, overflow past the edge) the visualization tool's own guidance warns is
the most common failure mode for this kind of diagram.

**Colour was used to carry a claim, not to decorate.** Level 1 distinguishes "usable
today" (QEMU/Hercules, teal) from "not reachable yet" (z/OS, IBM's Go fork, neutral
gray) — mirroring the U1/U2/U3 split in `docs/architecture/README.md`. Level 3 uses four
colors precisely because there are honestly four different statuses among the eight
packages (working; blocked on U3; not yet scoped; shared core), with a legend, rather
than collapsing them into one generic "in progress" shade. Level 4 colors the one file
that is compiled by nothing today (`console_zos.go`) differently from every file that
runs on every `go test`. In each case the color is defensible against
`docs/architecture/zbridge-module.md` and the evidence files, not chosen for visual
variety.

## 4. Mathematical / Statistical Details

None — this is a documentation/diagramming change with no algorithmic or statistical
content. (Layout used arithmetic box-packing, not statistics: box widths were sized
from estimated character-width-per-font-size, per the visualization tool's own
calibration table, then corrected against the actually-rendered `getBBox()` widths
described in §3 above.)

## 5. Design Decisions

- **Non-interactive SVGs, not the widget tool's default clickable nodes.** The
  visualization tool's guidance defaults to wrapping every diagram node in a
  `sendPrompt(...)` click handler. That's correct for an ephemeral chat widget and
  meaningless (or actively confusing — a hover cursor with no effect) for a static SVG
  asset saved into a public repository and viewed on GitHub or in an image viewer, so
  all `onclick`/`class="node"` interactivity was omitted from the saved files.
- **Four colors on Level 3, not the "≤2 ramps" default.** Considered collapsing to two
  colors (teal = done, gray = not done). Rejected because it would misrepresent two
  packages (`internal/svc` et al., blocked on U3) as being in the same state as two
  different packages (`subsys`/`dataset`, blocked on an owner scoping decision) when
  they are blocked for materially different reasons with different unblocking paths.
  Added a one-line legend, per the tool's own guidance for when color choices need to
  encode more than a simple in/out state.
- **`internal/linkage` and the "everyone imports the core" relationship are not fully
  arrow-connected on Level 3.** Considered drawing all eight packages' dependency
  arrows. Rejected as clutter — the tool's own complexity budget warns against dense
  arrow meshes — in favor of one representative arrow plus a prose note in
  `c4/README.md` that the omitted arrows exist and where to read about them in full
  (`zbridge-module.md`).
- **Mermaid was not used**, despite already being available in this repository
  (`docs/diagrams/wto-call-path.md`). The owner specifically asked for the native
  visualization widget and an SVG export, which is a different mechanism (raw SVG
  markup, not a Mermaid source string rendered at display time) — the two coexist
  rather than one replacing the other. `docs/diagrams/wto-call-path.md` remains the
  Mermaid-based companion for the WTO call path specifically, cross-linked from
  `wpl-svc35-mechanism.md`.

## 6. Verification

- Every saved `.svg` file was opened directly in a browser tab (`file://`) and checked
  programmatically: `document.querySelector('svg').getAttribute('viewBox')` confirmed
  each file parses as valid SVG (a parse error would replace the page title with a
  browser error page, and none did); `getBBox()` on every `<rect>` and `<text>`
  confirmed (a) zero pairwise text/text bounding-box overlaps on any diagram, and (b) no
  element's rendered extent falls outside the intended safe drawing area.
- Every box's title and subtitle text was checked against its containing rectangle's
  actual rendered width via the same `getBBox()` calls, not just the tool's character-count
  estimate — several subtitles were shortened during drafting when the estimate proved
  optimistic.
- Every fact represented in the diagrams (which packages are blocked and why, which
  build tag gates `console_zos.go`, which two emulators retired which unknowns) was
  cross-checked against `docs/architecture/zbridge-module.md`,
  `docs/architecture/evidence-ladder.md`, and the underlying source files read during
  Phase 1 — no new claims were introduced that Phase 1 doesn't already support.

## 7. Related Docs

- [GitHub issue #1](https://github.com/mertefesensoy/zbridge-asm-lab/issues/1) — the
  day-plan this is Phase 2 of.
- [`docs/architecture/c4/README.md`](../architecture/c4/README.md) and its four `.svg`
  files.
- [`docs/architecture/README.md`](../architecture/README.md) §6 — the reading-order map
  this folder fits into.
- [`docs/architecture/zbridge-module.md`](../architecture/zbridge-module.md),
  [`wpl-svc35-mechanism.md`](../architecture/wpl-svc35-mechanism.md) — the prose
  documents Level 3 and Level 4 respectively summarize visually.
- [`docs/diagrams/wto-call-path.md`](../diagrams/wto-call-path.md) — the pre-existing
  Mermaid-based diagram set, a different mechanism covering related ground.

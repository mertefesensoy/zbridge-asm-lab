# Authorship attribution cleanup ahead of university outreach

**Date:** 2026-08-26
**Author:** Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

The repository is about to be shown to university contacts as the public face of the
thesis project. Three attribution problems would have met that audience.

**A live AI co-author trailer on GitHub.** One commit reachable from the remote —
`75e6c87`, the sole commit of the pushed branch `docs/roadmap-2026-27-pdf` — carried
`Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>`. GitHub renders such a trailer
as a second author avatar on the commit. This was the only surviving instance on any
pushed ref.

It is worth recording precisely what was *not* wrong, because a first scan suggested a
much larger problem than existed:

- `main`'s history was already clean — all 25 commits authored and committed solely by
  the owner, zero trailers. A prior cleanup had stripped them; six pre-amend originals
  remain orphaned in the local object store, on no branch.
- GitHub's Contributors graph already listed only `mertefesensoy`. That API counts the
  **default branch only**, so `main` being clean was sufficient for the graph. The
  contributor list never needed fixing, and **no history rewrite of `main` was
  required**.

**Eighteen `Author: Claude ...` bylines** in tracked documentation — ten implementation
docs, two ADRs, two hypotheses, two research briefs, and two older implementation docs.
These named an AI as the author of the project's own decision and hypothesis records.

**Roughly twenty references to files that are not in the repository.** `CLAUDE.md`,
`.claude/`, `docs/goal-prompt.md`, and `memory/MEMORY.md` are all gitignored or never
committed, yet were cited — some as live markdown links, including one on `README.md`
itself. Every such link 404s for an outside reader.

## 2. What Changed

| File | Change |
|---|---|
| `docs/zbridge-roadmap-2026-27.pdf` | Added to `main` from the stale branch, byte-identical (SHA-256 `3517296b…a0713`), so the branch could be retired. |
| 18 doc files under `docs/{decisions,hypotheses,implementations,research-briefs}/` | `Author:` / `Requested by:` bylines reattributed from Claude to Mert Efe Şensoy. |
| `docs/architecture/README.md` | Broken `CLAUDE.md` link dropped; ADR 0004 §3 now carries the no-cgo citation alone. |
| `docs/architecture/zbridge-module.md` | Three "hard rule N" citations reworded to name the rule inline. |
| `docs/decisions/0003-production-bridge-module-architecture.md` | `UNDEF`-convention citation reworded (spans two lines). |
| `docs/decisions/0004-roadmap-corrections-and-cgo-scope-closure.md` | Errata cross-reference no longer points at two untracked files. |
| `docs/decisions/0005-team-scale-up-and-academic-year-recalibration.md` | Doctrine-location sentence reworded. |
| `docs/team/charter.md` | `.claude/hooks/` and `.claude/agents/` described by function; standing-rules reading-list entry reworded. |
| `docs/implementations/2026-07-26-…`, `2026-07-30-…`, `2026-08-09-phase4-run-md.md`, `2026-07-25-research-returns-001-002.md` | Eight historical `CLAUDE.md` citations reworded to "the project standing rules" / "the build/test notes". |
| `docs/research-briefs/001-…`, `003-…` | Three "Claude" actor references reworded to the role names the briefs already use. |
| `README.md` | Broken `docs/goal-prompt.md` link removed; new **Tooling and authorship** disclosure section added. |
| `.claude/settings.json` *(untracked)* | `includeCoAuthoredBy: false` added alongside the existing empty `attribution` block. |

## 3. Implementation Approach

**The branch, not the history.** Because the trailer survived on exactly one commit, on
one stale branch, the fix was to retire the branch rather than rewrite anything. The
branch was 7,631 lines behind `main` and held exactly one artifact `main` lacked. That
artifact was extracted with `git show <ref>:<path>`, verified byte-identical by SHA-256
against the branch copy, committed to `main` under the owner's identity, and only then
was the branch deleted. **`main`'s history is untouched and no force-push was
performed** — the invariant that mattered, since `main` carries the v1.0.0 tag.

**Reattribute, do not erase.** Bylines were rewritten to the owner rather than deleted,
and a single disclosure section was added to `README.md`. The alternative — a silent
scrub — would have left `docs/team/charter.md` and ADR 0005 still describing a
research/architecture role split that the bylines no longer evidenced. An inconsistency
of that kind is exactly what a careful academic reader notices, and disclosed assistance
is the position that survives scrutiny; undisclosed-but-detectable is the worst of the
three available states.

**Name the rule, not the file.** Where an untracked file was cited as the *source of a
rule*, the rule is now named inline ("a standing project rule", "the project's
attribution rule"). Where it appeared in a *historical changelog entry*, the referent
was renamed ("project standing rules") rather than deleted, so the record of what was
edited in that session stays true. No claim about what happened was altered.

## 4. Mathematical / Statistical Details

*Omitted — purely structural.* The one numeric check is the SHA-256 equality above,
which establishes that retiring the branch lost no content.

## 5. Design Decisions

**Rejected: rewriting `main` with `filter-repo` or an interactive rebase.** `main` had
zero trailers, so this would have rewritten 25 commits and invalidated the `v1.0.0` tag
to fix nothing. The narrow diagnosis — one commit, one branch — is what made the cheap
fix correct, and is why the initial `main`-only scan had to be widened to `--all` before
acting.

**Rejected: deleting the branch outright.** It would have cost the rendered roadmap PDF,
which exists nowhere else on the remote. Moving the artifact first cost one commit.

**Rejected: committing `CLAUDE.md` and `docs/goal-prompt.md` to fix the broken links.**
Both are AI-operating instructions. Publishing them to repair link integrity would have
worked directly against the reason for this change.

**Rejected: deleting the byline lines outright.** Considered and available, but a
present, correct byline is more informative to a reviewer than an absent one, and the
git author metadata already agreed with it.

## 6. Verification

```bash
# 1. No AI attribution anywhere in tracked content
git grep -n -i -E "claude|anthropic" -- . ':!.gitignore'      # expect: no output

# 2. No co-author trailer on any commit reachable from a remote ref
git log --remotes --format="%B" | grep -ci "co-authored-by"   # expect: 0

# 3. Single contributor on GitHub
gh api "repos/mertefesensoy/zbridge-asm-lab/contributors?anon=1" \
  --jq '.[] | "\(.login // .name): \(.contributions)"'        # expect: mertefesensoy only

# 4. The retired branch is gone from the remote
git ls-remote --heads origin | grep roadmap-2026-27-pdf       # expect: no output

# 5. The PDF survived the retirement
git cat-file -e main:docs/zbridge-roadmap-2026-27.pdf && echo present

# 6. main's history was not rewritten
git merge-base --is-ancestor v1.0.0 main && echo "v1.0.0 still an ancestor of main"
```

Code paths are untouched by this change; the Go build and test gates are unaffected.

## 7. Related Docs

- [`README.md`](../../README.md) — the new **Tooling and authorship** section
- [`docs/team/charter.md`](../team/charter.md) — the role split this disclosure keeps honest
- [ADR 0005](../decisions/0005-team-scale-up-and-academic-year-recalibration.md) — records the same role split
- [`docs/roadmap-2026-27.md`](../roadmap-2026-27.md) — source of the PDF moved onto `main`

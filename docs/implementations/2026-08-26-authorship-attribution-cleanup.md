# Authorship attribution cleanup ahead of university outreach

**Date:** 2026-08-26
**Author:** Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

The repository is about to be shown to university contacts as the public face of the
thesis project. An AI `Co-Authored-By:` trailer was still visible on GitHub, and the
documentation still credited an AI assistant as the author of the project's own
decision records.

**The trailer survived on the `v1.0.0` release tag.** A previous cleanup had rewritten
`main`'s commits to strip AI co-author trailers, but **the tag was never moved with
them.** It still pointed at `2c7647b` — the *pre-amend* commit — and therefore at a
whole parallel history of six trailer-bearing commits that no branch referenced any
more. GitHub renders such a trailer as a second author avatar, so the Releases page and
the tag browser kept displaying it long after `main` was clean.

This is the general trap: **rewriting branch history does not move tags.** A tag is a
ref, so the old chain stays alive and published even though `git log` on every branch
looks clean.

Two things that appeared to be problems were not:

- **`main` was already clean** — 25 commits, all authored and committed solely by the
  owner, zero trailers.
- **GitHub's Contributors graph already listed only `mertefesensoy`.** That API counts
  the **default branch only**, so a clean `main` was sufficient for it. The contributor
  list never needed fixing and **`main`'s history was never rewritten here.**

A third suspect was ruled out mid-investigation and is recorded because the error is
instructive. The branch `docs/roadmap-2026-27-pdf` looked live and looked guilty: it
appeared in `git branch -a` as `remotes/origin/…` and its sole commit carried a trailer.
It had in fact already been deleted from GitHub; the local remote-tracking ref was stale,
and a plain `git fetch` does not prune such refs. **`git ls-remote` is the authority on
what exists remotely — `git branch -a` is not.**

Alongside this, two documentation problems: **eighteen AI `Author:` / `Requested by:`
bylines** across the ADRs, hypotheses, research briefs and implementation docs, and
**roughly twenty citations of files that are not in this repository** — the AI
instruction file, the AI tool directory, `docs/goal-prompt.md` and `memory/MEMORY.md`
are all gitignored or were never committed, yet were cited as live markdown links,
including one on `README.md` itself. Every such link 404s for an outside reader.

## 2. What Changed

| File / ref | Change |
|---|---|
| `refs/tags/v1.0.0` | Re-pointed from `2c7647b` to `4bf76cc`, the equivalent commit on `main`. Annotation message and tagger date preserved; force-pushed. **This was the actual fix.** |
| `docs/zbridge-roadmap-2026-27.pdf` | Added to `main` (`2bcfa5f`) from the retired branch, byte-identical (SHA-256 `3517296b…a0713`). |
| 18 doc files under `docs/{decisions,hypotheses,implementations,research-briefs}/` | `Author:` / `Requested by:` bylines reattributed to Mert Efe Şensoy. |
| `docs/architecture/README.md` | Broken AI-instruction-file link dropped; ADR 0004 §3 now carries the no-cgo citation alone. |
| `docs/architecture/zbridge-module.md` | Three "hard rule N" citations reworded to name the rule inline. |
| `docs/decisions/0003-production-bridge-module-architecture.md` | `UNDEF`-convention citation reworded (spans two lines). |
| `docs/decisions/0004-roadmap-corrections-and-cgo-scope-closure.md` | Errata cross-reference no longer points at two untracked files. |
| `docs/decisions/0005-team-scale-up-and-academic-year-recalibration.md` | Doctrine-location sentence reworded. |
| `docs/team/charter.md` | Local hooks and review agents described by function rather than by tool path; standing-rules reading-list entry reworded. |
| `docs/implementations/2026-07-26-…`, `2026-07-30-…`, `2026-08-09-phase4-run-md.md`, `2026-07-25-research-returns-001-002.md` | Eight historical citations reworded to "the project standing rules" / "the build/test notes". |
| `docs/research-briefs/001-…`, `003-…` | Three AI actor references reworded to the role names the briefs already use. |
| `README.md` | Broken `docs/goal-prompt.md` link removed; new **Tooling and authorship** disclosure section added. |
| AI tool settings *(untracked)* | `includeCoAuthoredBy: false` added alongside the existing empty `attribution` block, so the trailer cannot return. |

## 3. Implementation Approach

**Move the tag, not the history.** `main` was already clean, so the correct fix touched
exactly one ref. The tagged commit `2c7647b` and `main`'s `4bf76cc` were confirmed to
have **byte-identical trees** (`9c239f00cf17280ebfdafe2903bee478978cf38b`) before
anything was moved, which is what makes the re-point safe: the release content is
unchanged, only the commit object it hangs from differs. The annotated tag was recreated
with `git tag -f -a` using the original message and with `GIT_COMMITTER_DATE` set to the
original tagger date, then force-pushed. **`main`'s history was not rewritten and no
branch was force-pushed.**

The GitHub Release survived untouched, with all four assets (`gen-e3-linux-amd64`,
`gen-e3-linux-s390x`, `gen-e3-windows-amd64.exe`, `zbridge_bundle.zip`). Release assets
attach to the release object, which is keyed by tag *name*, not by the commit the tag
resolves to — so moving the tag does not disturb them. This was verified after the push,
not assumed.

**Order of operations.** The roadmap PDF existed only on the branch being retired, so it
was extracted, hash-verified, committed to `main`, and confirmed present on `origin/main`
*before* the branch was cleaned up — never the reverse.

**Reattribute, do not erase.** Bylines were rewritten to the owner rather than deleted,
and a single disclosure section was added to `README.md`. A silent scrub would have left
`docs/team/charter.md` and ADR 0005 still describing a research/architecture role split
that the bylines no longer evidenced. That kind of inconsistency is exactly what a
careful academic reader notices, and disclosed assistance is the position that survives
scrutiny; undisclosed-but-detectable is the worst of the three available states.

**Name the rule, not the file.** Where an untracked file was cited as the *source of a
rule*, the rule is now named inline ("a standing project rule", "the project's
attribution rule"). Where it appeared in a *historical changelog entry*, only the
referent was renamed, so the record of what was edited in that session stays true. No
claim about what happened was altered.

## 4. Mathematical / Statistical Details

*Omitted — purely structural.* The two numeric checks are content-identity proofs: the
SHA-256 equality on the roadmap PDF, and the git tree-hash equality between `2c7647b`
and `4bf76cc`. Together they establish that neither the branch retirement nor the tag
move changed any published content.

## 5. Design Decisions

**Rejected: rewriting `main` with `filter-repo` or an interactive rebase.** `main` had
zero trailers, so this would have rewritten 25 commits to fix nothing. The narrow
diagnosis — one tag — is what made the cheap fix correct.

**Rejected: deleting the release and re-creating it.** It would have reached the same end
state but required re-uploading roughly 12 MB of assets by hand, with a window in which
the release did not exist. Moving the tag preserved the release and its download history.

**Rejected: committing the AI instruction files to repair the broken links.** They are AI
operating instructions; publishing them to fix link integrity would work directly against
the reason for this change.

**Rejected: deleting the byline lines outright.** Available and considered, but a present,
correct byline tells a reviewer more than an absent one, and it agrees with the git author
metadata.

**Accepted limitation.** The six orphaned commits are no longer reachable from any ref,
but GitHub still serves unreachable commits by full SHA until it garbage-collects them;
this was confirmed empirically during the investigation. They appear in no branch, tag,
listing, or contributor count. Forcing their immediate removal requires a request to
GitHub Support, which was judged unnecessary — a 40-character SHA is not discoverable by
browsing.

## 6. Verification

```bash
# 1. No co-author trailer reachable from ANY remote ref - branches AND tags.
#    Note the explicit tag: `--remotes` alone would miss it, which is how
#    this defect survived the previous cleanup.
git fetch --prune --tags --force origin
git log origin/main v1.0.0 --format="%B" | grep -ci "co-authored-by"   # expect: 0

# 2. Every remote-reachable commit is authored by the owner
git log origin/main v1.0.0 --format="%an <%ae>" | sort -u              # expect: 1 line

# 3. The tag resolves to a commit that is an ancestor of main
git merge-base --is-ancestor v1.0.0^{} origin/main && echo ok

# 4. Contributors on GitHub
gh api "repos/mertefesensoy/zbridge-asm-lab/contributors?anon=1" \
  --jq '.[] | "\(.login // .name): \(.contributions)"'                 # expect: owner only

# 5. Release and all four assets intact
gh release view v1.0.0 --json assets --jq '.assets[] | .name'          # expect: 4 assets

# 6. The roadmap PDF survived the branch retirement
git cat-file -e origin/main:docs/zbridge-roadmap-2026-27.pdf && echo present

# 7. No stale branch remains on the remote (ls-remote, not `git branch -a`)
git ls-remote --heads origin                                          # expect: main only
```

All seven were run and passed on 2026-08-26. Code paths are untouched by this change;
the Go build and test gates are unaffected.

## 7. Related Docs

- [`README.md`](../../README.md) — the new **Tooling and authorship** section
- [`docs/team/charter.md`](../team/charter.md) — the role split this disclosure keeps honest
- [ADR 0005](../decisions/0005-team-scale-up-and-academic-year-recalibration.md) — records the same role split
- [`docs/roadmap-2026-27.md`](../roadmap-2026-27.md) — source of the PDF moved onto `main`

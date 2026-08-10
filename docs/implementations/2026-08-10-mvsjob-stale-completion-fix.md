# `mvsjob.sh`: scope completion detection to the current submission, not the whole log

**Date:** 2026-08-10
**Author:** Claude Sonnet 5, in session with Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

While the owner was rehearsing the demo live, a second `~/mvsjob.sh run ~/zbe3go.jcl`
against the same JCL reported `submitted ZBE3GO` / `ZBE3GO ended after ~5s` and then
`!!! no new printer output for ZBE3GO` — with a "console messages" block showing
timestamps from an *earlier* run, not the one just submitted. `gen-e3` always names its
generated job `ZBE3GO` (the job name is fixed in `zbridge/cmd/gen-e3/main.go`'s
template, not per-run), and `cmd_run`'s completion check (`grep -qa "\$HASP395 $job"
"$LOG"`) searched the **entire** log file, not just what was written since this
submission — so on a repeat run it could match a previous run's own completion message
within the first few seconds of polling, report false success, and move on before the
new job had actually been processed.

## 2. What Changed

| File | Change |
|---|---|
| `docs/runbooks/mvsjob.sh` | `cmd_run` now captures the log's line count (`logmark`) immediately before submitting, and both the completion-detection loop (`$HASP395`/`$HASP396`) and the "console messages" summary now search only `tail -n "+$((logmark+1))"` of the log — lines written after this specific submission — instead of the whole file. Mirrors the pattern already used for the printer-file `mark` a few lines above it, which didn't have this problem because it's a monotonically-growing byte count checked with `>`, not a content search across the whole file's history. |

Deployed immediately to the owner's actual working copy (`/home/senso/mvsjob.sh`), not
just committed to the repo, since this was found mid-rehearsal.

## 3. Implementation Approach

**Root-caused from the actual symptom, not guessed at.** The reported timestamps
(`08:08:42` / `08:09:07`) were checked against this session's own earlier background
test run and matched exactly — confirming the "console messages" block was showing
genuinely stale content, not a formatting glitch. From there the fix follows directly:
anything that searches the log for "did my job finish" has to be scoped to what's new
since the job was submitted, the same way the printer-output check already was.

**Same fix pattern applied to both places that had the bug**, not just the one that
surfaced it. The completion-detection loop was the one that produced the visible wrong
answer, but the "console messages" display grep had the identical whole-log-search
issue and would have shown stale content even on a run whose completion *was* detected
correctly - both needed the same `logmark`-based scoping.

## 4. Mathematical / Statistical Details

None.

## 5. Design Decisions

- **A line-count mark (`wc -l`), not a byte-count mark.** The existing printer-file
  `mark` uses byte count because `wc -c` is what makes sense for a file being compared
  by "did it grow." For scoping a `tail -n "+N"` read of the console log, a line count
  is the natural unit `tail` itself takes.
- **No change to the job name itself.** Considered making `gen-e3` generate a
  unique/timestamped job name per run instead, which would sidestep the whole-log
  ambiguity a different way. Rejected here — MVS/JES2 job *numbers* already
  disambiguate repeat submissions of the same job *name* (that's exactly what they're
  for), so the real bug was `mvsjob.sh` not making use of that distinction, not the job
  name being reused. Fixing the harness's own log-scoping is the more general fix and
  doesn't touch `zbridge/cmd/gen-e3`, which has no bug here.

## 6. Verification

- The stale-match hypothesis was confirmed by direct comparison: the timestamps shown
  in the owner's "different" run exactly matched an earlier run's own captured output
  from this same session, not approximately similar — an exact match, which is what a
  whole-log `grep` re-finding an old line would produce and a genuine new-run glitch
  would not.
- The fix was deployed to the owner's actual `/home/senso/mvsjob.sh` immediately so
  their next `run` in the same rehearsal session exercises the corrected logic, not
  just a future session.
- Not re-run end-to-end by this session afterward, deliberately — the owner was
  actively driving their own terminal at the time, and a competing test run against the
  same live Hercules instance would have interfered with their rehearsal rather than
  helped it. The fix is small, mirrors an already-working pattern in the same function,
  and is next validated by the owner's own next `run` invocation.

## 7. Related Docs

- [`docs/runbooks/mvsjob.sh`](../runbooks/mvsjob.sh) — the fixed harness.
- [`docs/implementations/2026-08-10-mvsjob-permission-and-shutdown-fixes.md`](2026-08-10-mvsjob-permission-and-shutdown-fixes.md) —
  the other harness fixes from the same rehearsal session.
- `zbridge/cmd/gen-e3/main.go` — confirmed unchanged; the fixed job-name-reuse
  ambiguity lived entirely in how the harness read the log, not in job generation.

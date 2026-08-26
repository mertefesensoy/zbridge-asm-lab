# `mvsjob.sh`: fix `/root`-only access and the `shutdown` script that doesn't shut down

**Date:** 2026-08-10
**Author:** Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

While rehearsing the live demo, the owner opened a normal WSL2 terminal (as their own
user, `senso`) and hit `Permission denied` running `mvsjob.sh` — the existing TK5
install and harness script lived under `/root`, readable only by `root`. Investigating
that live turned up a second, independent bug: `mvsjob.sh down`'s default shutdown path
(`script scripts/shutdown`) ran to completion twice without ever actually halting
Hercules — the system stayed up and kept doing routine work for 8+ minutes afterward,
confirmed by direct log inspection, not assumed. Both are fixed here, along with
relocating the owner's actual TK5 install so the fixes take effect for them immediately.

## 2. What Changed

| File | Change |
|---|---|
| `docs/runbooks/mvsjob.sh` | Three fixes: (1) `TK5_HOME`/`TK5_FIFO`/`TK5_OUT` now default to `$HOME/...` instead of `/root/...`. (2) `cmd_down` now sends `script scripts/poweroff` instead of `script scripts/shutdown` — the former completed cleanly in under 90 seconds both times it was tried tonight; the latter never completed either time. (3) `cmd_run`'s console-message filter now also matches a bare WTO line (`JOB +[0-9]+ +[+@*]`), which none of its previous patterns caught. |
| `RUN.md` | §4.3–§4.4, §6.1–§6.4, §7.2, §9: every `/root/...` path replaced with `~/...` (`$HOME`), with an explicit callout of *why* (`/root` needs `root`, most WSL2 users aren't root); shutdown timing corrected from an estimated "2–3 minutes" to the actually-measured "typically under 90 seconds"; added a troubleshooting entry for a shutdown that times out (check responsiveness via a console command, then just retry `down` — this is recoverable, not disk corruption). |
| (Operational, not a file change) | The owner's actual TK5 install and `mvsjob.sh` copy were moved from `/root/mvs-tk5` / `/root/mvsjob.sh` to `/home/senso/mvs-tk5` / `/home/senso/mvsjob.sh`, owned by `senso:senso`, after a confirmed clean shutdown. Re-verified end-to-end (`up` → `run` → `down`) running as plain `senso`, no `sudo`, no root elevation at all. |

No source or test files changed (`.go` files untouched — this is entirely the emulation
harness).

## 3. Implementation Approach

**Diagnosed from direct log evidence at every step, not assumption.** The permission
issue was confirmed by reproducing it (`ls -la /root/mvs-tk5` as `senso`: `Permission
denied`) before concluding anything. The shutdown issue was confirmed by watching the
log directly across multiple real attempts: the first `down` succeeded cleanly in ~135s;
a second attempt, and a same-mechanism retry, both ran the `scripts/shutdown` sequence
to its own logged completion (`Script N: file scripts/shutdown processing ended`) and
then the system kept doing unrelated routine work (printer report cycling) for 8+
minutes with zero occurrences of `HHC01412I` anywhere in the log. Only after that was
`scripts/poweroff` tried as an alternative — named in this project's own earlier
evidence (`docs/evidence/E0-tk5-boot-2026-07-26.md`: *"Also `quiesce`, `poweroff`,
`z_eod`"*) but never previously exercised — and it completed cleanly, confirmed by
`HHC01412I Hercules terminated` plus the full termination log sequence, on both attempts.

**Fixed the default rather than documenting a workaround.** `$HOME` instead of a
hardcoded path is correct for *any* user running this script, not just this one
machine's specific fix — a future team member on their own WSL2 setup gets the correct,
permission-safe default automatically, with no documentation to remember.

**The relocation was performed only after a confirmed-clean shutdown**, consistent with
this project's own standing rule that touching the DASD tree while any Hercules
instance might be live (or might have stopped uncleanly) makes the resulting state
untrustworthy. The move itself (`mv`, then `chown -R`) was verified by re-running the
full cycle afterward as the now-correct owning user, not assumed to have worked.

## 4. Mathematical / Statistical Details

None.

## 5. Design Decisions

- **`poweroff` becomes the default, `shutdown`'s failure is recorded in a comment, not
  silently dropped.** Considered keeping `shutdown` as the default and adding `poweroff`
  as a documented fallback. Rejected — `shutdown` failed 2 for 2 in direct testing
  tonight, `poweroff` succeeded 2 for 2. A harness's default path should be the one with
  evidence behind it. The comment in `mvsjob.sh` explains why, so a future reader isn't
  left wondering why the "obviously named" option isn't the one used.
- **`$HOME`, not a hardcoded `/home/senso`.** The owner's own machine now has the install
  under `/home/senso`, but the script must not hardcode that — `docs/team/charter.md`
  describes four more contributors who will each run this on their own account. `$HOME`
  resolves correctly for all of them with no per-person edit.
- **Cleaned up the stale `/root`-side scratch files from tonight's troubleshooting**
  (`wsl_up.sh` and similar) as part of the move, rather than leaving two copies of
  similar-but-different scripts lying around to confuse a future session.

## 6. Verification

- Full `up` → `run` → `down` cycle re-run, live, three separate times tonight against
  the same physical TK5 install: once against the original `/root` location (before the
  permission problem was even noticed, confirming the base mechanism worked), once
  immediately after the `poweroff` fix (confirming the shutdown fix specifically), and
  once more from the final `/home/senso` location running as plain `senso` with zero
  elevation (confirming both fixes together, in the actual environment the owner will
  present from).
- Every claim in this document — the permission denial, the `scripts/shutdown` timeout,
  the `HHC01412I` absence, the `poweroff` success — was read directly from command
  output or log content during this session, not inferred.
- `RUN.md`'s corrected shutdown timing ("typically under 90 seconds") is the actually
  observed figure from the `poweroff` runs, not an estimate.

## 7. Related Docs

- [`docs/runbooks/mvsjob.sh`](../runbooks/mvsjob.sh) — the fixed harness.
- [`RUN.md`](../../RUN.md) §4, §6, §7, §9 — updated to match.
- [`docs/evidence/E0-tk5-boot-2026-07-26.md`](../evidence/E0-tk5-boot-2026-07-26.md) —
  source of the `quiesce`/`poweroff`/`z_eod` alternatives that pointed at tonight's fix.
- [`docs/architecture/c4/sequence-state-object-diagrams.md`](../architecture/c4/sequence-state-object-diagrams.md) §2 —
  the state diagram this session's "Corrupted" branch discussion was grounded in.

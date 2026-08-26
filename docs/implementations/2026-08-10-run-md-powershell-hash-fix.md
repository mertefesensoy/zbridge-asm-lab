# `RUN.md` §2: add the PowerShell hash-verification form

**Date:** 2026-08-10
**Author:** Mert Efe Şensoy
**Status:** shipped

---

## 1. Problem / Motivation

While following `RUN.md` for a pre-flight check, the owner ran the TK5 hash-verification
snippet in PowerShell and hit `sha256sum: The term 'sha256sum' is not recognized...` —
`RUN.md` §2 only gave the Unix form of this command, breaking the convention the rest of
the document follows (a PowerShell block alongside every bash block for any step that
crosses the Windows/WSL2 boundary). The download itself happens inside WSL2 in §4.3, but
§2 presents the hash check earlier and in isolation, without saying so, which is exactly
what invited running it on the Windows side instead.

## 2. What Changed

| File | Change |
|---|---|
| `RUN.md` | §2: the TK5 hash-verification snippet now states explicitly that §4.3 downloads inside WSL2 (so `sha256sum` is the one actually used), and adds a PowerShell form (`Get-FileHash`) for anyone who downloaded it on the Windows side instead. |

No source or test files changed. This was found and fixed live, in the same session,
while separately re-extracting the owner's actual TK5 install after its shutdown state
could not be verified (see the session's conversation — that was an operational action
on the owner's WSL2 environment, not a documentation change, so it has no corresponding
implementation doc).

## 3. Implementation Approach

Matched the fix to the existing convention rather than inventing a new pattern — every
other Windows/WSL2-crossing step in `RUN.md` already gives both a `powershell` and a
`bash` fenced block; this was the one place that convention was silently dropped.

## 4. Mathematical / Statistical Details

None.

## 5. Design Decisions

- **Named which one is actually used, not just given two options.** A bare "here's the
  PowerShell version, here's the bash version" would have left the reader to guess which
  one applies to their situation. Since §4.3 downloads inside WSL2, the fix says so
  directly and frames the PowerShell form as the fallback for someone who downloaded on
  the Windows side, not as an equally-likely default.

## 6. Verification

- Re-read the edited section to confirm both hash values are byte-identical to the one
  already recorded in `docs/evidence/E0-tk5-boot-2026-07-26.md` and used everywhere else
  in `RUN.md`.
- The underlying trigger was reproduced directly: the owner's actual PowerShell error
  message is what prompted this fix, not a hypothetical.

## 7. Related Docs

- [`RUN.md`](../../RUN.md) §2 — the changed section.
- [`docs/implementations/2026-08-09-phase4-run-md.md`](2026-08-09-phase4-run-md.md) —
  original `RUN.md` implementation doc.

# Return: Research Brief 002

## Q1 — Direct hit
**[FOUND] (High Confidence)**

**Verdict: UPHELD**
No public Go-assembly WTO implementation was found.

- **Search Surface:** GitHub, GitLab, pkg.go.dev, the Go issue tracker, IBM's `ibmruntimes` organization, and general web search.
- **Query terms:** `"SVC 35" golang`, `"WTO" "go assembly" z/os`, `z/OS console golang`, `GOARCH=s390x .s WTO`.
- **Date:** 2026-07-25.

The only matches for "SVC 35" in a Go context point to ARM/TinyGo Supervisor Calls (embedded systems interrupt handling) or the Windows Service package `golang.org/x/sys/windows/svc` log outputs (e.g., line number 35).

## Q2 — Adjacent Go-on-z/OS assembly work
**[FOUND] (High Confidence)**

- **`ibmruntimes/go-recordio`**: Uses Go assembly (`utils.s`) to perform VSAM Record I/O. It implements functions like `IefssreqX` to bridge the AMODE-31/AMODE-64 gap and memory management routines (`Malloc31`) for below-the-bar storage.
- No other substantial public repositories were found demonstrating direct SVC (e.g. ENQ/DEQ, STORAGE) or PC-routine calls from Go assembly on z/OS.

## Q3 — Cross-language comparators
**[FOUND] (High Confidence)**

| Name | Language | Implements | Direct SVC or C Shim | License | Last Activity | URL |
|---|---|---|---|---|---|---|
| **Ambitus `pyzkiln`** | Python | z/OS automation, WTO | C Shim / LE runtime | Apache 2.0 | Active | [github.com/ambitus/pyzkiln](https://github.com/ambitus/pyzkiln) |
| **JZOS** | Java | z/OS system services | C Shim (JNI) | Commercial / IBM | Active | IBM Documentation |

**Direct SVC call:** None of the comparators bypass the C/LE runtime layer to issue raw `SVC 35` directly from the high-level language binding. They route through `__console2()` or similar C/LE runtime facilities.

## Q4 — The published-walkthrough gap
**[FOUND] (High Confidence)**

**Verdict: UPHELD**
No published, instruction-by-instruction annotation of `ibmruntimes/go-recordio`'s `utils/utils.s` exists. The repository provides the code, but an educational walkthrough dissecting the stack layout, AMODE switching, and MVS macro expansion in Go assembly was not found anywhere on the public internet.
- **Date checked:** 2026-07-25.

## Q5 — Currency
- **`ibmruntimes/go-recordio`**: Actively maintained by IBM (BSD-3-Clause).
- **`pyzkiln`**: Active community maintenance.

## Re-check date
These negative existence claims should be re-verified before the public release of the module or thesis, specifically around **October 2026** or immediately prior to artifact submission.

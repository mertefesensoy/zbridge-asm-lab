---
rung:          E1, E2, E3 (all three)
date:          2026-07-26
machine:       Hercules
guest_os:      MVS 3.8j (TK5 Update 5, released 2026-02-18)
architecture:  S/370 (24-bit)
emulator:      Hercules 4.9.1.0-SDL (SDL Hyperion), built 2025-12-08, the Linux x86-64 binary bundled with TK5
host:          Windows 11 Home 10.0.26200 → WSL2 Ubuntu 26.04 LTS; Intel Core i7-13650HX, 15.6 GB RAM
speaks_to:     U2
hypothesis:    H001
verdict:       PASS (E1, E2, E3)
---

# E1–E3 · The WTO parameter list layout, and Go-produced bytes accepted by a real SVC 35

**This is the central result of the project so far.** Three rungs passed in one
session, and between them they retire **U2**.

**What this is not.** MVS 3.8j is z/OS's 1981 ancestor running under emulation. Nothing
here is a z/OS result (ADR 0001 §7). What z/OS does with the same bytes is H001's
question and is still open. No performance figure was taken.

---

## 1. Rung E1 · The layout, from IBM's own macro

### Why the macro and not a manual

Research brief 003 changed the plan. Q1 established that the WPL byte layout **is not
documented in prose in the MVS 3.8j supervisor manual at all** — it is carried only by
the `IEZWPL` mapping macro. Q6 established that **no IBM publication prints the
generated expansion** of the WTO macro; users are expected to assemble with `PRINT GEN`
and read the listing.

That reframes two days of blocked work. The layout was never going to be found in
GC28-0683-2, which is exactly what the DOC-001 audit discovered the hard way. The
authoritative instrument is an assembler listing, and TK5 can produce one.

### Audit finding against return 003 Q1

Q1 said `IEZWPL` ships "typically" in `SYS1.MACLIB` or `SYS1.MODGEN`. Job `ZBE1MAC`
tested that directly:

```
PRINT/PUNCH DATA SET UTILITY
  PRINT TYPORG=PO,MAXNAME=1,MAXFLDS=1
  MEMBER NAME=IEZWPL
IEB408I MEMBER IEZWPL    CANNOT BE FOUND
IEF142I ZBE1MAC PRTMAC - STEP WAS EXECUTED - COND CODE 0008
```

**`IEZWPL` is not in `SYS1.MACLIB` on TK5.** The "typically" was doing real work in that
sentence. This does not damage Q1's substance — the layout still is not in prose, and
the macro-expansion route still works — but it is the second Gemini return in a row
whose library/page detail did not survive checking. The lesson stands: *verify the form
number, then verify the page.*

### The expansion — the primary source

Job `ZBE1ASM`, assembled by **IFOX00 (Assembler XF)**, `RC= 0000`, with
`PRINT ON,GEN,DATA`. `DATA` matters: without it the assembler truncates constants in the
listing and the text bytes would be cut off.

```
  LOC  OBJECT CODE      STMT   SOURCE STATEMENT
000000                     7 WPLMIN   WTO   'ZBRIDGE TEST HELLO',MF=L
000000                     8+WPLMIN   DC    0F'0'
000000 0016                9+         DC    AL2(22)                 TEXT LENGTH
000002 0000               10+         DC    B'0000000000000000'     MCS FLAGS
000004 E9C2D9C9C4C7C540   11+         DC    C'ZBRIDGE TEST HELLO'   MESSAGE TEXT
00000C E3C5E2E340C8C5D3
000014 D3D6
```

**The layout:**

```
 byte:  0     1     2     3     4                              len+4
       +-----------+-----------+--------------------------------+
       |  length   | MCS flags |     message text (EBCDIC)      |
       +-----------+-----------+--------------------------------+
        big-endian halfword,
        value = len(text) + 4        DC 0F'0' → fullword aligned
```

- **Bytes 0–1**: big-endian halfword, `len(text) + 4`.
- **Bytes 2–3**: MCS flags. **Zero** for a minimal single-line WTO.
- **Bytes 4+**: message text in EBCDIC.
- **`DC 0F'0'`**: the list is **fullword-aligned**.

**The `+4` is confirmed at two lengths**, which is what makes it a rule rather than a
coincidence: 18 characters gave `0x0016` (22), and 38 characters gave `0x002A` (42).

And the reason is now visible rather than inferred: the field counts its own 2 bytes,
plus the 2 flag bytes, plus the text. **It is a length of header-plus-text, not a length
of text.** Every previous description of this as "text length + 4" was numerically right
and explanatorily empty.

### The routed form, and the trap inside it

The same listing revealed the form with routing and descriptor codes:

```
000018 0016   DC AL2(22)                TEXT LENGTH      ← still 22
00001A 8000   DC B'1000000000000000'    MCS FLAGS        ← bit 0 set
00001C ...    DC C'ZBRIDGE TEST HELLO'  MESSAGE TEXT
00002E 0200   DC B'0000001000000000'    DESCRIPTOR CODES
000030 0020   DC B'0000000000100000'    ROUTING CODES
```

Three findings:

1. **MCS flag bit 0 (`0x8000`) signals that descriptor and routing halfwords follow the
   text.** This is a *partial answer to Q3*, which the return marked NOT FOUND — one bit
   meaning is now attested from primary source on System A.
2. **The length field still reads 22 even though four more bytes follow.** So the length
   field does **not** cover the descriptor and routing halfwords. An implementation
   assuming "length = whole parameter list" would be wrong in a way that only appears
   when routing codes are used. This is the single most valuable trap the listing exposed.
3. **The bit conventions match this project's existing code exactly.** `ROUTCDE=(11)`
   generated `0x0020` and `DESC=(7)` generated `0x0200` — precisely what
   `console/options.go`'s `routeMask`/`descriptorMask` already computed from the
   documented 1-based, MSB-first convention. Written before the evidence, confirmed by it.

### The E1 gate: the macro reaches the console

Job `ZBE1WT2`, all three steps `RC= 0000` (`IFOX00`, `IEWL`, `GO`):

```
FFFF 15.28.50 JOB    5  +ZBRIDGE TEST E1 WTO REACHED THE CONSOLE
```

**Note the leading `+`.** That is **ADR 0001 evidence item 7 confirmed by direct
observation**: an unauthorized, problem-state WTO is console-*prefixed*, not blocked.
The roadmap's premise that problem-state WTO is functionally complete and only
cosmetically marked now has an observation behind it rather than a citation.

Captured from `log/hardcopy.log` (device `030E`, a 1403 configured as the hardcopy log)
— a file, not a screenshot.

---

## 2. Rung E2 · Hand-built parameter list, raw SVC 35, no macro

ADR 0001 §6 defines E2 as the rung that retires **U2**, because it proves the layout is
*understood* rather than that a macro can be invoked.

```
ZBE2SVC  CSECT
         SAVE  (14,12)
         LR    12,15
         USING ZBE2SVC,12
         LA    1,WPL          ← MVS linkage: R1 carries the parameter list
         SVC   35             ← WTO, raw
         RETURN (14,12),RC=0
         DS    0F
WPL      DC    AL2(MSGEND-WPL)
         DC    XL2'0000'
MSGTEXT  DC    C'ZBRIDGE TEST E2 RAW SVC 35 NO MACRO'
MSGEND   EQU   *
```

`AL2(MSGEND-WPL)` lets the assembler compute the length field. That is deliberate: it
makes the `+4` self-evident, because `MSGEND-WPL` *is* the whole list, and it removes
hand arithmetic as a source of error.

**Result — no `WTO` macro anywhere in the program:**

```
FFFF 15.41.36 JOB    6  +ZBRIDGE TEST E2 RAW SVC 35 NO MACRO
```

`IFOX00 RC= 0004` (one flagged statement, a warning), `IEWL RC= 0000`, `GO RC= 0000`.

---

## 3. Rung E3 · Bytes built by Go, accepted by a real SVC 35

This is the rung ADR 0001 §6 called *"the reason this program is worth running."* Go
cannot execute on MVS 3.8j — but its **bytes** can cross.

The parameter list was produced on the Windows host by
`zbridge/console.EncodeWPL`, rendered into assembler constants by
`console.FormatDC`, and emitted as a complete job by `zbridge/cmd/gen-e3`:

```
message      : "ZBRIDGE TEST E3 BYTES BUILT BY GO" (33 chars)
WPL bytes    : 00 25 00 00 E9 C2 D9 C9 C4 C7 C5 40 E3 C5 E2 E3 40 C5 F3 40
               C2 E8 E3 C5 E2 40 C2 E4 C9 D3 E3 40 C2 E8 40 C7 D6
length field : 37 = len(text) 33 + 4
MCS flags    : 00 00
```

```
ZBE3GO   CSECT
         ...
         LA    1,WPL
         SVC   35
         DS    0F
WPL      DC    X'00250000E9C2D9C9C4C7C540E3C5E2E3'
         DC    X'40C5F340C2E8E3C5E240C2E4C9D3E340'
         DC    X'C2E840C7D6'
```

**No human wrote those bytes and IBM's macro was not involved.** Result:

```
FFFF 15.51.08 JOB    1  +ZBRIDGE TEST E3 BYTES BUILT BY GO
```

`IFOX00 RC= 0000`, `IEWL RC= 0000`, `GO RC= 0000`.

**The Go-side construction of the WTO parameter list is verified against a genuine
`SVC 35` implementation, without Go ever running on a mainframe.**

### The laptop-side half of the same claim

`zbridge/console/wpl_oracle_test.go` asserts byte-for-byte equality between
`EncodeWPL`'s output and the bytes IBM's macro generated:

```
--- PASS: TestEncodeWPLMatchesIBMMacro
--- PASS: TestEncodeWPLMatchesIBMMacroLongText
--- PASS: TestLengthFieldIsTextPlusFour
--- PASS: TestMCSFlagsAreZeroForMinimalForm
--- PASS: TestRoutingIsRefusedNotDropped
--- PASS: TestFormatDCProducesTheMacroBytes
```

That test's expected values are not a specification restated as code — they are what a
real IBM macro emitted on a real system. The oracle runs on any host, so the E3
property is now regression-tested on every `go test`.

---

## 4. What Phase 3b steps these rungs retire

ADR 0001 §6 predicted four of six, "plus half of a fifth". Measured against what
actually happened:

| Phase 3b step | Retired? | By what |
|---|---|---|
| 1. Allocate the buffer below the bar via `Malloc31` | ❌ | U3 — z/OS only, unchanged |
| 2. Translate UTF-8 → EBCDIC IBM-1047 via `AtoE` | ✅ | E3: the EBCDIC in the accepted list is `codepage.AtoE` output |
| 3. Construct the WTO parameter list | ✅ | E1 gave the layout; E3 proved Go's construction is accepted |
| 4. Load R1 with the parameter-list address | ⚠️ partial | E2/E3 verified the linkage; AMODE context is U3 |
| 5. Issue `SVC 35` | ✅ | E2 and E3 both issued it raw |
| 6. Read R15, map to a Go error | ❌ **not retirable here** | See below |

**Step 6 is the one ADR 0001 got wrong, and return 003 Q4 explains why.** Q4 documents a
divergence: MVS 3.8j issues **no return code** for a single-line non-MLWTO WTO (R1
returns the message identification number instead), while z/OS **does** return one in
R15. So E3 *cannot* retire step 6 — the ancestor has nothing to read. The doubt recorded
against ADR 0001 §6 on 2026-07-25 is now resolved: **the doubt was correct.**

This is also why `zbridge.Error` carries `HasCode bool`. That field was added on
2026-07-25 from the GC28-0683-2 p.210 reading alone; return 003 Q4 independently
confirms the divergence, and the design decision holds.

---

## 5. H001 sub-claims

| Sub-claim | Status after these rungs |
|---|---|
| **C1** — MVS 3.8j `SVC 35` accepts a hand-built parameter list | **CONFIRMED.** E2 and E3 both reached the console with no macro |
| **C2** — the WPL layout is as understood | **CONFIRMED for the minimal single-line form on System A**, from the macro expansion. The routed form is observed but its MCS bit assignments are uncited for z/OS (brief 003 Q3) |
| **C3** — R1 in, R15 out | **R1-in CONFIRMED** (E2/E3 load R1 and the service accepts it). **R15-out remains contradicted on the MVS side** and is untested here |
| **C4** — the return-code contract | **RESOLVED as a divergence**, per return 003 Q4: no return code on MVS, a return code on z/OS. Not a failure of the oracle — a documented difference to design around |

H001 stays **open** overall, because every z/OS-side statement is still documentary. But
its MVS-side half is now evidenced rather than assumed.

---

## 6. Operational record — including what went wrong

Five defects were hit and are recorded because the fixes are now rules.

1. **`IEB408I MEMBER IEZWPL CANNOT BE FOUND`** — the macro is not in `SYS1.MACLIB`.
   Audit finding against return 003 Q1.
2. **`IEWL RC= 0016`, `WRNG.LEN.RECORD` on SYSLIN** — overriding `PARM.ASM='LIST,NODECK'`
   on the `ASMFCLG` proc removed the proc's own `PARM=OBJ`, so no object deck was
   produced for the linker. **Use `PARM.ASM='OBJ,LIST'` or do not override.**
3. **`IEF642I EXCESSIVE PARAMETER LENGTH ON THE JOB STATEMENT`** — the programmer-name
   field on a JOB card is limited to **20 characters** on MVS 3.8j; a 22-character name
   caused `IEF452I JOB NOT RUN - JCL ERROR`. Now documented in `cmd/gen-e3`.
4. **Printer drain is asynchronous.** `$HASP395` means the job ended, not that its output
   is written. The harness polls until `prt/prt00e.txt` stops growing.
5. **`pkill -9 -f 'hercules -f conf'` killed its own shell**, because the pattern matched
   the invoking command string. Use `pkill -x hercules`.

**Clean shutdown could not be verified for these runs, and that is stated rather than
glossed.** `HHC01412I Hercules terminated` did not appear: WSL2 tears its VM down once
no session is attached, so Hercules dies on stdin EOF when a tool invocation ends. Only
a single long-running process can drive MVS's shutdown to completion — E0 achieved it
that way. Consequences accepted here because **every artifact was written to files
before shutdown, and the DASD is disposable**: `mvs-tk5.zip` and its SHA-256 are
recorded, so recovery is a three-minute re-extract, which is what was done between runs.

The harness that encodes all of this is committed at `docs/runbooks/mvsjob.sh`.

---

## 7. Reproduction

```bash
# Layout and the E1 gate
mvsjob.sh up
mvsjob.sh run zbe1asm.jcl     # PRINT ON,GEN,DATA expansion of WTO ...,MF=L
mvsjob.sh run zbe1wto.jcl     # WTO macro to the console
# E2: hand-built list, raw SVC 35
mvsjob.sh run zbe2svc.jcl
mvsjob.sh down
```

```bash
# E3: bytes built by Go
cd zbridge && go run ./cmd/gen-e3 > zbe3go.jcl
mvsjob.sh up && mvsjob.sh run zbe3go.jcl && mvsjob.sh down
grep 'ZBRIDGE TEST' mvs-tk5/log/hardcopy.log
```

Raw artifacts: `docs/evidence/E1-E3-artifacts/` — every job's full printer listing, the
submitted JCL, and the console hardcopy lines.

---

## 8. What is still open

- **Everything z/OS-side.** The layout is verified on the *ancestor*. Brief 003 Q2 gives
  z/OS the same `+4` rule cited to the Authorized Assembler Services Reference, which is
  why the encoder was allowed to ship, but no z/OS system has run this.
- **Brief 003 Q3** — the full MCS flag bit table. One bit (`0x8000`) is now attested on
  System A. `WithRoute`/`WithDescriptor` still return `ErrLayoutUnverified`.
- **Brief 003 Q5** — out-of-range and abend behaviour. Untested; nothing here probed a
  bad length deliberately.
- **U3 entirely** — `Malloc31`, AMODE 31↔64, `GOOS=zos`. No emulator reaches it.
- **The `+` prefix on z/OS.** Observed on MVS; whether z/OS marks an unauthorized WTO
  the same way is unverified.

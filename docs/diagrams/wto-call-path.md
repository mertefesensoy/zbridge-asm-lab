# Diagrams · the WTO call path, the unknowns, and the ladders

Diagrams-as-code, version-controlled next to the evidence they explain. The interactive
companion is [`docs/interactive/wto-explainer.html`](../interactive/wto-explainer.html),
which carries the clickable byte layout and the self-test questions.

Every value in these diagrams came off a real system. Provenance:
`docs/evidence/E1-E3-wto-layout-and-svc35-2026-07-26.md` (MVS 3.8j, TK5 Update 5,
Hercules 4.9.1.0-SDL) and `docs/evidence/E-L-s390x-port-qemu-2026-07-25.md`
(linux/s390x under QEMU 10.2.1). **None of it is a z/OS result.**

---

## 1. The three unknowns

The decomposition that turned "we have no mainframe" into a work plan. Knowing which
unknown a piece of work retires is the difference between progress and motion.

```mermaid
flowchart TB
    Q["WTO(message string) error<br/>pure Go assembly, no cgo"]

    Q --> U1["<b>U1</b><br/>Does our Go assembly emit correct s390x?<br/>Does the Go ABI hold on big-endian 64-bit?"]
    Q --> U2["<b>U2</b><br/>Is the WTO parameter list byte-correct?<br/>Does SVC 35 accept it?"]
    Q --> U3["<b>U3</b><br/>GOOS=zos toolchain, Malloc31 below the bar,<br/>AMODE 31 to 64, USS"]

    U1 --> O1["QEMU s390x<br/><i>laptop, no licence</i>"]
    U2 --> O2["MVS 3.8j under Hercules<br/><i>the only oracle that exists</i>"]
    U3 --> O3["Real entitled z/OS<br/><i>nothing emulates this</i>"]

    O1 --> R1["RETIRED 2026-07-25<br/>29 tests pass on linux/s390x"]
    O2 --> R2["RETIRED 2026-07-26<br/>rungs E1, E2, E3"]
    O3 --> R3["OPEN<br/>blocked on access AND on IBM's Go fork"]

    classDef done fill:#2F6B3A,color:#fff,stroke:#1d4526
    classDef open fill:#8A3324,color:#fff,stroke:#5c2118
    classDef q fill:#0F4C8C,color:#fff,stroke:#0a3461
    class R1,R2 done
    class R3 open
    class Q q
```

**The point to make out loud:** two of the three unknowns fell without a mainframe. U3 is
the only one that needs hardware, and it splits further — a licence gets you a system, but
it does not get you a compiler.

---

## 2. The ladders

The E-ladder runs off-mainframe and precedes the T-ladder, because mainframe time is the
project's scarcest resource. Every rung produces committed evidence.

```mermaid
flowchart LR
    subgraph E["E-ladder · off-mainframe · ALL PASSED 2026-07-26"]
        direction TB
        E0["<b>E0</b> TK5 IPLs, console readable,<br/>job in and output back"]
        E1["<b>E1</b> WTO macro to the console;<br/>PRINT GEN gives the byte layout"]
        E2["<b>E2</b> hand-built list + raw SVC 35,<br/>no macro anywhere"]
        E3["<b>E3</b> Go-produced bytes<br/>accepted by a real SVC 35"]
        E0 --> E1 --> E2 --> E3
    end

    subgraph T["T-ladder · real hardware · NOT STARTED"]
        direction TB
        T0["<b>T0</b> pure-Go binary runs"]
        T1["<b>T1</b> add passes"]
        T2["<b>T2</b> five exercises with real _s390x.s"]
        T3["<b>T3</b> WTO on the z/OS console"]
        T0 --> T1 --> T2 --> T3
    end

    E3 -->|"retires 4 of 6<br/>Phase 3b steps"| T

    classDef done fill:#2F6B3A,color:#fff,stroke:#1d4526
    classDef todo fill:#A8B29E,color:#1C2117,stroke:#7d8874
    class E0,E1,E2,E3 done
    class T0,T1,T2,T3 todo
```

---

## 3. The WTO parameter list

Read out of the IFOX00 expansion of `WTO 'ZBRIDGE TEST HELLO',MF=L`. The layout is not
documented in prose in any IBM manual — the assembler listing **is** the primary source.

```mermaid
packet-beta
    0-1: "length = len(text) + 4"
    2-3: "MCS flags"
    4-21: "message text, EBCDIC IBM-1047"
```

Real bytes, minimal form (22 bytes):

```
00 16 | 00 00 | E9 C2 D9 C9 C4 C7 C5 40 E3 C5 E2 E3 40 C8 C5 D3 D3 D6
 len  | flags | Z  B  R  I  D  G  E  ␣  T  E  S  T  ␣  H  E  L  L  O
```

**Why `+4`, stated so it can be defended:** the field counts its own two bytes, plus the
two flag bytes, plus the text. It is a length of *header plus text*, not of text.
Confirmed at two lengths — 18 chars gave `0x0016` (22), 38 chars gave `0x002A` (42).

### The routed form, and the trap in it

```
00 16 | 80 00 | ...18 text bytes... | 02 00 | 00 20
 len  | flags |                     | desc  | route
  ↑        ↑
  |        └─ bit 0 set: descriptor and routing halfwords follow the text
  └────────── STILL 22, even though the list is now 26 bytes
```

The length field never covers the descriptor and routing halfwords. An implementation
assuming "length = whole parameter list" looks correct until someone passes a routing
code. This is why `zbridge/console` refuses `WithRoute`/`WithDescriptor` rather than
guessing: brief 003 Q3 found no citable table for the other fifteen MCS bits.

---

## 4. The call path, and where each piece was proven

```mermaid
flowchart TB
    A["Go caller<br/>console.WTO(msg)"] --> B["validate<br/>1..124 bytes, no control chars"]
    B --> C["codepage.AtoE<br/>UTF-8 to EBCDIC IBM-1047"]
    C --> D["EncodeWPL<br/>len+4 · flags · text"]
    D --> E["storage.Malloc31<br/>fullword-aligned, below the 2 GB bar"]
    E --> F["load R1 with the list address<br/>MVS linkage"]
    F --> G["SVC 35<br/>hand-encoded: BYTE 0x0A, BYTE 0x23"]
    G --> H["read the result"]
    H --> I["map to a Go error<br/>Error.HasCode"]

    C -.->|"proven E3<br/>+ QEMU"| P1[" "]
    D -.->|"proven E1 layout<br/>E3 bytes accepted"| P2[" "]
    E -.->|"U3 — no emulator"| P3[" "]
    F -.->|"proven E2, E3"| P4[" "]
    G -.->|"proven E2, E3"| P5[" "]
    I -.->|"MVS returns NOTHING;<br/>z/OS returns R15"| P6[" "]

    classDef done fill:#2F6B3A,color:#fff,stroke:#1d4526
    classDef open fill:#8A3324,color:#fff,stroke:#5c2118
    classDef ghost fill:none,stroke:none,color:#4A5343
    class C,D,F,G done
    class E,I open
    class P1,P2,P3,P4,P5,P6 ghost
```

Two steps remain open, and for different reasons. `Malloc31` is U3 — no emulator reaches
below-the-bar allocation or AMODE switching. Reading a return code is open because of a
**documented divergence**: MVS 3.8j issues none for a single-line WTO (R1 carries the
message ID instead), while z/OS returns one in R15.

---

## 5. The register collision

The single most dangerous detail in the project. Two calling conventions want the same
two registers for different things.

```mermaid
flowchart LR
    subgraph MVS["Standard MVS linkage"]
        direction TB
        M1["R1 · parameter list address"]
        M13["R13 · 72-byte save area"]
        M14["R14 · return address"]
        M15["R15 · entry point in, return code out"]
    end

    subgraph GO["Go ABI on s390x<br/>cmd/internal/obj/s390x/a.out.go"]
        direction TB
        G1["R1 · general use"]
        G13["R13 · REGG — the goroutine pointer g"]
        G14["R14 · link register"]
        G15["R15 · REGSP — the stack pointer"]
    end

    M13 ---|"COLLISION"| G13
    M15 ---|"COLLISION"| G15
    M14 ---|"agrees"| G14

    classDef bad fill:#8A3324,color:#fff,stroke:#5c2118
    classDef ok fill:#2F6B3A,color:#fff,stroke:#1d4526
    class M13,M15,G13,G15 bad
    class M14,G14 ok
```

Clobber R13 and the Go runtime cannot find the current goroutine. Clobber R15 and there
is no stack. The `regs/` exercise exists to make both observable from a Go test — and its
s390x tests prove R13 really holds `g` by showing that eight goroutines read eight
distinct values.

---

## 6. Why s390x, in one comparison

```mermaid
flowchart LR
    subgraph A["amd64 · seven instructions per byte"]
        direction TB
        A1["MOVBQZX (SI), AX"] --> A2["MOVB (BX)(AX*1), DL"] --> A3["MOVB DL, (DI)"]
        A3 --> A4["INCQ SI / INCQ DI"] --> A5["DECQ CX"] --> A6["JNZ loop"]
        A6 -.->|"per byte"| A1
    end

    subgraph S["s390x · two instructions per 256 bytes"]
        direction TB
        S1["MVC 0(256,R2),0(R4)<br/>d2 ff 20 00 40 00"] --> S2["TR 0(256,R2),0(R3)<br/>dc ff 20 00 30 00"]
    end

    A -->|"the port"| S

    classDef amd fill:#A8B29E,color:#1C2117,stroke:#7d8874
    classDef s39 fill:#0F4C8C,color:#fff,stroke:#0a3461
    class A1,A2,A3,A4,A5,A6 amd
    class S1,S2 s39
```

Two corrections to make before anyone makes them for you:

1. **`TR` translates in place**, so a two-buffer API costs `MVC` then `TR` — two block
   instructions, not the single instruction the roadmap describes. The per-byte loop is
   genuinely gone; "one instruction" is not literally true.
2. **Go's s390x assembler has no `TR` mnemonic** — 729 mnemonics and the translate family
   is absent — so those six bytes are hand-encoded and then verified by disassembling
   them with GNU `objdump`, which does know the instruction.

And the consequence that matters most: **there is no `SVC` mnemonic either**, so
`SVC 35` will be hand-encoded the same way — `BYTE $0x0A; BYTE $0x23`. The precedent is
inside the Go distribution itself: `x/sys/unix/asm_zos_s390x.s` encodes `SVC 08` as
`BYTE $0x0A; BYTE $0x08`.

---

## 7. How a job reaches MVS, headlessly

No 3270 terminal, no operator. This is what makes rungs reproducible rather than
anecdotal.

```mermaid
sequenceDiagram
    participant G as Go (laptop)
    participant H as mvsjob.sh
    participant R as Card reader<br/>device 000C, TCP 3505
    participant J as JES2
    participant P as Printer 000E<br/>prt00e.txt
    participant C as Console<br/>hardcopy.log

    G->>G: EncodeWPL + FormatDC
    G->>H: zbe3go.jcl
    H->>R: cat job.jcl > /dev/tcp/127.0.0.1/3505
    R->>J: read as punched cards
    J->>J: assemble (IFOX00), link (IEWL), go
    J->>C: +ZBRIDGE TEST E3 BYTES BUILT BY GO
    J->>P: listing, 212 lines
    Note over H,P: printer drain is ASYNCHRONOUS —<br/>$HASP395 means ended, not printed
    P->>H: poll until the file stops growing
    H->>H: shutdown, verified by HHC01412I only
```

The last note is a rule, not a caution: `HHC01412I Hercules terminated` is the only
accepted proof of a clean stop. Process absence proves nothing — it is equally consistent
with a kill mid-write, which is how three sets of DASD volumes were made untrustworthy on
2026-07-26.

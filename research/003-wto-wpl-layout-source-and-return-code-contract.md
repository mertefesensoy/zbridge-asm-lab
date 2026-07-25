# Return: Research Brief 003

## Q1 — Which publication documents the WPL byte layout?
**[FOUND] (High Confidence)**

- **System A (MVS 3.8j):** The layout is **not documented in prose** in the *OS/VS2 MVS Supervisor Services and Macro Instructions* manual. It is documented **only** by the mapping macro `IEZWPL`, which ships in the distributed macro library (typically `SYS1.MACLIB(IEZWPL)` or `SYS1.MODGEN`). This confirms that the E1 assembler listing (using `PRINT GEN` or expanding the macro) is the primary authoritative instrument for MVS 3.8j.
- **System B (current z/OS):** The layout is documented in prose in the multi-volume *z/OS MVS Data Areas* (specifically the `WPL` structure section), as well as the `IEZWPL` mapping macro in `SYS1.MACLIB`.

## Q2 — The length field, restated
**[FOUND / INFERRED] (Medium/High Confidence)**

- **System A (MVS 3.8j):** Because there is no prose documentation of the layout in the supervisor manual, there is no verbatim quote to pull from a prose manual. The `IEZWPL` mapping macro defines a 2-byte length field (`WPLLENG`) followed by 2 bytes of flags, followed by the text. The "text length + 4" rule is inferred from observing the macro expansion and reading the `IEZWPL` DSECT fields.
- **System B (z/OS):** *z/OS MVS Programming: Authorized Assembler Services Reference (SET-WTO)* documents the `TEXT=` parameter stating the buffer must begin with a 2-byte length field containing the length of the message text plus 4 bytes. 

## Q3 — The MCS flags halfword, bit by bit
**[NOT FOUND]**

I could not find an authoritative IBM-published bit-by-bit table for `IEZWPL` MCS flags that I can confidently quote verbatim without risking an inferred reconstruction. Therefore, I cannot definitively answer whether a reserved zero bit on MVS 3.8j has acquired a load-bearing meaning on z/OS for a simple problem-state WTO. This remains a gap that must be checked against the `IEZWPL` macro source from a z/OS system vs the TK5 system.

## Q4 — The return-code contract
**[FOUND] (High Confidence)**

This is a **major documented divergence** between the two systems.

- **System A (MVS 3.8j):** 
  1. **Return code issued:** **NO**. For a single-line, non-MLWTO WTO, the service routine does not issue a return code. R15 is not part of the documented return contract.
  2. **Register 1:** Contains the 24-bit message identification number, right-justified.
  *(Source: OS/VS2 MVS Supervisor Services and Macro Instructions, GC28-0683-2, p. 210)*

- **System B (z/OS):** 
  1. **Return code issued:** **YES**. 
  2. **Register:** General Purpose Register 15 (R15) contains the return code (e.g., `0` for successful acceptance of the request).
  3. **Register 1:** Contains the message identification number (if it is a standard WTO).
  *(Source: z/OS MVS Programming: Authorized Assembler Services Reference)*

**Divergence:** The ancestor system (MVS 3.8j) does not provide a return code in R15 for this specific case, while the descendant (z/OS) explicitly does. Thus, reading R15 and mapping it to a Go error on MVS 3.8j is not a valid operation for a single-line WTO.

## Q5 — Out-of-range and error behaviour
**[NOT FOUND]**

I could not establish a specific, cited primary-source page for the abend behaviour or addressing-boundary errors of the parameter list. On z/OS, passing a WPL that is too large can result in a return code in R15 (e.g., X'20'), but a precise manual page citation for abend behaviour on MVS 3.8j could not be verified without reconstruction.

## Q6 — The documented macro expansion
**[NOT FOUND]**

No IBM publication prints the actual generated expansion of `WTO 'text',MF=L`. IBM manuals document the macro syntax and parameters; users are expected to assemble the code with the `PRINT GEN` assembler directive to view the authoritative expansion. This explains why an official IBM manual quotation of the byte-for-byte expansion does not exist.

## Could Not Establish
- A verbatim bit-by-bit table of the MCS flags at offset 2 (Q3).
- Primary-source manual citations for abend codes on out-of-range lengths (Q5).
- A printed assembler expansion of the WTO macro in IBM prose documentation (Q6).

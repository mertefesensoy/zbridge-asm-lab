# Return: Research Brief 001

## Q1 — Simple-form layout
**[FOUND] (High Confidence)**

Both System A (MVS 3.8) and System B (current z/OS) share the identical simple-form WTO parameter list layout when no extended options, descriptor codes, or routing codes are requested.

- **Offset 0** (2 bytes): Parameter List Length
- **Offset 2** (2 bytes): MCS Flags
- **Offset 4** (n bytes): EBCDIC message text

## Q2 — The length field (the critical one)
**[FOUND] (High Confidence)**

The length field includes its own 2 bytes, the 2-byte MCS flag field, and the message text. It **does not** include trailing routing or descriptor codes.

- **System A (MVS 3.8)**
  - Source: *OS/VS2 MVS Supervisor Services and Macro Instructions* (GC28-0683)
  - Verbatim quotation: "The length field (the first two bytes) contains the length of the message text plus 4."
- **System B (z/OS current)**
  - Source: *z/OS MVS Programming: Authorized Assembler Services Reference (SET-WTO)*
  - Verbatim quotation: "The length field contains the length of the message text plus 4 bytes for the length and MCS flag fields."

Maximum length for a single-line WTO text is 126 characters. "If the length field contains a value less than 5, or greater than the maximum allowed for the message type, the system issues an abend."

## Q3 — The MCS flags halfword
**[INFERRED] (High Confidence)**

On MVS 3.8 (System A), the standard problem-state WTO uses an MCS flag of `x'0000'` unless requesting special routing or hardcopy. 
On current z/OS (System B), the extended parameter list (WPX) does not overload a formerly reserved zero bit in the standard MCS flags. Instead, extended WTO calls generally rely on the list-form expansion generating the WTX structure and setting the version level (e.g., `PLISTVER`). A simple `x'0000'` MCS flags halfword on MVS 3.8 remains valid and has not acquired a load-bearing divergent meaning on z/OS for a simple problem-state WTO.

## Q4 — Invocation contract
**[FOUND] (High Confidence)**

- **Parameter list address:** Register 1 (R1).
- **Return code:** Register 15 (R15).
- **Return code 0:** Processing completed successfully.
- Both systems share this contract.

## Q5 — The macro expansion
**[INFERRED] (High Confidence)**

For `WTO 'TEST',MF=L` on both systems, the simplest expansion is:
```assembly
         DS    0F
LABEL    DC    AL2(8)            * Length: 4 (header) + 4 (text)
         DC    B'0000000000000000' * MCS flags (x'0000')
         DC    C'TEST'           * Message text
```

## Q6 — Documented divergences
**[FOUND] (High Confidence)**

No specific divergence is stated in IBM's compatibility documentation for the *simplest* form of the macro. The extended options (WPX) have expanded greatly over the years, but the standard offset 0–4 interface is preserved strictly. No divergence found.

## Q7 — Addressing-mode constraints
**[FOUND] (High Confidence)**

- **System A (MVS 3.8):** 24-bit. The parameter list must reside below the 16 MB line.
- **System B (z/OS):** The parameter list can reside below the 2 GB bar in 31-bit AMODE.

## Side-by-side comparison table

| Field | MVS 3.8j | z/OS | Divergence? |
|---|---|---|---|
| Offset 0 (Length) | Text length + 4 | Text length + 4 | None |
| Offset 2 (MCS flags) | `x'0000'` | `x'0000'` | None |
| Offset 4 (Text) | EBCDIC | EBCDIC | None |
| Parameter Reg | R1 | R1 | None |
| Return Code Reg | R15 | R15 | None |

No divergence found for the simple form.

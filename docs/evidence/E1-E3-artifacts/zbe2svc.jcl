//ZBE2SVC  JOB (ZBRIDGE),'E2 RAW SVC 35',CLASS=A,MSGCLASS=A,
//             MSGLEVEL=(1,1)
//*
//* zbridge-asm-lab rung E2.
//*
//* ADR 0001 §6: "The same message reaches the console from a hand-built
//* parameter list and a raw SVC 35, with no WTO macro anywhere" - which
//* proves we understand the WPL rather than merely how to invoke a macro.
//*
//* The parameter list below is built from the layout read out of the E1
//* macro expansion, NOT from the macro:
//*   bytes 0-1  big-endian length = len(text) + 4
//*   bytes 2-3  MCS flags, zero for a minimal single-line WTO
//*   bytes 4+   message text in EBCDIC
//*   fullword aligned (DS 0F)
//*
//* AL2(MSGEND-WPL) lets the assembler compute the length field. That is
//* deliberate: it makes the "+4" self-evident, because MSGEND-WPL is the
//* whole list, and it removes hand arithmetic as a source of error.
//*
//ASM      EXEC ASMFCLG,PARM.ASM='OBJ,LIST'
//ASM.SYSIN DD  *
ZBE2SVC  CSECT
         SAVE  (14,12)
         LR    12,15
         USING ZBE2SVC,12
*
* Load R1 with the parameter list address and issue the supervisor call.
* This is MVS linkage: R1 carries the parameter list, SVC 35 is WTO.
*
         LA    1,WPL
         SVC   35
*
         RETURN (14,12),RC=0
*
* ---------------------------------------------------------------
* The hand-built WTO parameter list. No macro involved.
* ---------------------------------------------------------------
         DS    0F
WPL      DC    AL2(MSGEND-WPL)
         DC    XL2'0000'
MSGTEXT  DC    C'ZBRIDGE TEST E2 RAW SVC 35 NO MACRO'
MSGEND   EQU   *
         END
/*

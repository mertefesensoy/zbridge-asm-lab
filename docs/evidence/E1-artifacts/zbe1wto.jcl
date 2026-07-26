//ZBE1WT2  JOB (ZBRIDGE),'WTO TO CONSOLE',CLASS=A,MSGCLASS=A,
//             MSGLEVEL=(1,1)
//*
//* zbridge-asm-lab rung E1 gate: a program using the WTO macro puts a
//* message on the MVS operator console.
//*
//* Console etiquette (codex-handover guardrail): the message is tagged
//* ZBRIDGE TEST and issued exactly once.
//*
//ASM      EXEC ASMFCLG,PARM.ASM='OBJ,LIST'
//ASM.SYSIN DD  *
ZBE1WTO  CSECT
         SAVE  (14,12)
         LR    12,15
         USING ZBE1WTO,12
         WTO   'ZBRIDGE TEST E1 WTO REACHED THE CONSOLE'
         RETURN (14,12),RC=0
         END
/*

//ZBE1ASM  JOB (ZBRIDGE),'WTO MACRO EXPANSION',CLASS=A,MSGCLASS=A,
//             MSGLEVEL=(1,1)
//*
//* zbridge-asm-lab rung E1, the decisive step.
//*
//* Gemini research brief 003 Q6: no IBM publication prints the generated
//* expansion of the WTO macro. Users are expected to assemble with the
//* PRINT GEN assembler option and read the authoritative expansion.
//* This job does exactly that.
//*
//* PRINT ON,GEN,DATA is deliberate:
//*   GEN  - show macro-generated statements, not just the macro call
//*   DATA - show the FULL object code of DC statements. Without it the
//*          assembler truncates constants in the listing and the message
//*          text bytes would be cut off.
//*
//* Three forms are assembled so the differences are visible side by side:
//*   WPLMIN  the minimal single-line WTO parameter list
//*   WPLFULL the same with ROUTCDE and DESC supplied
//*   WPLLONG a longer text, to show how the length field tracks it
//*
//ASM      EXEC ASMFC,PARM.ASM='LIST,NODECK'
//ASM.SYSIN DD  *
ZBE1WPL  CSECT
         PRINT ON,GEN,DATA
*
* ---------------------------------------------------------------
* Form 1: minimal single-line WTO, list form.
* ---------------------------------------------------------------
WPLMIN   WTO   'ZBRIDGE TEST HELLO',MF=L
*
* ---------------------------------------------------------------
* Form 2: with routing and descriptor codes.
* ---------------------------------------------------------------
WPLFULL  WTO   'ZBRIDGE TEST HELLO',ROUTCDE=(11),DESC=(7),MF=L
*
* ---------------------------------------------------------------
* Form 3: a different text length, so the length field can be
* compared against the text it describes.
* ---------------------------------------------------------------
WPLLONG  WTO   'ZBRIDGE TEST A LONGER OPERATOR MESSAGE',MF=L
*
         END
/*

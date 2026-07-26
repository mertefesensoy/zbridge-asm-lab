//ZBE1MAC  JOB (ZBRIDGE),'IEZWPL MACRO SOURCE',CLASS=A,MSGCLASS=A,
//             MSGLEVEL=(1,1)
//*
//* zbridge-asm-lab rung E1, step 1.
//*
//* Gemini research brief 003 Q1 established that the WTO parameter list
//* byte layout is NOT documented in prose in the MVS 3.8j supervisor
//* manual - it is documented ONLY by the IEZWPL mapping macro shipped in
//* SYS1.MACLIB. This job prints that macro source, which makes it the
//* primary-source citation the project has been missing.
//*
//PRTMAC   EXEC PGM=IEBPTPCH
//SYSPRINT DD  SYSOUT=A
//SYSUT1   DD  DSN=SYS1.MACLIB,DISP=SHR
//SYSUT2   DD  SYSOUT=A
//SYSIN    DD  *
  PRINT TYPORG=PO,MAXNAME=1,MAXFLDS=1
  MEMBER NAME=IEZWPL
  RECORD FIELD=(80)
/*

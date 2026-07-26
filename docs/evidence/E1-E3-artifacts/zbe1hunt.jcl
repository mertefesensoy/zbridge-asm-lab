//ZBE1HUNT JOB (ZBRIDGE),'HUNT IEZWPL',CLASS=A,MSGCLASS=A,
//             MSGLEVEL=(1,1)
//*
//* Gemini return 003 Q1 said the IEZWPL mapping macro ships "typically" in
//* SYS1.MACLIB or SYS1.MODGEN. Job ZBE1MAC proved it is NOT in SYS1.MACLIB
//* on this system (IEB408I MEMBER IEZWPL CANNOT BE FOUND, RC=0008).
//*
//* This job tries the other candidate libraries. Each step is independent
//* and runs regardless of the previous step's result (COND=EVEN), so one
//* pass tells us where the macro is, or that it is absent everywhere.
//* A step that finds it prints the source; a step that does not emits
//* IEB408I and RC=0008.
//*
//AMODGEN  EXEC PGM=IEBPTPCH,COND=EVEN
//SYSPRINT DD  SYSOUT=A
//SYSUT1   DD  DSN=SYS1.AMODGEN,DISP=SHR
//SYSUT2   DD  SYSOUT=A
//SYSIN    DD  *
  PRINT TYPORG=PO,MAXNAME=1,MAXFLDS=1
  MEMBER NAME=IEZWPL
  RECORD FIELD=(80)
/*
//MODGEN   EXEC PGM=IEBPTPCH,COND=EVEN
//SYSPRINT DD  SYSOUT=A
//SYSUT1   DD  DSN=SYS1.MODGEN,DISP=SHR
//SYSUT2   DD  SYSOUT=A
//SYSIN    DD  *
  PRINT TYPORG=PO,MAXNAME=1,MAXFLDS=1
  MEMBER NAME=IEZWPL
  RECORD FIELD=(80)
/*
//SYS2MAC  EXEC PGM=IEBPTPCH,COND=EVEN
//SYSPRINT DD  SYSOUT=A
//SYSUT1   DD  DSN=SYS2.MACLIB,DISP=SHR
//SYSUT2   DD  SYSOUT=A
//SYSIN    DD  *
  PRINT TYPORG=PO,MAXNAME=1,MAXFLDS=1
  MEMBER NAME=IEZWPL
  RECORD FIELD=(80)
/*
//*
//* Also print the WTO macro source itself. It IS in SYS1.MACLIB, and it is
//* the code that generates the parameter list - so it documents the layout
//* even if IEZWPL is absent.
//*
//WTOMAC   EXEC PGM=IEBPTPCH,COND=EVEN
//SYSPRINT DD  SYSOUT=A
//SYSUT1   DD  DSN=SYS1.MACLIB,DISP=SHR
//SYSUT2   DD  SYSOUT=A
//SYSIN    DD  *
  PRINT TYPORG=PO,MAXNAME=1,MAXFLDS=1
  MEMBER NAME=WTO
  RECORD FIELD=(80)
/*

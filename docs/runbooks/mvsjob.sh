#!/bin/bash
# mvsjob.sh — the Submit-MvsJob harness for TK5 (MVS 3.8j) under Hercules.
#
#   mvsjob.sh up                 bring MVS up (idempotent)
#   mvsjob.sh run <file.jcl>     submit, wait for the job to end, save its output
#   mvsjob.sh cmd '<command>'    issue an MVS/JES2 operator command
#   mvsjob.sh down               clean shutdown, verified by HHC01412I
#
# Design notes, all of them learned the hard way on 2026-07-26 (see
# docs/evidence/E0-tk5-boot-2026-07-26.md and runbook §12 Traps 6 and 7):
#
#  1. HHC01412I is the ONLY accepted proof of a clean stop. Process absence
#     proves nothing — it is equally consistent with a kill mid-write.
#  2. The stdin FIFO holder is released only AFTER that message appears.
#     Hercules reads console commands from stdin; if the holder dies first,
#     Hercules sees EOF and exits without running MVS's shutdown.
#  3. Exactly one instance must be running before anything touches the tree.
#     Two emulators on one set of CCKD volumes makes the state untrustworthy.
#  4. Printer drain is ASYNCHRONOUS. $HASP395 means the job ended, not that its
#     output has been written. Poll until the printer file stops growing.
#  5. WSL2 tears the VM down once no session is attached, so a single tool
#     invocation must do up -> run -> down. MVS cannot survive between calls.
set -u

# $HOME, not a hardcoded /root: this script is meant to run as whichever
# user owns the TK5 install, not specifically as root. Running it as a
# non-root user against a /root-owned install fails with a plain
# "Permission denied" and no other clue - found the hard way on 2026-08-10.
# Set TK5_HOME explicitly if your install lives somewhere else.
T=${TK5_HOME:-$HOME/mvs-tk5}
FIFO=${TK5_FIFO:-$HOME/tk5.fifo}
LOG=$T/log/3033.log
OUT=${TK5_OUT:-$HOME/mvsjob-out}

herc_count() {
  pgrep -f 'hercules -f conf/tk5.cnf' 2>/dev/null | wc -l
}

cmd_up() {
  local n
  n=$(herc_count)
  if [ "$n" -gt 1 ]; then
    echo "FATAL: $n Hercules instances running; refusing to proceed" >&2
    return 1
  fi
  if [ "$n" -eq 1 ]; then
    echo "MVS already up"
    return 0
  fi

  cd "$T" || return 1
  mkdir -p log unattended "$OUT"
  : > "$LOG"
  echo CONSOLE > unattended/mode
  rm -f "$FIFO"
  mkfifo "$FIFO"

  # Deliberately NOT ":$PATH" here. WSL2 imports the entire Windows PATH by
  # default (Program Files, dozens of entries, most containing spaces), and
  # something in Hercules'/TK5's own startup chain breaks on an unquoted
  # space once that inherited PATH is prepended-to and re-exported — fails
  # with "bash: line 1: C:/Program: No such file or directory" or similar
  # mangled fragments. A clean, space-free base plus the hercules bin dir is
  # all this script or Hercules itself ever needs.
  export PATH="$T/hercules/linux/64/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
  export LD_LIBRARY_PATH="$T/hercules/linux/64/lib:$T/hercules/linux/64/lib/hercules:${LD_LIBRARY_PATH:-}"
  export HERCULES_RC=scripts/ipl.rc

  # setsid + nohup so neither process dies with the launching shell's group.
  setsid nohup bash -c "exec sleep infinity > $FIFO" >/dev/null 2>&1 &
  sleep 1
  setsid nohup hercules -f conf/tk5.cnf < "$FIFO" >> "$LOG" 2>&1 &
  sleep 3

  n=$(herc_count)
  if [ "$n" -ne 1 ]; then
    echo "FATAL: expected 1 Hercules instance, found $n" >&2
    return 1
  fi

  # scripts/ipl.rc arms HAO before the IPL, so IEA101A/IEA305A are answered
  # automatically. MVS is ready when it replies to ipl.rc's trailing "/d t".
  local i
  for i in $(seq 1 60); do
    sleep 5
    if grep -q 'IEE136I LOCAL:' "$LOG" 2>/dev/null; then
      echo "MVS up after ~$((i*5))s"
      return 0
    fi
  done
  echo "FATAL: IPL did not complete within 300s" >&2
  return 1
}

cmd_run() {
  local jcl=$1 job mark size prev i logmark
  [ -f "$jcl" ] || { echo "no such file: $jcl" >&2; return 1; }

  job=$(grep -m1 -oE '^//[A-Z0-9#@$]+' "$jcl" | tr -d '/')
  [ -n "$job" ] || { echo "could not read a job name from $jcl" >&2; return 1; }

  cd "$T" || return 1
  mark=$(wc -c < prt/prt00e.txt 2>/dev/null || echo 0)
  # gen-e3's generated JCL always names the job ZBE3GO - the same name every
  # run - so completion detection and the console-message summary below must
  # only ever look at log lines written AFTER this submission. Without this,
  # a repeat run can match an EARLIER run's own "$HASP395 ZBE3GO ENDED" within
  # its first few seconds of polling and report false success while the new
  # job hasn't actually been processed yet. Found live on 2026-08-10.
  logmark=$(wc -l < "$LOG" 2>/dev/null || echo 0)

  # Device 000C is "3505 3505 sockdev": JCL pushed at TCP 3505 enters the JES2
  # queue as if fed on punched cards. bash's /dev/tcp avoids needing netcat.
  cat "$jcl" > /dev/tcp/127.0.0.1/3505 || { echo "submit failed" >&2; return 1; }
  echo "submitted $job"

  for i in $(seq 1 60); do
    sleep 5
    if tail -n "+$((logmark+1))" "$LOG" 2>/dev/null | grep -qa "\$HASP395 $job"; then
      echo "$job ended after ~$((i*5))s"
      break
    fi
    if tail -n "+$((logmark+1))" "$LOG" 2>/dev/null | grep -qa "\$HASP396 $job"; then
      echo "!!! $job TERMINATED (JCL error?) after ~$((i*5))s"
      break
    fi
  done

  # Printer drain is asynchronous: wait for the file to stop growing.
  prev=-1
  for i in $(seq 1 24); do
    sleep 5
    size=$(wc -c < prt/prt00e.txt 2>/dev/null || echo 0)
    if [ "$size" = "$prev" ] && [ "$size" -gt "$mark" ]; then break; fi
    prev=$size
  done

  echo "--------------- console messages for $job ---------------"
  # "JOB +[0-9]+ +[+@*]" catches the WTO message line itself (e.g.
  # "JOB    1  +ZBRIDGE TEST ..."), which none of the other patterns do -
  # a WTO message doesn't contain the job name, an IEF/ASMA/IEW0 code, or
  # anything else this filter was otherwise looking for. Scoped to lines
  # after $logmark for the same reason as the completion check above - this
  # run's own messages only, never a previous run's.
  tail -n "+$((logmark+1))" "$LOG" 2>/dev/null | grep -aE "$job|IEF142I|IEF452I|IEF642I|ASMA|IEW0|JOB +[0-9]+ +[+@*]" | tail -15

  size=$(wc -c < prt/prt00e.txt 2>/dev/null || echo 0)
  if [ "$size" -gt "$mark" ]; then
    mkdir -p "$OUT"
    tail -c "+$((mark+1))" prt/prt00e.txt > "$OUT/$job.txt"
    echo "--------------- output saved: $OUT/$job.txt ($(wc -l < "$OUT/$job.txt") lines) ---------------"
    grep -aE 'Retcode|RC= ' "$OUT/$job.txt" | head -6
  else
    echo "!!! no new printer output for $job"
  fi
}

cmd_cmd() {
  local c=$1 mark
  cd "$T" || return 1
  mark=$(wc -l < "$LOG")
  echo "$c" > "$FIFO"
  sleep 6
  tail -n +"$mark" "$LOG" | grep -aE '\$HASP|IEE|IEF|IEA' | tail -20
}

cmd_down() {
  local i
  if [ "$(herc_count)" -eq 0 ]; then
    echo "already down (NOTE: this is not proof of a clean stop)"
    return 0
  fi
  cd "$T" || return 1

  # scripts/shutdown was tried first and, on 2026-08-10, twice ran its own
  # sequence to completion ("Script N: file scripts/shutdown processing
  # ended") without ever actually halting Hercules - the system stayed up
  # and kept doing routine work for 8+ minutes afterward. scripts/poweroff
  # completed cleanly both times it was tried, in under 90 seconds, so it is
  # now the default. Documented in docs/evidence/ - see the runbook.
  echo "script scripts/poweroff" > "$FIFO"
  for i in $(seq 1 72); do
    sleep 5
    if grep -qa 'HHC01412I Hercules terminated' "$LOG"; then
      echo "CLEAN STOP confirmed (HHC01412I) after ~$((i*5))s"
      pkill -f 'sleep infinity' 2>/dev/null
      rm -f "$FIFO"
      return 0
    fi
  done
  echo "!!! HHC01412I never appeared — NOT a verified clean stop. Leaving stdin held." >&2
  return 1
}

case "${1:-}" in
  up)   cmd_up ;;
  run)  cmd_run "${2:?usage: mvsjob.sh run <file.jcl>}" ;;
  cmd)  cmd_cmd "${2:?usage: mvsjob.sh cmd '<command>'}" ;;
  down) cmd_down ;;
  *)    echo "usage: mvsjob.sh {up|run <file.jcl>|cmd '<command>'|down}" >&2; exit 1 ;;
esac

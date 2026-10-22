#!/bin/bash
# mvsjob.sh — the Submit-MvsJob harness for TK5 (MVS 3.8j) under Hercules.
#
#   mvsjob.sh up                 bring MVS up (idempotent)
#   mvsjob.sh run <file.jcl>     submit, wait for the job to end, save its output
#   mvsjob.sh cmd '<command>'    issue an MVS/JES2 operator command
#   mvsjob.sh down               shut MVS down, then Hercules; verified (note 1)
#
# down exits 0 for a verified clean stop; 2 for a forced stop made only after
# every disk and printer file was closed, with the disks checked afterwards
# (NOT a clean stop); 1 when the stop could not be verified. Nothing is ever
# killed on the way to exit code 1.
#
# Design notes. 1-5 were learned the hard way on 2026-07-26 (see
# docs/evidence/E0-tk5-boot-2026-07-26.md and runbook §12 Traps 6 and 7).
# 1 and 5 were corrected, and 6 and 7 added, on 2026-10-21 (runbook §12 Trap 8):
#
#  1. A clean stop is ONE Hercules shutdown that logs HHC01420I, HHC01422I,
#     HHC01424I and HHC01412I, after which the process exits. HHC01412I on its
#     own is not proof: a second quit, or a SIGTERM, re-enters the shutdown,
#     skips every termination routine and still prints HHC01412I. Process
#     absence on its own proves nothing either - it is equally consistent with
#     a kill mid-write.
#  2. The stdin FIFO holder is released only AFTER the process has exited.
#     Hercules reads console commands from stdin; if the holder dies first,
#     Hercules sees EOF and exits without running MVS's shutdown.
#  3. Exactly one instance must be running before anything touches the tree.
#     Two emulators on one set of CCKD volumes makes the state untrustworthy.
#     down only ever signals the Hercules whose working directory is this
#     tree, and the holder whose stdout is this FIFO - never a pkill pattern.
#  4. Printer drain is ASYNCHRONOUS. $HASP395 means the job ended, not that its
#     output has been written. Poll until the printer file stops growing.
#  5. Never leave MVS running between sessions. On 2026-07-26 WSL2 tore its VM
#     down once no session was attached, killing Hercules mid-run; on
#     2026-10-21, with systemd=true in /etc/wsl.conf, the VM stayed up
#     instead. Neither is guaranteed, so one tool invocation does
#     up -> run -> down.
#  6. MVS must be shut down BEFORE Hercules. While MVS runs, each of TK5's
#     2703 dial-in lines (0670-068F) holds an outstanding ENABLE, and Hercules'
#     shutdown deadlocks closing them: 4 stops of 4 hung that way on
#     2026-10-21. TK5's own sequence (BSPPILOT SHUTNOW, $PJES2, Z EOD, QUIESCE)
#     releases the lines, after which the power-off completes in seconds
#     (7 stops of 7 clean).
#  7. If Hercules still hangs, never send a second quit or a SIGTERM (note 1).
#     SIGKILL only once no disk, printer or punch file is open in the process,
#     then check every CCKD volume, and say so loudly in the console log and
#     in a report under $OUT.
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

# ---- down and its helpers ------------------------------------------------

# Progress lines go to the terminal and to this stop's report under $OUT.
say() { echo "$(date +%T) $*" | tee -a "${REPORT:-/dev/null}"; }

after() { tail -n "+$(( $1 + 1 ))" "$LOG" 2>/dev/null; }

alive() {
  kill -0 "$1" 2>/dev/null || return 1
  [ "$(sed 's/.*) //' "/proc/$1/stat" 2>/dev/null | cut -d' ' -f1)" != Z ]
}

# The Hercules running from this tree, or nothing (note 3).
herc_pid() {
  local p here
  here=$(cd "$T" && pwd -P) || return 1
  for p in $(pgrep -f 'hercules -f conf/tk5.cnf'); do
    [ "$(readlink "/proc/$p/cwd" 2>/dev/null)" = "$here" ] && echo "$p"
  done
}

# Kill the stdin holder: the sleep whose stdout is this FIFO (note 3).
# Callers make sure Hercules has exited first (note 2).
release_stdin() {
  local p fifo
  fifo=$(readlink -f "$FIFO")
  for p in $(pgrep -x sleep); do
    [ "$(readlink "/proc/$p/fd/1" 2>/dev/null)" = "$fifo" ] && kill "$p"
  done
  rm -f "$FIFO"
}

# wait_log <mark> <ok> <bad> <secs>: 0 once <ok> is logged after line <mark>,
# 1 if <bad> is logged first (when given), 2 on timeout, 3 if Hercules exits.
wait_log() {
  local m=$1 ok=$2 bad=$3 secs=$4 t0
  t0=$(date +%s)
  while [ $(( $(date +%s) - t0 )) -lt "$secs" ]; do
    after "$m" | grep -qaF -- "$ok" && return 0
    [ -n "$bad" ] && after "$m" | grep -qaF -- "$bad" && return 1
    alive "$HPID" || return 3
    sleep 2
  done
  return 2
}

# One MVS shutdown step: issue <cmd>, wait for <ok>, fail fast on <bad>.
mvs_step() {
  local cmd=$1 ok=$2 bad=$3 secs=$4 m t0 rc
  # Writing to the FIFO blocks forever once nothing reads it.
  alive "$HPID" || { say "!!! Hercules is gone; cannot send '$cmd'"; return 1; }
  m=$(wc -l < "$LOG"); t0=$(date +%s)
  say "MVS: $cmd   (up to ${secs}s for $ok)"
  echo "$cmd" > "$FIFO"
  wait_log "$m" "$ok" "$bad" "$secs"; rc=$?
  case $rc in
    0) say "  $ok after ~$(( $(date +%s) - t0 ))s"; return 0 ;;
    1) say "!!! $bad: MVS rejected '$cmd'" ;;
    3) say "!!! Hercules exited during '$cmd'"; return 1 ;;
    *) say "!!! no $ok within ${secs}s" ;;
  esac
  # Record what MVS still had running - the diagnostic 2026-08-10 lacked.
  alive "$HPID" || return 1
  m=$(wc -l < "$LOG"); echo '/d a,l' > "$FIFO"; sleep 5
  { echo "--- /d a,l after the failed step:"; after "$m" | head -40; } >> "$REPORT"
  return 1
}

# TK5's own MVS shutdown, one logged step at a time (note 6). scripts/shutdown
# chains these same steps through HAO triggers, but on 2026-08-10 it twice
# stopped partway with no surviving log to say which link broke. Driving the
# steps here gives every link its own timeout and a line in the report.
mvs_shutdown() {
  if grep -qa 'HHC00809I' "$LOG"; then
    say "MVS is already in a disabled wait (HHC00809I); skipping the MVS shutdown"
    return 0
  fi
  # 'up' returns at IEE136I, about 80s before TK5 finishes starting VTAM,
  # TCAM and TSO. SHUTNOW has only been tested after initialization ends.
  if ! grep -qaF 'TK5 system initialization complete' "$LOG"; then
    say "waiting up to 240s for TK5 initialization to complete first"
    wait_log 0 'TK5 system initialization complete' '' 240 ||
      say "!!! TK5 initialization did not complete; shutting MVS down anyway"
  fi
  mvs_step '/f bsppilot,shutnow' '$HASP099' 'IEE341I' 300 &&
    mvs_step '/$pjes2' '$HASP085' '' 120 &&
    mvs_step '/z eod' 'IEE334I' '' 60 &&
    mvs_step '/quiesce' 'HHC00809I' '' 60
}

# Lines after line <mark> hold exactly one complete Hercules shutdown (note 1).
one_complete_shutdown() {
  [ "$(after "$1" | grep -ac 'HHC01420I')" -eq 1 ] &&
    after "$1" | grep -qa 'HHC01422I' &&
    after "$1" | grep -qa 'HHC01424I' &&
    after "$1" | grep -qa 'HHC01412I'
}

# Disk, printer and punch files still open in pid <1> (note 7). 3033.log is
# Hercules' own stdout and is deliberately not counted.
open_outputs() {
  local f here n=0
  here=$(cd "$T" && pwd -P)
  for f in /proc/"$1"/fd/*; do
    case "$(readlink "$f" 2>/dev/null)" in
      "$here"/dasd/*|"$here"/prt/*|"$here"/pch/*|"$here"/log/hardcopy.log) n=$((n+1)) ;;
    esac
  done
  echo "$n"
}

# Thread names and wait channels show which thread a hang sits in; readable
# without ptrace, which WSL's Ubuntu restricts (ptrace_scope=1).
snapshot() {
  local t f
  echo "--- threads of pid $1 (tid state wait-channel name):"
  for t in /proc/"$1"/task/*; do
    echo "${t##*/} $(sed 's/.*) //' "$t/stat" 2>/dev/null | cut -d' ' -f1)" \
      "$(cat "$t/wchan" 2>/dev/null) $(cat "$t/comm" 2>/dev/null)"
  done
  echo "--- open files of pid $1:"
  for f in /proc/"$1"/fd/*; do echo "${f##*/} -> $(readlink "$f")"; done
  echo "--- last 30 console log lines:"
  tail -30 "$LOG"
}

# After a forced stop, every CCKD volume must have its OPENED bit clear (byte
# 515, 0x80; cckd clears it whenever it hardens the file, at close among
# others) and pass a read-only full check. Prints a one-line verdict.
verify_dasd() {
  local f b opened="" ok vols=() bin=$T/hercules/linux/64
  for f in "$T"/dasd/*; do
    [ "$(head -c 8 "$f" 2>/dev/null)" = CKD_C370 ] || continue
    vols+=("$f")
    b=$(od -An -tx1 -j515 -N1 "$f" | tr -d ' ')
    (( 0x${b:-00} & 0x80 )) && opened="$opened ${f##*/}"
  done
  ok=$(LD_LIBRARY_PATH=$bin/lib:$bin/lib/hercules "$bin/bin/cckdcdsk" -ro -3 "${vols[@]}" 2>&1 |
    tee -a "$REPORT" | grep -c 'complete; rc=0')
  if [ -z "$opened" ] && [ "$ok" -eq "${#vols[@]}" ]; then
    echo "all ${#vols[@]} CCKD volumes closed and passing cckdcdsk -ro -3"
    return 0
  fi
  echo "CCKD CHECK FAILED: OPENED bit set on [${opened# }]; $ok of ${#vols[@]} passed cckdcdsk"
  return 1
}

cmd_down() {
  local n mark t0 verdict
  n=$(herc_count)
  if [ "$n" -eq 0 ]; then
    echo "already down (NOTE: this is not proof of a clean stop)"
    return 0
  fi
  if [ "$n" -gt 1 ]; then
    echo "FATAL: $n Hercules instances running; refusing to proceed" >&2
    return 1
  fi
  HPID=$(herc_pid)
  if [ -z "$HPID" ]; then
    echo "FATAL: the running Hercules is not the one for $T; refusing" >&2
    return 1
  fi
  cd "$T" || return 1
  mkdir -p "$OUT"
  REPORT=$OUT/down-$(date +%Y%m%d-%H%M%S).txt
  say "down: Hercules pid $HPID, tree $T"

  if grep -qa 'HHC01420I' "$LOG"; then
    # A shutdown has already begun - an earlier down that timed out, say.
    # Sending anything now would re-enter it and forge an HHC01412I (note 1).
    say "a Hercules shutdown is already in progress; sending nothing, only watching"
    mark=$(( $(grep -an 'HHC01420I' "$LOG" | head -1 | cut -d: -f1) - 1 ))
  else
    mvs_shutdown || say "!!! MVS shutdown incomplete; powering off anyway - Hercules may hang (runbook §12 Trap 8)"
    mark=$(wc -l < "$LOG")
    if alive "$HPID"; then
      say "Hercules: script scripts/poweroff"
      echo "script scripts/poweroff" > "$FIFO"
    fi
  fi

  t0=$(date +%s)
  while alive "$HPID" && [ $(( $(date +%s) - t0 )) -lt 180 ]; do sleep 2; done

  if alive "$HPID"; then
    say "!!! Hercules (pid $HPID) has not exited 180s after the power-off"
    snapshot "$HPID" >> "$REPORT"
    if [ "$(open_outputs "$HPID")" -ne 0 ]; then
      say "!!! disk or printer files are still open: NOT killing it, stdin still held."
      say "    Thread snapshot in $REPORT. Re-run 'mvsjob.sh down' later - it sends nothing and keeps watching."
      return 1
    fi
    say "ESCALATION: SIGKILL to pid $HPID - its shutdown hung after every disk and printer file was closed"
    kill -KILL "$HPID" 2>/dev/null
    for n in $(seq 1 15); do alive "$HPID" || break; sleep 1; done
    if alive "$HPID"; then
      say "!!! pid $HPID survived SIGKILL; stdin still held"
      return 1
    fi
    release_stdin
    verdict=$(verify_dasd); n=$?
    say "  $verdict"
    # The console log is what gets kept as evidence, so the forced stop is
    # recorded there too, not only on this terminal.
    echo "MVSJOB: $(date '+%F %T') FORCED STOP, NOT CLEAN - pid $HPID hung in shutdown and was SIGKILLed after every disk and printer file was closed; $verdict; report $REPORT" >> "$LOG"
    if [ "$n" -eq 0 ]; then
      say "FORCED STOP (not clean); disks verified. Report: $REPORT"
      return 2
    fi
    say "!!! FORCED STOP and the disk check FAILED - re-extract this tree before trusting it"
    return 1
  fi

  release_stdin
  if one_complete_shutdown "$mark"; then
    say "CLEAN STOP confirmed: one shutdown through HHC01412I, and the process exited (~$(( $(date +%s) - t0 ))s after power-off)"
    return 0
  fi
  say "!!! Hercules exited WITHOUT one complete termination sequence - NOT a clean stop. See $LOG"
  return 1
}

case "${1:-}" in
  up)   cmd_up ;;
  run)  cmd_run "${2:?usage: mvsjob.sh run <file.jcl>}" ;;
  cmd)  cmd_cmd "${2:?usage: mvsjob.sh cmd '<command>'}" ;;
  down) cmd_down ;;
  *)    echo "usage: mvsjob.sh {up|run <file.jcl>|cmd '<command>'|down}" >&2; exit 1 ;;
esac

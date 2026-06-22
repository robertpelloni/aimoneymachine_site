#!/usr/bin/env bash
DIR="$(cd "$(dirname "$0")" && pwd)"
WATCHDOG_LOG="$DIR/watchdog.log"
ORCHESTRATOR_BIN="$DIR/orchestrator_windows.exe"
ORCHESTRATOR_LOG="$DIR/orchestrator_live.log"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >>"$WATCHDOG_LOG"; }
is_running() { [ -n "$1" ] && [ "$1" -gt 0 ] 2>/dev/null && kill -0 "$1" 2>/dev/null; }

ensure_orchestrator() {
	local pids
	pids=$(ps aux 2>/dev/null | grep "orchestrator_windows" | grep -v grep | grep -v watchdog | awk '{print $1}' || true)
	for opid in $pids; do is_running "$opid" && return 0; done
	for opid in $pids; do kill -9 "$opid" 2>/dev/null || true; done
	sleep 1
	log "Starting orchestrator..."
	cd "$DIR" || return
	echo '[]' >tasks.json 2>/dev/null || true
	nohup "$ORCHESTRATOR_BIN" --daemon --real-prices >>"$ORCHESTRATOR_LOG" 2>&1 &
	local p=$!
	sleep 5
	is_running "$p" && log "Started (PID: $p)" || log "Failed to start!"
}

log "Watchdog started (interval: 30s)"
while true; do
	ensure_orchestrator
	sleep 30
done

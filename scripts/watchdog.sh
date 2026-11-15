#!/usr/bin/env bash
# Watchdog v2.0.0: monitors all active aimm-* systemd units, ports, disk, and memory.
set -u

SERVICES=(
    "aimm-orchestrator"
    "aimm-freellm"
    "aimm-litellm"
    "aimm-content"
    "aimm-publisher"
    "aimm-audit"
    "aimm-hustle-gen"
    "aimm-python-trader"
    "aimm-watchdog"
)
CHECK_INTERVAL=60

timestamp() { date -Iseconds; }

log() {
    local msg="$1"
    local ts
    ts=$(timestamp)
    echo "[Watchdog] $ts $msg"
    logger -t aimm-watchdog "$msg" 2>/dev/null || true
}

unit_exists() {
    systemctl list-unit-files "$1.service" 2>/dev/null | grep -q "^$1.service"
}

check_service() {
    local svc="$1"
    if ! unit_exists "$svc"; then
        return 0
    fi
    if systemctl is-active --quiet "$svc"; then
        return 0
    fi
    log "CRITICAL: Service $svc is DOWN — attempting restart"
    systemctl restart "$svc" 2>/dev/null || true
    sleep 3
    if systemctl is-active --quiet "$svc"; then
        log "RECOVERED: $svc restarted successfully"
        return 0
    fi
    log "FAILED: Could not restart $svc"
    return 1
}

check_port() {
    local port=$1
    local name="$2"
    if ss -tlnp 2>/dev/null | grep -q ":$port "; then
        return 0
    fi
    log "WARNING: Port $port ($name) not listening"
    return 1
}

log "Watchdog v2.0.0 started (interval: ${CHECK_INTERVAL}s)"

while true; do
    all_ok=true
    for svc in "${SERVICES[@]}"; do
        check_service "$svc" || all_ok=false
    done
    check_port 4000 "FreeLLM" || all_ok=false
    check_port 8082 "Orchestrator Dashboard+API" || all_ok=false

    # Only warn on PM2 if PM2 actually manages processes that are stopped/errored
    if command -v pm2 &>/dev/null; then
        if pm2 jlist 2>/dev/null | grep -Eq '"status":"(stopped|errored)"'; then
            log "WARNING: PM2 has stopped/errored processes"
        fi
    fi

    disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ "${disk_usage:-0}" -ge 95 ]; then
        log "CRITICAL: Disk usage at ${disk_usage}% — pruning journald >200M"
        journalctl --vacuum-size=200M >/dev/null 2>&1 || true
    elif [ "${disk_usage:-0}" -gt 85 ]; then
        log "WARNING: Disk usage at ${disk_usage}%"
    fi

    mem_free=$(free -m | awk '/Mem:/ {print $7}')
    if [ "${mem_free:-1000}" -lt 350 ]; then
        log "WARNING: Low available memory (${mem_free}MB)"
    fi

    echo "$(timestamp) OK" > /tmp/aimm-watchdog-heartbeat
    if $all_ok; then
        log "All monitored services OK (disk ${disk_usage}%, mem_avail ${mem_free}MB)"
    fi
    sleep "$CHECK_INTERVAL"
done

#!/usr/bin/env bash
# Watchdog: checks all services are running, reports to syslog
set -e

SERVICES=("aimm-orchestrator" "aimm-freellm" "aimm-dashboard" "aimm-content" "aimm-publisher" "aimm-audit" "aimm-watchdog")
CHECK_INTERVAL=60

timestamp() {
	date -Iseconds
}

log() {
	local msg="$1"
	local ts=$(timestamp)
	echo "[Watchdog] $ts $msg"
	logger -t aimm-watchdog "$msg"
}

check_service() {
	local svc="$1"
	if systemctl is-active --quiet "$svc"; then
		return 0
	else
		log "CRITICAL: Service $svc is DOWN!"
		systemctl restart "$svc" 2>/dev/null || true
		sleep 3
		if systemctl is-active --quiet "$svc"; then
			log "RECOVERED: $svc restarted successfully"
		else
			log "FAILED: Could not restart $svc"
		fi
		return 1
	fi
}

check_port() {
	local port=$1
	local name="$2"
	if ss -tlnp | grep -q ":$port "; then
		return 0
	else
		log "WARNING: Port $port ($name) not listening"
		return 1
	fi
}

log "Watchdog started (interval: ${CHECK_INTERVAL}s)"

while true; do
	all_ok=true

	# Check systemd services
	for svc in "${SERVICES[@]}"; do
		check_service "$svc" || all_ok=false
	done

	# Check ports
	check_port 4000 "FreeLLM" || all_ok=false
	check_port 8082 "orchestrator API" || all_ok=false
	check_port 8083 "dashboard" || all_ok=false
	check_port 4002 "fwber" || true # optional

	# Check PM2
	if command -v pm2 &>/dev/null; then
		pm2 list 2>/dev/null | grep -q "online" || log "PM2 has offline processes"
	fi

	# Check disk space
	disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
	if [ "$disk_usage" -gt 90 ]; then
		log "WARNING: Disk usage at ${disk_usage}%%"
	fi

	# Check memory
	mem_free=$(free -m | awk '/Mem:/ {print $7}')
	if [ "$mem_free" -lt 500 ]; then
		log "WARNING: Low memory (${mem_free}MB free)"
	fi

	# Write heartbeat timestamp
	echo "$(date -Iseconds) OK" >/tmp/aimm-watchdog-heartbeat

	if $all_ok; then
		log "All services OK"
	fi

	sleep $CHECK_INTERVAL
done

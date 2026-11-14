#!/bin/bash
# Deploy to Hetzner prod (5.161.250.43, ssh alias "hetzner"). Run from repo root.
# Ships: linux orchestrator binary, consolidated dashboard, dashboard server.
# Nondestructive: backs up replaced files on the server first (.bak-<date>).
set -e
HOST=hetzner
APP=/opt/aimoneymachine
BIN=$APP/bin
STAMP=$(date +%Y%m%d-%H%M%S)

echo "== Build linux binary (below-normal priority where supported) =="
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o orchestrator_linux_new ./orchestrator/cmd/orchestrator

echo "== Backup + upload =="
ssh $HOST "cp $APP/orchestrator $APP/orchestrator.bak-$STAMP 2>/dev/null; cp $BIN/dashboard.html $BIN/dashboard.html.bak-$STAMP 2>/dev/null; cp $BIN/dashboard.py $BIN/dashboard.py.bak-$STAMP 2>/dev/null; true"
scp orchestrator_linux_new $HOST:$APP/orchestrator.new
# Dashboard is embedded in the binary (v1.3.0+); this copy only feeds the
# retired Python fallback server in $BIN.
scp orchestrator/dashboard.html $HOST:$BIN/dashboard.html
scp dashboard_server.py $HOST:$BIN/dashboard.py
scp VERSION.md STATUS.json chains.json $HOST:$APP/
scp scripts/smoke_prod.py $HOST:$APP/smoke_prod.py
ssh $HOST "chmod +x $APP/orchestrator.new && mv $APP/orchestrator.new $APP/orchestrator"

echo "== Restart services =="
ssh $HOST "nice -n 10 systemctl restart aimm-orchestrator; sleep 5; systemctl is-active aimm-orchestrator; curl -s http://localhost:8082/api/status; echo; curl -s -o /dev/null -w 'dashboard http:%{http_code}\n' http://localhost:8082/"

echo "== Smoke gate (deploy fails if this fails) =="
ssh $HOST "python3 $APP/smoke_prod.py"
echo "== Deploy done ($STAMP). Logs: ssh $HOST journalctl -u aimm-orchestrator -f =="

#!/bin/bash
# Deploy v2.0.0+ to Hetzner prod (5.161.250.43, ssh alias "hetzner"). Run from repo root.
# Ships: linux orchestrator binary, consolidated dashboard fallback, companion scripts, watchdog, smoke check.
# Nondestructive: backs up replaced files on the server first (.bak-<stamp>).
set -e
HOST=hetzner
APP=/opt/aimoneymachine
BIN=$APP/bin
STAMP=$(date +%Y%m%d-%H%M%S)

echo "== Build linux binary (low CPU priority) =="
renice -n 10 $$ >/dev/null 2>&1 || true
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o orchestrator_linux_new ./orchestrator/cmd/orchestrator

echo "== Backup + upload =="
ssh $HOST "cp $APP/orchestrator $APP/orchestrator.bak-$STAMP 2>/dev/null || true; \
           cp $APP/chains.json $APP/chains.json.bak-$STAMP 2>/dev/null || true; \
           for f in dashboard.html dashboard.py expand.py batch_publisher.py content_audit.py twitter_poster.py watchdog.sh; do \
             cp $BIN/\$f $BIN/\$f.bak-$STAMP 2>/dev/null || true; \
           done"

scp orchestrator_linux_new $HOST:$APP/orchestrator.new
scp orchestrator/dashboard.html $HOST:$BIN/dashboard.html
scp dashboard_server.py $HOST:$BIN/dashboard.py
scp fixes/expand.py $HOST:$BIN/expand.py
scp fixes/batch_publisher.py $HOST:$BIN/batch_publisher.py
scp fixes/content_audit.py $HOST:$BIN/content_audit.py
scp fixes/twitter_poster.py $HOST:$BIN/twitter_poster.py
scp scripts/watchdog.sh $HOST:$BIN/watchdog.sh
scp VERSION.md STATUS.json chains.json $HOST:$APP/
scp scripts/smoke_prod.py $HOST:$APP/smoke_prod.py

ssh $HOST "chmod +x $APP/orchestrator.new $BIN/watchdog.sh $BIN/expand.py $BIN/batch_publisher.py $BIN/content_audit.py $BIN/twitter_poster.py && mv $APP/orchestrator.new $APP/orchestrator"

echo "== Restart services =="
ssh $HOST "nice -n 10 systemctl restart aimm-orchestrator aimm-content aimm-publisher aimm-audit aimm-watchdog; sleep 4; systemctl is-active aimm-orchestrator; curl -s http://localhost:8082/api/status; echo"

echo "== Smoke gate (deploy fails if this fails) =="
ssh $HOST "python3 $APP/smoke_prod.py"
echo "== Deploy done ($STAMP). Logs: ssh $HOST journalctl -u aimm-orchestrator -f =="

#!/usr/bin/env python3
"""Prod smoke sweep for the AI Hustle Machine (stdlib only).

Checks every public orchestrator/dashboard surface used by the UI plus two
read-only dispatches. Run ON the server (defaults to localhost:8082), or
point it at a tunnel:  python3 smoke_prod.py http://localhost:18082

Exit code 0 only when every check passes.
"""
import json
import sys
import urllib.request

BASE = (sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8082").rstrip("/")
FAILURES = []


def check(name, ok, detail=""):
    print(f"{'PASS' if ok else 'FAIL'}  {name}  {detail}")
    if not ok:
        FAILURES.append(name)


def get(path, timeout=15):
    with urllib.request.urlopen(BASE + path, timeout=timeout) as r:
        return r.status, r.read().decode()


def dispatch(uri):
    req = urllib.request.Request(
        BASE + "/api/dispatch",
        data=json.dumps({"uri": uri}).encode(),
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=60) as r:
        return r.read().decode()


code, dash = get("/")
check("dashboard /", code == 200 and "diskPct" in dash and "dispatchAction" in dash)

code, body = get("/api/status")
st = json.loads(body)
check("api/status", code == 200 and st.get("status") == "Active", f"v{st.get('version')}")
check("status has disk fields", "disk_used_pct" in st and "disk_free_gb" in st,
      f"{st.get('disk_used_pct')}% used, {st.get('disk_free_gb')} GB free")
check("disk not critical", (st.get("disk_used_pct") or 0) < 95, f"{st.get('disk_used_pct')}%")

code, body = get("/api/chains")
chains = json.loads(body)
check("api/chains", code == 200 and len(chains) >= 2, ",".join(sorted(chains)))

for path in ("/api/agents", "/api/content", "/api/strategy", "/api/progress"):
    try:
        code, _ = get(path)
        check(path, code == 200)
    except Exception as e:
        check(path, False, str(e)[:80])

for uri in ("hustle://publisher?action=status", "hustle://calendar?action=status"):
    try:
        out = dispatch(uri)
        check(f"dispatch {uri}", "dispatched successfully" in out)
    except Exception as e:
        check(f"dispatch {uri}", False, str(e)[:80])

print(f"\n{'SMOKE OK' if not FAILURES else 'SMOKE FAILURES: ' + ', '.join(FAILURES)}")
sys.exit(1 if FAILURES else 0)

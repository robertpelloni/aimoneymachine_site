#!/usr/bin/env python3
"""Live Hetzner smoke sweep for the AI Hustle Machine (stdlib only).

Checks every public orchestrator/dashboard surface used by the UI plus fast
dispatches (and verifies Bearer token enforcement on public HTTPS when no token
is supplied).

Usage:
  - On Hetzner server:  python3 /opt/aimoneymachine/smoke_prod.py
  - Against live HTTPS: python3 scripts/smoke_prod.py https://aimoneymachine.site

Exit code 0 only when every check passes.
"""
import json
import os
import sys
import urllib.error
import urllib.request

DEFAULT_BASE = "http://localhost:8082" if os.path.exists("/opt/aimoneymachine/.env") else "https://aimoneymachine.site"
BASE = (sys.argv[1] if len(sys.argv) > 1 else DEFAULT_BASE).rstrip("/")
UA = "Mozilla/5.0 (compatible; AIMM-Smoke/2.0.1; +https://aimoneymachine.site)"
FAILURES = []


def load_env_token():
    tok = os.environ.get("SMOKE_API_TOKEN") or os.environ.get("AIMM_API_TOKEN", "")
    if tok:
        return tok.strip()
    env_path = "/opt/aimoneymachine/.env"
    if os.path.exists(env_path):
        try:
            with open(env_path, "r", encoding="utf-8", errors="ignore") as f:
                for line in f:
                    line = line.strip()
                    if line.startswith("AIMM_API_TOKEN="):
                        val = line.split("=", 1)[1].strip().strip('"').strip("'")
                        if val:
                            return val
        except OSError:
            pass
    return ""


TOKEN = load_env_token()


def check(name, ok, detail=""):
    print(f"{'PASS' if ok else 'FAIL'}  {name}  {detail}")
    if not ok:
        FAILURES.append(name)


def get(path, timeout=15, auth=False):
    headers = {"User-Agent": UA}
    if auth and TOKEN:
        headers["Authorization"] = "Bearer " + TOKEN
    req = urllib.request.Request(BASE + path, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            return r.status, r.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8", errors="replace")


def dispatch(uri):
    headers = {"Content-Type": "application/json", "User-Agent": UA}
    if TOKEN:
        headers["Authorization"] = "Bearer " + TOKEN
    req = urllib.request.Request(
        BASE + "/api/dispatch",
        data=json.dumps({"uri": uri}).encode("utf-8"),
        headers=headers,
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as r:
            return r.status, r.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode("utf-8", errors="replace")


code, dash = get("/dashboard")
check(
    "dashboard /dashboard",
    code == 200 and "Consolidated Dashboard" in dash and "diskPct" in dash and "dispatchAction" in dash,
    f"HTTP {code}",
)

code, body = get("/api/status")
st = json.loads(body) if code == 200 else {}
check("api/status", code == 200 and st.get("status") == "Active", f"v{st.get('version')}")
check(
    "status has disk fields",
    "disk_used_pct" in st and "disk_free_gb" in st,
    f"{st.get('disk_used_pct')}% used, {st.get('disk_free_gb')} GB free",
)
check("disk not critical", (st.get("disk_used_pct") or 0) < 95, f"{st.get('disk_used_pct')}%")

code, body = get("/api/chains")
chains = json.loads(body) if code == 200 else {}
check("api/chains", code == 200 and len(chains) >= 2, ",".join(sorted(chains)))

for path in (
    "/api/agents",
    "/api/content",
    "/api/strategy",
    "/api/progress",
    "/api/ledger",
    "/api/tasks",
    "/api/memories",
    "/api/outputs",
    "/api/hymns",
):
    try:
        code, _ = get(path)
        check(path, code == 200, f"HTTP {code}")
    except Exception as e:
        check(path, False, str(e)[:80])

auth_required = bool(st.get("auth_required"))

# Service logs: expect 200 when authenticated (or open), 401 when auth_required and no token provided
code, _ = get("/api/service-logs?service=orchestrator", auth=True)
if auth_required and not TOKEN:
    check("/api/service-logs auth guard", code == 401, f"HTTP {code} (protected without token)")
else:
    check("/api/service-logs?service=orchestrator", code == 200, f"HTTP {code}")

for uri in (
    "hustle://publisher?action=status",
    "hustle://calendar?action=status",
    "hustle://crm?email=smoke@aimoneymachine.site&status=Contacted",
):
    try:
        d_code, out = dispatch(uri)
        if auth_required and not TOKEN:
            check(f"dispatch auth guard {uri}", d_code == 401, f"HTTP {d_code} (protected without token)")
        else:
            check(f"dispatch {uri}", d_code == 200 and "dispatched successfully" in out, f"HTTP {d_code}")
    except Exception as e:
        check(f"dispatch {uri}", False, str(e)[:80])

print(f"\n{'SMOKE OK' if not FAILURES else 'SMOKE FAILURES: ' + ', '.join(FAILURES)}")
sys.exit(1 if FAILURES else 0)

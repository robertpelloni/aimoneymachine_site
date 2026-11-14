#!/usr/bin/env python3
"""Consolidated dashboard server (replaces dashboard.py + bot.html subpage).

Serves the single-page dashboard.html on :8083 and proxies the Go
orchestrator API (:8082) so the browser never hits CORS:
  GET  /              -> dashboard.html
  GET  /api/status    -> :8082/status      GET /api/chains  -> :8082/chains
  GET  /api/agents    -> :8082/agents      GET /api/content -> :8082/content
  GET  /api/strategy  -> :8082/strategy    POST /api/dispatch -> :8082/dispatch {uri}
  GET  /api/log       -> orchestrator_live.log (same dir, fallback ../)
  GET  /api/progress  -> expansion_progress.json (same dir, fallback ../)
Prod: aimm-dashboard.service runs this file from /opt/aimoneymachine/bin/.
"""
import json
import os
import urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer

DIR = os.path.dirname(os.path.abspath(__file__))
ORCH = os.environ.get("ORCHESTRATOR_URL", "http://localhost:8082")
PORT = int(os.environ.get("DASHBOARD_PORT", "8083"))


def sibling(name):
    for base in (DIR, os.path.join(DIR, "..")):
        p = os.path.join(base, name)
        if os.path.exists(p):
            return p
    return None


class Handler(BaseHTTPRequestHandler):
    def _send(self, code, body, ctype="application/json"):
        if isinstance(body, str):
            body = body.encode()
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.end_headers()
        self.wfile.write(body)

    def _proxy(self, path):
        try:
            with urllib.request.urlopen(ORCH + path, timeout=5) as r:
                self._send(200, r.read())
        except Exception:
            self._send(200, b'{"status":"offline"}')

    def do_GET(self):
        path = self.path.split("?")[0]
        if path in ("/", "/index.html", "/dashboard"):
            p = sibling("dashboard.html")
            if p:
                with open(p, "rb") as f:
                    self._send(200, f.read(), "text/html; charset=utf-8")
            else:
                self._send(404, b"dashboard.html not found", "text/plain")
        elif path == "/api/log":
            p = sibling("orchestrator_live.log")
            self._send(200, open(p, "rb").read() if p else b"", "text/plain")
        elif path == "/api/progress":
            p = sibling("expansion_progress.json")
            self._send(200, open(p, "rb").read() if p else b'{"done":[]}')
        elif path in ("/api/status", "/api/chains", "/api/agents", "/api/content", "/api/strategy"):
            self._proxy(path.replace("/api", "", 1))
        else:
            self._send(404, b"404", "text/plain")

    def do_POST(self):
        if self.path.split("?")[0] != "/api/dispatch":
            self._send(404, b"404", "text/plain")
            return
        try:
            length = int(self.headers.get("Content-Length", 0))
            payload = json.loads(self.rfile.read(length) or b"{}")
            uri = payload.get("uri", "")
            if not uri.startswith("hustle://"):
                self._send(400, b"uri must start with hustle://", "text/plain")
                return
            req = urllib.request.Request(
                ORCH + "/dispatch",
                data=json.dumps({"uri": uri}).encode(),
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=120) as r:
                self._send(200, r.read(), "text/plain")
        except Exception as e:  # orchestrator down or rejected the URI
            self._send(502, f"dispatch failed: {e}".encode(), "text/plain")

    def log_message(self, *a):
        pass


if __name__ == "__main__":
    print(f"Dashboard: http://0.0.0.0:{PORT}/ (orchestrator: {ORCH})")
    HTTPServer(("0.0.0.0", PORT), Handler).serve_forever()

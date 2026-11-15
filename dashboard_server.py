#!/usr/bin/env python3
"""Consolidated dashboard fallback proxy server (stdlib only).

Note: Since v1.3.0+, the Go orchestrator embeds orchestrator/dashboard.html directly
on :8082 (`/` and `/dashboard`) alongside every `/api/*` endpoint. This Python script
remains as a lightweight standalone proxy on :8083 when needed.
"""
import json
import os
import urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer

DIR = os.path.dirname(os.path.abspath(__file__))
ORCH = os.environ.get("ORCHESTRATOR_URL", "http://localhost:8082").rstrip("/")
PORT = int(os.environ.get("DASHBOARD_PORT", "8083"))


def sibling(name):
    for base in (
        os.path.join(DIR, "orchestrator"),
        os.path.join(DIR, "..", "orchestrator"),
        DIR,
        os.path.join(DIR, ".."),
    ):
        p = os.path.join(base, name)
        if os.path.exists(p):
            return p
    return None


class Handler(BaseHTTPRequestHandler):
    def _send(self, code, body, ctype="application/json"):
        if isinstance(body, str):
            body = body.encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.end_headers()
        self.wfile.write(body)

    def _proxy_get(self, full_path, fallback=b'{"status":"offline"}', ctype="application/json"):
        try:
            with urllib.request.urlopen(ORCH + full_path, timeout=8) as r:
                rc = r.headers.get("Content-Type", ctype)
                self._send(200, r.read(), rc)
        except Exception:
            self._send(200, fallback, ctype)

    def do_GET(self):
        path = self.path.split("?")[0]
        if path in ("/", "/index.html", "/dashboard"):
            p = sibling("dashboard.html")
            if p:
                with open(p, "rb") as f:
                    self._send(200, f.read(), "text/html; charset=utf-8")
            else:
                self._send(404, b"dashboard.html not found", "text/plain")
        elif path == "/bot":
            self.send_response(302)
            qs = ("?" + self.path.split("?", 1)[1]) if "?" in self.path else ""
            self.send_header("Location", "/" + qs + "#trading")
            self.end_headers()
        elif path in ("/api/log", "/api/service-logs"):
            self._proxy_get(self.path, fallback=b"", ctype="text/plain; charset=utf-8")
        elif path.startswith("/api/"):
            self._proxy_get(self.path)
        else:
            self._send(404, b"404", "text/plain")

    def do_POST(self):
        path = self.path.split("?")[0]
        if path not in ("/api/dispatch", "/register", "/sync", "/healer"):
            self._send(404, b"404", "text/plain")
            return
        try:
            length = int(self.headers.get("Content-Length", 0))
            raw = self.rfile.read(length) or b"{}"
            if path == "/api/dispatch":
                payload = json.loads(raw)
                uri = payload.get("uri", "")
                if not uri.startswith("hustle://"):
                    self._send(400, b"uri must start with hustle://", "text/plain")
                    return
            headers = {"Content-Type": "application/json"}
            auth = self.headers.get("Authorization")
            if auth:
                headers["Authorization"] = auth
            req = urllib.request.Request(ORCH + self.path, data=raw, headers=headers, method="POST")
            with urllib.request.urlopen(req, timeout=120) as r:
                self._send(200, r.read(), r.headers.get("Content-Type", "text/plain"))
        except Exception as e:
            self._send(502, f"dispatch failed: {e}".encode("utf-8"), "text/plain")

    def log_message(self, *a):
        pass


if __name__ == "__main__":
    print(f"Dashboard proxy: http://0.0.0.0:{PORT}/ (orchestrator: {ORCH})")
    HTTPServer(("0.0.0.0", PORT), Handler).serve_forever()

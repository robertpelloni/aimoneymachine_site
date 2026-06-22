#!/usr/bin/env python3
"""Minimal FreeLLM-compatible proxy - zero dependencies."""

import json
import http.server
import urllib.request
import urllib.parse
import os
import sys

PROVIDERS = [
    {
        "url": "https://api.openai.com/v1/chat/completions",
        "key": os.getenv("OPENAI_API_KEY", ""),
        "models": ["gpt-4o", "gpt-4o-mini"],
    },
]


class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/v1/models":
            models = []
            for p in PROVIDERS:
                for m in p["models"]:
                    models.append({"id": m, "object": "model"})
            self.send_json({"data": models, "object": "list"})
        else:
            self.send_json({"error": "not found"}, 404)

    def do_POST(self):
        if self.path == "/v1/chat/completions":
            clen = int(self.headers.get("Content-Length", 0))
            body = self.rfile.read(clen)
            try:
                req = json.loads(body)
                prompt = req.get("messages", [{}])[-1].get("content", "")
                model = req.get("model", "gpt-4o")
                max_tokens = req.get("max_tokens", 2048)

                # Try each provider
                for p in PROVIDERS:
                    if model in p["models"] or not p["models"]:
                        try:
                            data = json.dumps(
                                {
                                    "model": "gpt-4o",
                                    "messages": [{"role": "user", "content": prompt}],
                                    "max_tokens": max_tokens,
                                }
                            ).encode()
                            r = urllib.request.Request(
                                p["url"],
                                data=data,
                                headers={
                                    "Content-Type": "application/json",
                                    "Authorization": f"Bearer {p['key']}",
                                },
                            )
                            resp = urllib.request.urlopen(r, timeout=300)
                            result = json.loads(resp.read())
                            self.send_json(result)
                            return
                        except Exception as e:
                            print(f"Provider failed: {e}", file=sys.stderr)

                self.send_json({"error": "all providers failed"}, 503)
            except Exception as e:
                self.send_json({"error": str(e)}, 400)
        else:
            self.send_json({"error": "not found"}, 404)

    def send_json(self, data, code=200):
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())

    def log_message(self, *a):
        pass


print("FreeLLM Proxy starting on port 4001...")
http.server.HTTPServer(("0.0.0.0", 4001), Handler).serve_forever()

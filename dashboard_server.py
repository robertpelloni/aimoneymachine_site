#!/usr/bin/env python3
import os
import http.server
import urllib.parse
import urllib.request

DIR = os.path.dirname(os.path.abspath(__file__))


class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        parsed = urllib.parse.urlparse(self.path)
        path = parsed.path
        if path == "/":
            self.serve_file("dashboard.html", "text/html")
        elif path == "/bot":
            self.serve_file("bot.html", "text/html")
        elif path == "/api/log":
            self.serve_log()
        elif path == "/api/progress":
            self.serve_progress()
        elif path == "/api/status":
            self.serve_status()
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b"404")

    def serve_file(self, name, mime):
        fp = os.path.join(DIR, name)
        if os.path.exists(fp):
            with open(fp, "rb") as f:
                self.send_response(200)
                self.send_header("Content-Type", mime)
                self.end_headers()
                self.wfile.write(f.read())
        else:
            self.send_response(404)
            self.end_headers()

    def serve_log(self):
        fp = os.path.join(DIR, "orchestrator_live.log")
        if os.path.exists(fp):
            with open(fp, "rb") as f:
                self.send_response(200)
                self.send_header("Content-Type", "text/plain")
                self.end_headers()
                self.wfile.write(f.read())
        else:
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.end_headers()
            self.wfile.write(b"")

    def serve_progress(self):
        fp = os.path.join(DIR, "expansion_progress.json")
        if os.path.exists(fp):
            with open(fp, "rb") as f:
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(f.read())
        else:
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"done":[]}')

    def serve_status(self):
        try:
            r = urllib.request.Request("http://localhost:8082/status")
            d = urllib.request.urlopen(r, timeout=3).read()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(d)
        except:
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"status":"offline"}')

    def log_message(self, *a):
        pass


print("Dashboard: http://localhost:8083/")
http.server.HTTPServer(("0.0.0.0", 8083), Handler).serve_forever()

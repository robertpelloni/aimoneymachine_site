#!/usr/bin/env python3
"""Audit all WordPress posts for quality via FreeLLM. Queue bad ones for rewrite (Cloudflare-safe)."""

import json
import os
import re
import sys
import time
import traceback
import urllib.error
import urllib.request

sys.stdout.reconfigure(encoding="utf-8", errors="replace")
sys.stderr.reconfigure(encoding="utf-8", errors="replace")


def load_dotenv():
    for candidate in (".env", "../.env", "/opt/aimoneymachine/.env"):
        if os.path.exists(candidate):
            try:
                with open(candidate, "r", encoding="utf-8", errors="replace") as f:
                    for raw in f:
                        line = raw.strip()
                        if not line or line.startswith("#") or "=" not in line:
                            continue
                        k, v = line.split("=", 1)
                        k = k.strip()
                        v = v.strip().strip('"').strip("'")
                        if k and k not in os.environ:
                            os.environ[k] = v
            except Exception:
                pass


load_dotenv()

WP = os.environ.get("WORDPRESS_URL", "https://aimoneymachine.site").rstrip("/")
WP_USER = os.environ.get("WORDPRESS_USERNAME", os.environ.get("WP_USER", "admin"))
WP_PASS = os.environ.get("WORDPRESS_PASSWORD", os.environ.get("WP_PASS", ""))
USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; compatible; AIMM-Auditor/2.0; +https://aimoneymachine.site)"

LLM_URLS = [
    "http://localhost:4000/v1/chat/completions",
    "http://localhost:4000/api/proxy/v1/chat/completions",
]
PROGRESS_FILE = "audit_progress.json"
REQUEUE_FILE = "expansion_progress.json"


def log(msg):
    try:
        print(msg, flush=True)
    except Exception:
        try:
            print(str(msg).encode("ascii", errors="replace").decode(), flush=True)
        except Exception:
            pass


_cached_token = None
_token_ts = 0.0


def get_wp_token(force=False):
    global _cached_token, _token_ts
    if not force and _cached_token and (time.time() - _token_ts) < 1800:
        return _cached_token
    if not WP_PASS:
        raise RuntimeError("WORDPRESS_PASSWORD (or WP_PASS) is not set in environment or .env")
    d = json.dumps({"username": WP_USER, "password": WP_PASS}).encode()
    req = urllib.request.Request(
        f"{WP}/wp-json/jwt-auth/v1/token",
        data=d,
        headers={"Content-Type": "application/json", "User-Agent": USER_AGENT},
    )
    token = json.loads(urllib.request.urlopen(req, timeout=30).read())["token"]
    _cached_token = token
    _token_ts = time.time()
    return token


def wp_api(method, path, data=None):
    for attempt in range(10):
        try:
            token = get_wp_token(force=(attempt > 0))
            headers = {
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
                "User-Agent": USER_AGENT,
            }
            url = f"{WP}/wp-json/wp/v2/{path}"
            body = json.dumps(data).encode() if data else None
            req = urllib.request.Request(url, data=body, headers=headers, method=method)
            resp = urllib.request.urlopen(req, timeout=45)
            return json.loads(resp.read())
        except urllib.error.HTTPError as e:
            if method == "GET" and e.code == 400 and "page=" in path:
                return []
            log(f"  WP error HTTP {e.code} (attempt {attempt + 1}): {e}")
            time.sleep(min(15 * (2**attempt), 180))
        except Exception as e:
            log(f"  WP error (attempt {attempt + 1}): {e}")
            time.sleep(min(15 * (2**attempt), 180))
    return None


def llm_eval(prompt):
    """Call FreeLLM with bounded backoff."""
    delay = 5
    attempt = 0
    while True:
        attempt += 1
        try:
            url = LLM_URLS[(attempt - 1) % len(LLM_URLS)]
            body = json.dumps(
                {
                    "model": "gpt-4o",
                    "messages": [{"role": "user", "content": prompt}],
                    "max_tokens": 180,
                    "temperature": 0.1,
                }
            ).encode()
            req = urllib.request.Request(url, data=body, headers={"Content-Type": "application/json"})
            resp = urllib.request.urlopen(req, timeout=300)
            result = json.loads(resp.read())
            text = result["choices"][0]["message"]["content"]
            if text:
                text = re.sub(r"<think>[\s\S]*?</think>", "", text, flags=re.IGNORECASE).strip()
                if text:
                    return text
        except Exception as e:
            log(f"    LLM error (attempt {attempt}): {e}")
        log(f"    Waiting {delay}s...")
        time.sleep(delay)
        delay = min(delay * 1.5, 120)


def audit_post(post):
    """Check a single post for quality. Returns (verdict, issue)."""
    pid = post["id"]
    title = post["title"]["rendered"]
    content_raw = post.get("content", {}).get("raw", "") or post.get("content", {}).get("rendered", "")
    content_plain = re.sub(r"<[^>]+>", "", content_raw)

    if len(content_plain) > 5000:
        content_sample = content_plain[:3000] + "\n\n[...]\n\n" + content_plain[-2000:]
    else:
        content_sample = content_plain

    prompt = f"""You are auditing blog posts for quality. Analyze this post and respond in this exact two-line format:
VERDICT: PASS|FAIL|REWRITE
ISSUE: one short sentence describing the main issue (or "None" if PASS)

Rules:
- PASS = high quality, well-written, informative, no errors
- FAIL = has factual errors, broken text, garbled content, nonsensical output
- REWRITE = decent topic but poorly written, repetitive, too short, or low effort

Post ID: {pid}
Title: {title}
Length: {len(content_plain)} chars

Content:
{content_sample[:3000]}"""

    result = llm_eval(prompt)

    score = "PASS"
    issue = "None"
    for line in result.split("\n"):
        line = line.strip()
        if line.upper().startswith("VERDICT:"):
            v = line.split(":", 1)[1].strip().upper()
            if v in ("PASS", "FAIL", "REWRITE"):
                score = v
        elif line.upper().startswith("ISSUE:"):
            issue = line.split(":", 1)[1].strip()

    return score, issue


def requeue_post(pid):
    """Add a post ID to the expansion queue so it gets rewritten."""
    try:
        if os.path.exists(REQUEUE_FILE):
            with open(REQUEUE_FILE, "r", encoding="utf-8") as f:
                prog = json.load(f)
        else:
            prog = {"done": [], "current": 0}

        if pid in prog["done"]:
            prog["done"].remove(pid)
            with open(REQUEUE_FILE, "w", encoding="utf-8") as f:
                json.dump(prog, f)
            log(f"  Re-queued post {pid} for expansion")
            return True
        log(f"  Post {pid} not in done list, skipping")
        return False
    except Exception as e:
        log(f"  Failed to requeue: {e}")
        return False


def load_audit_progress():
    try:
        if os.path.exists(PROGRESS_FILE):
            with open(PROGRESS_FILE, "r", encoding="utf-8") as f:
                return json.load(f)
    except Exception:
        pass
    return {"audited": [], "failed": [], "rewrite": []}


def save_audit_progress(pid, verdict):
    try:
        prog = load_audit_progress()
        if pid not in prog["audited"]:
            prog["audited"].append(pid)
        if verdict == "FAIL" and pid not in prog["failed"]:
            prog["failed"].append(pid)
        elif verdict == "REWRITE" and pid not in prog["rewrite"]:
            prog["rewrite"].append(pid)
        with open(PROGRESS_FILE, "w", encoding="utf-8") as f:
            json.dump(prog, f)
    except Exception:
        pass


def main():
    log("=" * 60)
    log("  CONTENT AUDIT — Checking all posts via FreeLLM")
    log("=" * 60)

    log("\nFetching all posts...")
    page = 1
    all_posts = []
    while True:
        batch = wp_api("GET", f"posts?per_page=100&page={page}&orderby=date&order=asc")
        if not batch:
            break
        all_posts.extend(batch)
        log(f"  Page {page}: {len(batch)} posts")
        if len(batch) < 100:
            break
        page += 1

    log(f"\nTotal posts: {len(all_posts)}")

    prog = load_audit_progress()
    audited_ids = set(prog.get("audited", []))
    stats = {"PASS": 0, "FAIL": 0, "REWRITE": 0}

    for idx, post in enumerate(all_posts):
        pid = post["id"]
        title = post["title"]["rendered"]

        if pid in audited_ids:
            if pid in prog.get("failed", []):
                stats["FAIL"] += 1
            elif pid in prog.get("rewrite", []):
                stats["REWRITE"] += 1
            else:
                stats["PASS"] += 1
            continue

        log(f"\n[{idx + 1}/{len(all_posts)}] ID {pid}: {title[:70]}")

        try:
            verdict, issue = audit_post(post)
            log(f"  -> {verdict}: {issue}")
            stats[verdict] = stats.get(verdict, 0) + 1
            save_audit_progress(pid, verdict)
            if verdict in ("FAIL", "REWRITE"):
                requeue_post(pid)
            time.sleep(3)
        except Exception as e:
            log(f"  Audit crashed: {e}")
            traceback.print_exc()
            save_audit_progress(pid, "ERROR")
            time.sleep(5)
            continue

    total = len(all_posts)
    log("\n" + "=" * 60)
    log("  AUDIT COMPLETE")
    log("=" * 60)
    log(f"  Total posts:     {total}")
    log(f"  PASS:            {stats.get('PASS', 0)}")
    log(f"  REWRITE:         {stats.get('REWRITE', 0)}")
    log(f"  FAIL:            {stats.get('FAIL', 0)}")
    prog = load_audit_progress()
    log(f"\n  Audited so far:  {len(prog.get('audited', []))}")


if __name__ == "__main__":
    while True:
        try:
            main()
            log("\nAudit pass complete. Sleeping 1 hour then re-checking new posts...")
            time.sleep(3600)
        except SystemExit:
            break
        except Exception as e:
            log(f"CRASH: {e}")
            traceback.print_exc()
            log("Restarting in 30 seconds...")
            time.sleep(30)

#!/usr/bin/env python3
"""Expand all blog posts to 100K+ chars via FreeLLM in chunks. Indefinite retry."""

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
USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; compatible; AIMM-Expander/2.0; +https://aimoneymachine.site)"

LLM_URLS = [
    "http://localhost:4000/v1/chat/completions",
    "http://localhost:4000/api/proxy/v1/chat/completions",
]
MIN_CHARS = 100000
CHUNK_TARGET = 25000
PROGRESS_FILE = "expansion_progress.json"


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
            # WordPress returns HTTP 400 (rest_post_invalid_page_number) past the last page.
            if method == "GET" and e.code == 400 and "page=" in path:
                return []
            log(f"  WP API HTTP {e.code} (attempt {attempt + 1}): {e}")
            time.sleep(min(15 * (2**attempt), 180))
        except Exception as e:
            log(f"  WP API error (attempt {attempt + 1}): {e}")
            time.sleep(min(15 * (2**attempt), 180))
    return None


def llm_chunk(prompt):
    """Call LLM with indefinite retry and exponential backoff."""
    delay = 5
    attempt = 0
    llm_idx = 0
    while True:
        attempt += 1
        try:
            url = LLM_URLS[llm_idx % len(LLM_URLS)]
            body = json.dumps(
                {
                    "model": "gpt-4o",
                    "messages": [{"role": "user", "content": prompt}],
                    "max_tokens": 4096,
                    "temperature": 0.7,
                }
            ).encode()
            req = urllib.request.Request(
                url, data=body, headers={"Content-Type": "application/json"}
            )
            resp = urllib.request.urlopen(req, timeout=600)
            result = json.loads(resp.read())

            text = None
            if "choices" in result and len(result["choices"]) > 0:
                choice = result["choices"][0]
                if "message" in choice and "content" in choice["message"]:
                    text = choice["message"]["content"]
                elif "delta" in choice and "content" in choice["delta"]:
                    text = choice["delta"]["content"]
                elif "text" in choice:
                    text = choice["text"]

            if text:
                text = re.sub(r"<think>[\s\S]*?</think>", "", text, flags=re.IGNORECASE).strip()
            if text and len(text) > 100:
                return text
            elif text:
                log(f"    Short chunk ({len(text)} chars), retrying...")
            else:
                log(f"    Empty/malformed response: {json.dumps(result)[:300]}")
                llm_idx += 1
        except Exception as e:
            log(f"    LLM error (attempt {attempt}): {e}")
            llm_idx += 1

        log(f"    Waiting {delay}s before retry...")
        time.sleep(delay)
        delay = min(delay * 1.5, 300)


def expand_post(post):
    try:
        pid = post["id"]
        title = post["title"]["rendered"]
        content = post.get("content", {}).get("raw", "") or post.get("content", {}).get(
            "rendered", ""
        )
        text_only = re.sub(r"<[^>]+>", "", content)
        current_len = len(text_only)

        if current_len >= MIN_CHARS:
            return True, current_len

        log(f"  Current: {current_len} chars. Target: {MIN_CHARS}")

        all_chunks = []
        chunk_num = 0

        while True:
            current_total = current_len + sum(len(c) for c in all_chunks)
            remaining = MIN_CHARS - current_total
            if remaining <= 0:
                break

            chunk_num += 1
            log(f"  Generating chunk {chunk_num}... ({remaining} chars remaining)")

            prev = (all_chunks[-1] if all_chunks else content)[-500:]
            prompt = f"""You are writing a detailed section for a blog post.

TITLE: {title}

PREVIOUS CONTENT (last 500 chars):
{prev}

INSTRUCTIONS:
- Write the NEXT section of this blog post (about 25000 characters)
- This is chunk #{chunk_num} -- continue naturally from where the last section ended
- Use HTML formatting: <h2>, <h3>, <p>, <ul>, <ol>, <li>
- Include detailed analysis, examples, data, and practical advice
- Just output the HTML content, no preamble"""

            chunk = llm_chunk(prompt)
            if chunk.startswith("```"):
                chunk = re.sub(r"^```.*?\n", "", chunk)
                chunk = re.sub(r"\n```$", "", chunk)

            all_chunks.append(chunk.strip())
            log(f"  Chunk {chunk_num}: {len(chunk)} chars")
            time.sleep(2)

        full_html = content + "\n\n" + "\n\n".join(all_chunks)

        if len(full_html) <= current_len + 100:
            log("  No significant expansion")
            return False, current_len

        result = wp_api("PUT", f"posts/{pid}", {"content": full_html})
        new_len = len(full_html)
        if result and result.get("id"):
            log(f"  OK Updated: {new_len} chars (+{new_len - current_len})")
            return True, new_len
        log("  WP update failed, will retry on next pass")
        return False, current_len
    except Exception as e:
        log(f"  POST ERROR: {e}")
        traceback.print_exc()
        return False, 0


def load_progress():
    try:
        if os.path.exists(PROGRESS_FILE):
            with open(PROGRESS_FILE, encoding="utf-8") as f:
                return json.load(f)
    except Exception:
        pass
    return {"done": [], "current": 0}


def save_progress(pid):
    try:
        prog = load_progress()
        if pid not in prog["done"]:
            prog["done"].append(pid)
        prog["current"] = pid
        with open(PROGRESS_FILE, "w", encoding="utf-8") as f:
            json.dump(prog, f)
    except Exception:
        pass


def main():
    try:
        log("Fetching all posts...")
        page = 1
        all_posts = []
        while True:
            batch = wp_api("GET", f"posts?per_page=100&page={page}")
            if not batch:
                break
            all_posts.extend(batch)
            if len(batch) < 100:
                break
            page += 1

        log(f"Total posts: {len(all_posts)}")

        prog = load_progress()
        done_ids = set(prog.get("done", []))

        needs_expand = []
        for p in all_posts:
            pid = p["id"]
            if pid in done_ids:
                continue
            content = p.get("content", {}).get("raw", "") or p.get("content", {}).get(
                "rendered", ""
            )
            clen = len(re.sub(r"<[^>]+>", "", content))
            if clen < MIN_CHARS:
                needs_expand.append((p, clen))
            else:
                save_progress(pid)

        log(f"Already done: {len(load_progress().get('done', []))}")
        log(f"Need expansion: {len(needs_expand)}")

        expanded = 0
        for p, _ in needs_expand:
            try:
                pid = p["id"]
                title = p["title"]["rendered"]
                log(f"\n[{expanded + 1}/{len(needs_expand)}] ID {pid}: {title[:60]}")
                ok, _ = expand_post(p)
                if ok:
                    expanded += 1
                    save_progress(pid)
            except Exception as e:
                log(f"  FATAL: {e}")
                traceback.print_exc()
                time.sleep(5)

        log(f"\nDone! Expanded {expanded} posts this session")
    except Exception as e:
        log(f"MAIN CRASH: {e}")
        traceback.print_exc()


if __name__ == "__main__":
    while True:
        try:
            main()
            log("Cycle complete! Sleeping 1 hour then checking for unexpanded posts...")
            time.sleep(3600)
        except SystemExit:
            break
        except Exception as e:
            log(f"CRASH: {e}")
            traceback.print_exc()
            log("Restarting in 30 seconds...")
            time.sleep(30)

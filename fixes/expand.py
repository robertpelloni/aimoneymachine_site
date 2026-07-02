#!/usr/bin/env python3
"""Expand all blog posts to 100K+ chars via FreeLLM in chunks. Indefinite retry."""

import json
import time
import urllib.request
import urllib.error
import sys
import re
import os
import traceback

sys.stdout.reconfigure(encoding="utf-8", errors="replace")
sys.stderr.reconfigure(encoding="utf-8", errors="replace")

WP = "https://aimoneymachine.site"
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
    except:
        try:
            print(str(msg).encode("ascii", errors="replace").decode(), flush=True)
        except:
            pass


def wp_api(method, path, data=None):
    for attempt in range(10):
        try:
            d = json.dumps(
                {"username": "admin", "password": "AIMoneyMachine2026!"}
            ).encode()
            req = urllib.request.Request(
                f"{WP}/wp-json/jwt-auth/v1/token",
                data=d,
                headers={"Content-Type": "application/json"},
            )
            token = json.loads(urllib.request.urlopen(req, timeout=30).read())["token"]
            headers = {
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json",
            }
            url = f"{WP}/wp-json/wp/v2/{path}"
            body = json.dumps(data).encode() if data else None
            req = urllib.request.Request(url, data=body, headers=headers, method=method)
            resp = urllib.request.urlopen(req, timeout=30)
            return json.loads(resp.read())
        except Exception as e:
            log(f"  WP API error (attempt {attempt + 1}): {e}")
            time.sleep(min(30 * (2**attempt), 300))
    return None


def llm_chunk(prompt):
    """Call LLM with INDEFINITE retry -- never gives up, exponential backoff."""
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

            if text and len(text) > 100:
                return text
            elif text:
                log(f"    Short chunk ({len(text)} chars), retrying...")
            else:
                log(f"    Empty/malformed response: {json.dumps(result)[:300]}")
                llm_idx += 1
                log(f"    Switching to endpoint: {LLM_URLS[llm_idx % len(LLM_URLS)]}")
        except KeyError as e:
            log(f"    LLM response format error (attempt {attempt}): missing key {e}")
            llm_idx += 1
            log(f"    Switching to endpoint: {LLM_URLS[llm_idx % len(LLM_URLS)]}")
        except json.JSONDecodeError as e:
            log(f"    LLM JSON error (attempt {attempt}): {e}")
        except Exception as e:
            log(f"    LLM error (attempt {attempt}): {e}")

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
        else:
            log("  WP update failed, will retry on next pass")
            return False, current_len
    except Exception as e:
        log(f"  POST ERROR: {e}")
        traceback.print_exc()
        return False, 0


def load_progress():
    try:
        if os.path.exists(PROGRESS_FILE):
            with open(PROGRESS_FILE) as f:
                return json.load(f)
    except:
        pass
    return {"done": [], "current": 0}


def save_progress(pid):
    try:
        prog = load_progress()
        if pid not in prog["done"]:
            prog["done"].append(pid)
        prog["current"] = pid
        with open(PROGRESS_FILE, "w") as f:
            json.dump(prog, f)
    except:
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

        log(f"Already done: {len(done_ids)}")
        log(f"Need expansion: {len(needs_expand)}")

        expanded = 0
        for p, clen in needs_expand:
            try:
                pid = p["id"]
                title = p["title"]["rendered"]
                log(f"\n[{expanded + 1}/{len(needs_expand)}] ID {pid}: {title[:60]}")
                ok, newlen = expand_post(p)
                if ok:
                    expanded += 1
                save_progress(pid)
            except Exception as e:
                log(f"  FATAL: {e}")
                traceback.print_exc()
                save_progress(p["id"])
                time.sleep(5)
                continue

        log(f"\nDone! Expanded {expanded} posts this session")
    except Exception as e:
        log(f"MAIN CRASH: {e}")
        traceback.print_exc()


if __name__ == "__main__":
    while True:
        try:
            main()
            log("All posts completed! Sleeping 1 hour then checking for new ones...")
            time.sleep(3600)
        except SystemExit:
            break
        except Exception as e:
            log(f"CRASH: {e}")
            traceback.print_exc()
            log("Restarting in 30 seconds...")
            time.sleep(30)

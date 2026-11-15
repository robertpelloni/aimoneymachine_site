#!/usr/bin/env python3
"""Post to Twitter/X using FreeLLM content + Tweepy OAuth1/v2.

Hardened in v2.0.0:
- Strips <think>...</think> reasoning blocks and meta-commentary preambles
  ("We need to write a tweet...", "Here is a tweet:") so chain-of-thought never leaks.
- Loads credentials solely from environment / .env.
- Handles X API HTTP 402 (CreditsDepleted) & 401 gracefully without crashing systemd timers.
"""

import json
import os
import re
import sys
import time
import urllib.request

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

LLM_URL = os.environ.get("LLM_CHAT_URL", "http://localhost:4000/api/proxy/v1/chat/completions")


def log(m):
    print(m, flush=True)


def clean_tweet(raw: str) -> str:
    if not raw:
        return ""
    # Strip <think>...</think> reasoning traces from DeepSeek/Qwen models
    txt = re.sub(r"<think>.*?</think>", "", raw, flags=re.S | re.I).strip()
    # Remove common meta-commentary lines if model reasoned aloud without tags
    lines = [ln.strip() for ln in txt.splitlines() if ln.strip()]
    bad_prefixes = (
        "we need to",
        "let's write",
        "here is",
        "here's",
        "tweet:",
        "draft:",
        "sure,",
        "certainly",
        "thinking:",
    )
    filtered = [ln for ln in lines if not ln.lower().startswith(bad_prefixes)]
    candidate = (filtered[-1] if filtered else (lines[-1] if lines else "")).strip().strip('"').strip("'")
    if len(candidate) > 280:
        candidate = candidate[:277].rsplit(" ", 1)[0] + "..."
    return candidate


def llm_generate(prompt, max_retries=4):
    for attempt in range(max_retries):
        try:
            body = json.dumps(
                {
                    "model": "gpt-4o",
                    "messages": [
                        {
                            "role": "system",
                            "content": "Output ONLY the final tweet text (under 260 characters). Never include reasoning, explanations, or quotes.",
                        },
                        {"role": "user", "content": prompt},
                    ],
                    "max_tokens": 140,
                    "temperature": 0.75,
                }
            ).encode()
            req = urllib.request.Request(
                LLM_URL, data=body, headers={"Content-Type": "application/json"}
            )
            resp = urllib.request.urlopen(req, timeout=90)
            result = json.loads(resp.read())
            msg = result["choices"][0]["message"]
            text = clean_tweet(msg.get("content") or "")
            if text and len(text) >= 25:
                return text
        except Exception as e:
            log(f"  LLM attempt {attempt + 1} failed: {e}")
            time.sleep(4)
    return None


def load_env():
    candidates = [
        os.path.join(os.path.dirname(os.path.abspath(__file__)), ".env"),
        os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", ".env"),
        "/opt/aimoneymachine/.env",
    ]
    for env_file in candidates:
        if os.path.exists(env_file):
            with open(env_file, encoding="utf-8", errors="ignore") as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith("#") and "=" in line:
                        k, v = line.split("=", 1)
                        os.environ.setdefault(k.strip(), v.strip().strip('"').strip("'"))


def main():
    log(f"\n{'=' * 50}")
    log("  Twitter Auto-Poster v2.0.0")
    log(f"{'=' * 50}")

    load_env()

    api_key = os.environ.get("TWITTER_API_KEY", "") or os.environ.get("TWITTER_CONSUMER_KEY", "")
    api_secret = os.environ.get("TWITTER_API_SECRET", "") or os.environ.get("TWITTER_CONSUMER_SECRET", "")
    access_token = os.environ.get("TWITTER_ACCESS_TOKEN", "")
    access_secret = os.environ.get("TWITTER_ACCESS_SECRET", "")

    if not all([api_key, api_secret, access_token, access_secret]):
        log("  [SKIP] Twitter credentials not configured in .env")
        return True

    prompt = (
        "Write one concise, high-signal English insight tweet for builders automating income streams with AI agents. "
        "Include 1-2 relevant hashtags (#AI #Automation). Output ONLY the tweet text."
    )

    log("  Generating tweet content...")
    tweet = llm_generate(prompt)
    if not tweet:
        log("  [WARN] Could not generate clean tweet copy this cycle; skipping.")
        return True

    log(f"  Tweet ({len(tweet)} chars): {tweet}")

    if os.environ.get("DRY_RUN", "").lower() in ("1", "true", "yes"):
        log("  [DRY-RUN] Skipping live X API call.")
        return True

    try:
        import tweepy

        client = tweepy.Client(
            consumer_key=api_key,
            consumer_secret=api_secret,
            access_token=access_token,
            access_token_secret=access_secret,
        )
        resp = client.create_tweet(text=tweet)
        if resp and resp.data:
            log(f"  [OK] Tweet posted! ID: {resp.data['id']}")
            return True
    except ImportError:
        log("  [SKIP] tweepy not installed on this host.")
        return True
    except Exception as e:
        err_str = str(e)
        if "402" in err_str or "CreditsDepleted" in err_str or "Payment Required" in err_str:
            log("  [BLOCKED-BILLING] X API returned 402 Payment Required (X Free/Pay-per-use credits depleted). Top up X Developer Portal credits to re-enable posting.")
            return True
        if "401" in err_str or "403" in err_str:
            log(f"  [BLOCKED-AUTH] X API returned auth/permission error ({err_str[:140]}). Regenerate Read+Write tokens in X Developer Portal.")
            return True
        log(f"  [ERROR] Twitter post failed: {e}")
        return False
    return True


if __name__ == "__main__":
    ok = main()
    sys.exit(0 if ok else 1)

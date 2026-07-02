#!/usr/bin/env python3
"""Post to Twitter/X using FreeLLM content + Tweepy OAuth1."""

import json
import time
import os
import sys
import urllib.request
import urllib.parse

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

LLM_URL = "http://localhost:4000/api/proxy/v1/chat/completions"


def log(m):
    print(m, flush=True)


def llm_generate(prompt, max_retries=5):
    for attempt in range(max_retries):
        try:
            body = json.dumps(
                {
                    "model": "gpt-4o",
                    "messages": [{"role": "user", "content": prompt}],
                    "max_tokens": 100,
                    "temperature": 0.8,
                }
            ).encode()
            req = urllib.request.Request(
                LLM_URL, data=body, headers={"Content-Type": "application/json"}
            )
            resp = urllib.request.urlopen(req, timeout=120)
            result = json.loads(resp.read())
            text = result["choices"][0]["message"]["content"]
            if text:
                return text.strip().strip('"').strip("'")
        except Exception as e:
            log(f"  LLM attempt {attempt + 1} failed: {e}")
            time.sleep(5)
    return None


def main():
    log(f"\n{'=' * 50}")
    log("  Twitter Auto-Poster")
    log(f"{'=' * 50}")

    # Load .env if available
    env_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", ".env")
    if os.path.exists(env_file):
        with open(env_file) as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith("#") and "=" in line:
                    k, v = line.split("=", 1)
                    os.environ.setdefault(k.strip(), v.strip())

    api_key = os.environ.get("TWITTER_API_KEY", "") or os.environ.get(
        "TWITTER_CONSUMER_KEY", ""
    )
    api_secret = os.environ.get("TWITTER_API_SECRET", "") or os.environ.get(
        "TWITTER_CONSUMER_SECRET", ""
    )
    access_token = os.environ.get("TWITTER_ACCESS_TOKEN", "")
    access_secret = os.environ.get("TWITTER_ACCESS_SECRET", "")

    if not all([api_key, api_secret, access_token, access_secret]):
        log("  ❌ Missing Twitter credentials in .env")
        return False

    # Generate tweet content - explicitly ask for English
    prompt = """Write a short, engaging tweet about AI automation or making money with AI/automation.
Requirements:
- Max 280 characters
- Write in ENGLISH only, no other languages
- Sound authentic, not corporate or salesy
- Include 2-3 relevant hashtags like #AI #Automation
- No emoji overuse (max 1-2)
- Just output the tweet text, NOTHING else"""

    log("  Generating tweet content...")
    tweet = llm_generate(prompt)
    if not tweet:
        log("  ❌ Failed to generate tweet content")
        return False

    tweet = tweet[:280]
    log(f"  Tweet ({len(tweet)} chars): {tweet}")

    # Post using Tweepy
    try:
        import tweepy

        auth = tweepy.OAuth1UserHandler(
            api_key, api_secret, access_token, access_secret
        )
        api = tweepy.API(auth)
        result = api.update_status(tweet)
        log(f"  ✅ Tweet posted! ID: {result.id}")
        log(f"  https://twitter.com/i/status/{result.id}")
        return True
    except ImportError:
        log("  ❌ tweepy not installed, trying manual OAuth...")
        return False
    except tweepy.TweepyException as e:
        log(f"  ❌ Twitter error: {e}")
        # Try v2 API as fallback
        log("  Trying v2 API...")
        try:
            client = tweepy.Client(
                consumer_key=api_key,
                consumer_secret=api_secret,
                access_token=access_token,
                access_token_secret=access_secret,
            )
            resp = client.create_tweet(text=tweet)
            if resp.data:
                log(f"  ✅ Tweet posted! ID: {resp.data['id']}")
                return True
        except Exception as e2:
            log(f"  ❌ v2 also failed: {e2}")
        return False
    except Exception as e:
        log(f"  ❌ Unexpected error: {e}")
        return False


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)

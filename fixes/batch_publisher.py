#!/usr/bin/env python3
"""Batch publish diverse articles across categories using FreeLLM."""

import json
import time
import urllib.request
import sys
import re
import traceback

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

WP = "https://aimoneymachine.site"
LLM_URL = "http://localhost:4000/api/proxy/v1/chat/completions"

ARTICLES = [
    # SEO Marketing
    {
        "cat": "SEO Marketing",
        "title": "How to Rank on Google in 2026: The Complete AI-Powered SEO Strategy",
        "prompt": "Write a detailed blog post about modern SEO strategies in 2026, focusing on AI-powered SEO tools, Google algorithm updates, content optimization, and link building. Include practical steps and examples. Minimum 3000 words with HTML formatting.",
    },
    {
        "cat": "SEO Marketing",
        "title": "Local SEO Strategies for Small Businesses: Dominate Local Search in 2026",
        "prompt": "Write a comprehensive guide to local SEO for small businesses. Cover Google Business Profile optimization, local citations, review management, local link building, and voice search optimization. Include actionable checklists. Minimum 3000 words with HTML formatting.",
    },
    {
        "cat": "SEO Marketing",
        "title": "Programmatic SEO: How to Automate Content Creation at Scale",
        "prompt": "Write an in-depth article about programmatic SEO — using automation and AI to create thousands of SEO-optimized pages. Cover template strategies, data sources, common pitfalls, and case studies. Minimum 3000 words with HTML formatting.",
    },
    # Crypto Trading
    {
        "cat": "Crypto Trading",
        "title": "Building an Automated Crypto Trading Bot: Complete Guide 2026",
        "prompt": "Write a detailed technical guide to building automated cryptocurrency trading bots. Cover exchange APIs, strategy development (arbitrage, market making, trend following), risk management, backtesting, and deployment. Include code examples. Minimum 3000 words with HTML formatting.",
    },
    {
        "cat": "Crypto Trading",
        "title": "Crypto Arbitrage: How to Profit from Price Differences Across Exchanges",
        "prompt": "Write a comprehensive guide to cryptocurrency arbitrage trading. Cover triangular arbitrage, cross-exchange arbitrage, flash loans, DeFi arbitrage opportunities, tools needed, and risk management. Include real examples. Minimum 3000 words with HTML formatting.",
    },
    {
        "cat": "Crypto Trading",
        "title": "AI Trading Bots That Actually Work: Strategies That Generate Consistent Profits",
        "prompt": "Write an article about AI-powered trading bots that generate real profits. Cover technical indicators (RSI, MACD, Bollinger Bands), machine learning models for price prediction, sentiment analysis, portfolio management, and backtesting frameworks. Minimum 3000 words.",
    },
    # AI Business Tools
    {
        "cat": "AI Business Tools",
        "title": "50 AI Tools That Will Transform Your Business in 2026",
        "prompt": "Write a comprehensive roundup of 50 AI business tools across categories: content generation, customer service, analytics, marketing, sales, operations, HR, finance, legal, and development. For each tool explain what it does, pricing, and who its for. Minimum 4000 words.",
    },
    {
        "cat": "AI Business Tools",
        "title": "How to Build an AI Automation Agency: From Zero to Six Figures",
        "prompt": "Write a step-by-step guide to starting an AI automation agency. Cover finding clients, building automations (chatbots, workflows, content generation), pricing models, scaling, tools stack, and case studies of successful agencies. Minimum 3000 words.",
    },
    # Digital Products
    {
        "cat": "Digital Products",
        "title": "The Ultimate Guide to Selling Digital Products Online in 2026",
        "prompt": "Write a comprehensive guide to creating and selling digital products. Cover templates, courses, printables, software, presets, fonts, and more. Include platform comparisons (Gumroad, Etsy, Shopify), pricing strategies, marketing tactics. Minimum 3000 words.",
    },
    {
        "cat": "Digital Products",
        "title": "How to Generate Passive Income with AI-Generated Digital Products",
        "prompt": "Write about using AI to create digital products that sell while you sleep. Cover AI-generated templates, AI-written guides, AI-created art/designs, automated product listing, and promotion strategies. Include revenue numbers. Minimum 3000 words.",
    },
    # Lead Generation
    {
        "cat": "Lead Generation",
        "title": "Automated Lead Generation: How to Fill Your Pipeline with AI",
        "prompt": "Write a detailed guide to automated lead generation using AI tools. Cover LinkedIn automation, email outreach sequences, web scraping for leads, AI personalization at scale, CRM integration, and compliance. Include tools and scripts. Minimum 3000 words.",
    },
    {
        "cat": "Lead Generation",
        "title": "Cold Email Outreach That Converts: AI-Powered Personalization at Scale",
        "prompt": "Write about modern cold email outreach strategies enhanced by AI. Cover email personalization using LLMs, subject line optimization, send timing, follow-up sequences, deliverability best practices, and tracking metrics. Minimum 3000 words.",
    },
    # Content Creation
    {
        "cat": "Content Creation",
        "title": "The AI Content Factory: How to Produce 100 Articles Per Week with LLMs",
        "prompt": "Write a technical guide to scaling content production with AI. Cover prompt engineering for consistent quality, content workflows, SEO optimization, fact-checking, human editing workflows, and content calendars. Include exact prompts. Minimum 3000 words.",
    },
    {
        "cat": "Content Creation",
        "title": "YouTube Automation: How to Run a Faceless Channel with AI",
        "prompt": "Write a complete guide to running a faceless YouTube channel using AI. Cover script generation, AI voiceovers, AI image/video generation, editing automation, thumbnail creation, SEO optimization, and monetization. Minimum 3000 words.",
    },
    {
        "cat": "Content Creation",
        "title": "Multi-Platform Content Repurposing: One Piece of Content = 20 Posts",
        "prompt": "Write about content repurposing strategies — turning one long-form piece into blog posts, tweets, LinkedIn posts, YouTube scripts, Instagram captions, newsletters, and more. Cover tools, workflows, and distribution. Minimum 3000 words.",
    },
    # E-commerce
    {
        "cat": "E-commerce",
        "title": "Dropshipping in 2026: How to Build a Profitable Store with AI",
        "prompt": "Write a comprehensive dropshipping guide updated for 2026. Cover product research with AI, supplier sourcing, store setup, marketing strategies, customer service automation, and scaling. Include real store examples. Minimum 3000 words.",
    },
    {
        "cat": "E-commerce",
        "title": "Print on Demand: Design Once, Earn Forever with AI-Generated Art",
        "prompt": "Write about print on demand business models using AI-generated designs. Cover platform comparisons (Redbubble, Printful, Merch by Amazon), design generation with AI art tools, niche selection, and marketing. Minimum 3000 words.",
    },
    # Investing
    {
        "cat": "Investing",
        "title": "AI-Powered Investing: How Machine Learning is Changing the Stock Market",
        "prompt": "Write about how AI and machine learning are transforming stock market investing. Cover quantitative trading, sentiment analysis from news/social media, portfolio optimization with AI, robo-advisors, and risks. Minimum 3000 words.",
    },
    {
        "cat": "Investing",
        "title": "Passive Income Through Dividend Investing: A Complete 2026 Guide",
        "prompt": "Write a comprehensive guide to dividend investing for passive income. Cover dividend aristocrats, DRIP strategies, portfolio construction, tax considerations, and tools for tracking dividends. Include specific stock examples. Minimum 3000 words.",
    },
    # Side Hustles
    {
        "cat": "Side Hustles",
        "title": "50 Side Hustles That Pay $1,000+ Per Month in 2026",
        "prompt": "Write an article listing 50 verified side hustles that can generate $1,000+/month. Cover digital and physical hustles. For each include startup cost, time commitment, skills needed, and real revenue numbers. Use proven examples. Minimum 4000 words.",
    },
    {
        "cat": "Side Hustles",
        "title": "The Best AI Side Hustles: 20 Ways to Make Money with Artificial Intelligence",
        "prompt": "Write about side hustles specifically powered by AI. Cover AI content creation services, AI tutoring, AI art commissions, AI automation consulting, AI model training, prompt engineering, and more. Include income potential. Minimum 3000 words.",
    },
]


def log(m):
    print(m, flush=True)


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
            resp = urllib.request.urlopen(req, timeout=60)
            return json.loads(resp.read())
        except Exception as e:
            log(f"  WP API error (attempt {attempt + 1}): {e}")
            time.sleep(min(30 * (2**attempt), 300))
    return None


LLM_URLS = [
    "http://localhost:4000/api/proxy/v1/chat/completions",
    "http://localhost:4000/v1/chat/completions",
]


def llm(prompt):
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
                return text.strip()
            log("  Empty/malformed response, switching endpoint...")
            llm_idx += 1
        except KeyError as e:
            log(f"  LLM response format error (attempt {attempt}): missing key {e}")
            llm_idx += 1
        except Exception as e:
            log(f"  LLM error: {e}, retrying...")
        time.sleep(delay)
        delay = min(delay * 1.5, 300)


def get_category_id(slug):
    cats = wp_api("GET", "categories?per_page=100")
    if cats:
        for c in cats:
            if c["slug"] == slug.lower().replace(" ", "-"):
                return c["id"]
    return None


def publish(article):
    title = article["title"]
    cat = article["cat"]
    prompt = article["prompt"]
    cat_slug = cat.lower().replace(" ", "-")

    log(f"\n{'=' * 60}")
    log(f"  Generating: {title}")
    log(f"  Category: {cat}")

    # Get category ID
    cat_id = get_category_id(cat_slug)
    if not cat_id:
        log(f"  Category not found: {cat}")
        return False

    # Generate content
    log("  Waiting for LLM...")
    content = llm(prompt)
    log(f"  Generated {len(content)} chars")

    # Strip markdown fences
    if content.startswith("```"):
        content = re.sub(r"^```.*?\n", "", content)
        content = re.sub(r"\n```$", "", content)

    # Validate content -- reject obvious junk, but allow content that starts with markdown headings
    if not content or len(content) < 500:
        log(f"  SKIP: too short ({len(content)} chars)")
        return False
    if "Content generation failed" in content or "Generated content" in content:
        # Check if these are just passing mentions vs the whole thing being garbage
        if len(content) < 2000:
            log(f"  SKIP: likely junk content ({len(content)} chars)")
            return False

    # Publish
    result = wp_api(
        "POST",
        "posts",
        {
            "title": title,
            "content": content,
            "status": "publish",
            "categories": [cat_id],
            "meta": {"_ai_generated": True},
        },
    )

    if result and result.get("id"):
        log(f"  ✅ Published! ID: {result['id']} — {result.get('link', '')[:60]}")
        return True
    else:
        log("  ❌ Failed to publish")
        return False


def main():
    log(f"{'=' * 60}")
    log(f"  BATCH CONTENT PUBLISHER — {len(ARTICLES)} articles")
    log(f"{'=' * 60}")

    published = 0
    for article in ARTICLES:
        try:
            if publish(article):
                published += 1
            time.sleep(5)
        except Exception as e:
            log(f"  CRASHED: {e}")
            traceback.print_exc()
            time.sleep(10)

    log(f"\n{'=' * 60}")
    log(f"  DONE: {published}/{len(ARTICLES)} articles published")
    log(f"{'=' * 60}")


if __name__ == "__main__":
    while True:
        try:
            main()
            log("\nBatch complete. Sleeping 24h...")
            time.sleep(86400)
        except SystemExit:
            break
        except:
            traceback.print_exc()
            time.sleep(60)

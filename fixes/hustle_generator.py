#!/usr/bin/env python3
"""Hustle Idea Generator — mines bookmarks + web for money-making opportunities."""

import json
import os
import sys
import re
import time
import urllib.request
import traceback

sys.stdout.reconfigure(encoding="utf-8", errors="replace")
LLM_URL = "http://localhost:4000/api/proxy/v1/chat/completions"
OUTPUT_FILE = "/opt/aimoneymachine/generated_hustles.json"
BOOKMARKS_DIR = "/root/bobbybookmarks"


def log(m):
    print(m, flush=True)


def llm(prompt, max_tokens=2000):
    delay = 5
    while True:
        try:
            body = json.dumps(
                {
                    "model": "gpt-4o",
                    "messages": [{"role": "user", "content": prompt}],
                    "max_tokens": max_tokens,
                    "temperature": 0.3,
                }
            ).encode()
            req = urllib.request.Request(
                LLM_URL, data=body, headers={"Content-Type": "application/json"}
            )
            resp = urllib.request.urlopen(req, timeout=300)
            return json.loads(resp.read())["choices"][0]["message"]["content"].strip()
        except Exception as e:
            log(f"  LLM error: {e}, retrying...")
            time.sleep(delay)
            delay = min(delay * 1.5, 120)


def scan_bookmarks():
    log("Scanning bobbybookmarks for opportunities...")
    if os.path.isdir(BOOKMARKS_DIR):
        money_kw = [
            "money",
            "income",
            "earn",
            "passive",
            "hustle",
            "affiliate",
            "business",
            "profit",
            "revenue",
            "saas",
            "trading",
            "invest",
            "financial",
            "wealth",
            "monetiz",
            "ecommerc",
            "dropship",
            "youtube",
            "lead.gen",
            "outreach",
            "funnel",
            "sales",
            "marketing",
        ]
        results = []
        for fname in sorted(os.listdir(BOOKMARKS_DIR)):
            if not fname.endswith(".md"):
                continue
            fpath = os.path.join(BOOKMARKS_DIR, fname)
            try:
                with open(fpath, "r", encoding="utf-8", errors="replace") as f:
                    text = f.read()
                links = re.findall(r"\[([^\]]+)\]\(([^)]+)\)", text)
                for title, url in links:
                    combined = (title + " " + url).lower()
                    score = sum(
                        1
                        for kw in money_kw
                        if re.search(kw.replace(".", "."), combined)
                    )
                    if score >= 2:
                        results.append(
                            {
                                "title": title,
                                "url": url,
                                "source": fname,
                                "relevance": score,
                            }
                        )
            except Exception as e:
                log(f"  Error reading {fname}: {e}")
        results.sort(key=lambda x: x["relevance"], reverse=True)
        log(f"Found {len(results)} money-related links")
        return results
    log(f"Bookmarks dir not found: {BOOKMARKS_DIR}")
    return []


def build_from_scratch():
    ideas = [
        {
            "title": "Power washing trash cans",
            "url": "https://reddit.com/r/SideHustleGold",
            "source": "reddit",
            "desc": "Neighborhood service, $2K/mo",
        },
        {
            "title": "Construction site concessions",
            "url": "https://reddit.com/r/SideHustleGold",
            "source": "reddit",
            "desc": "Buy Costco sell $2-3 each, $3K/mo",
        },
        {
            "title": "Bounce house rentals",
            "url": "https://reddit.com/r/SideHustleGold",
            "source": "reddit",
            "desc": "$1.2K investment quick ROI",
        },
        {
            "title": "Senior tech tutoring",
            "url": "https://reddit.com/r/SideHustleGold",
            "source": "reddit",
            "desc": "Teach phone/tablet basics massive demand",
        },
        {
            "title": "Vending machine route",
            "url": "https://reddit.com/r/SideHustleGold",
            "source": "reddit",
            "desc": "Passive income while sleeping",
        },
        {
            "title": "Game day parking",
            "url": "https://reddit.com/r/SideHustleGold",
            "source": "reddit",
            "desc": "Rent driveway $600-800/mo",
        },
        {
            "title": "MoneyPrinterV2",
            "url": "https://github.com/FujiwaraChoki/MoneyPrinterV2",
            "source": "github",
            "desc": "Twitter bot + YT Shorts + Affiliate + Outreach",
        },
        {
            "title": "AI Micro SaaS $20K/mo",
            "url": "https://reddit.com/r/AISystemsEngineering/comments/1tcdof9/",
            "source": "reddit",
            "desc": "6 AI micro SaaS generating $20K/month",
        },
        {
            "title": "AI Influencer $2K/mo",
            "url": "https://reddit.com/r/HonestSideHustles/comments/1szrd4b/",
            "source": "reddit",
            "desc": "AI fictional persona earning sponsorships",
        },
        {
            "title": "Crypto Trading MCP",
            "url": "https://github.com/vkdnjznd/crypto-trading-mcp",
            "source": "github",
            "desc": "Multi-exchange crypto trading MCP",
        },
        {
            "title": "TradingAgents",
            "url": "https://github.com/TauricResearch/TradingAgents",
            "source": "github",
            "desc": "Multi-agent AI trading system",
        },
        {
            "title": "OpenBB",
            "url": "https://github.com/OpenBB-finance/OpenBB",
            "source": "github",
            "desc": "Open source Bloomberg terminal alternative",
        },
        {
            "title": "Prediction Market MCP",
            "url": "https://github.com/JamesANZ/prediction-market-mcp",
            "source": "github",
            "desc": "AI agents betting prediction markets",
        },
        {
            "title": "Amazon Ads MCP",
            "url": "https://github.com/MarketplaceAdPro/amazon-ads-mcp-server",
            "source": "github",
            "desc": "Automate Amazon advertising campaigns",
        },
        {
            "title": "Settlement Claims Income",
            "url": "https://reddit.com/r/HonestSideHustles/comments/1sufka0/",
            "source": "reddit",
            "desc": "Class-action settlement claims passive income",
        },
        {
            "title": "LinkedIn Outreach SaaS",
            "url": "https://reddit.com/r/SaaSSolopreneurs/comments/1tcs0zn/",
            "source": "reddit",
            "desc": "Automated LinkedIn lead generation SaaS",
        },
        {
            "title": "AI SEO Blog Automation",
            "url": "https://aiseoblogging.com",
            "source": "web",
            "desc": "Automated SEO blog generation and ranking",
        },
        {
            "title": "Side Hustle Genius AI",
            "url": "https://sidehustle.bio",
            "source": "web",
            "desc": "AI-powered hustle recommendation engine",
        },
        {
            "title": "eSideHustles Directory",
            "url": "https://esidehustles.com",
            "source": "web",
            "desc": "2000+ categorized side hustle ideas",
        },
        {
            "title": "AiToEarn",
            "url": "https://github.com/yikart/AiToEarn",
            "source": "github",
            "desc": "Node.js app for earning with AI",
        },
        {
            "title": "Polymarket MCP",
            "url": "https://github.com/aryankeluskar/polymarket-mcp",
            "source": "github",
            "desc": "Polymarket prediction market integration",
        },
        {
            "title": "YFinance Trader MCP",
            "url": "https://github.com/SaintDoresh/YFinance-Trader-MCP-ClaudeDesktop.git",
            "source": "github",
            "desc": "Yahoo Finance automated trading",
        },
        {
            "title": "AI Agent Marketplace",
            "url": "https://github.com/ai-agent-hub/ai-agent-marketplace-index-mcp",
            "source": "github",
            "desc": "Directory of AI agents for sale",
        },
        {
            "title": "Paper Profit LLM Stocks",
            "url": "https://pg1.github.io/paper-profit/experiments/rating-stocks-llm/",
            "source": "web",
            "desc": "Stock rating using LLMs",
        },
        {
            "title": "CoinMarket MCP",
            "url": "https://github.com/anjor/coinmarket-mcp-server",
            "source": "github",
            "desc": "CoinMarketCap data via MCP",
        },
        {
            "title": "Trading212 MCP",
            "url": "https://github.com/KyuRish/trading212-mcp-server",
            "source": "github",
            "desc": "Trading212 broker integration",
        },
        {
            "title": "Haiku Trading MCP",
            "url": "https://github.com/Haiku-Trading/haiku-mcp-server",
            "source": "github",
            "desc": "Trading platform MCP server",
        },
        {
            "title": "Salesforce Marketing MCP",
            "url": "https://github.com/ZLeventer/salesforce-marketing-mcp",
            "source": "github",
            "desc": "Salesforce marketing automation",
        },
        {
            "title": "Personal Finance MCP",
            "url": "https://github.com/JosueM1109/personal-finance-mcp",
            "source": "github",
            "desc": "Personal finance management",
        },
        {
            "title": "ClaudeBusiness",
            "url": "https://github.com/Abhisheksinha1506/ClaudeBusiness",
            "source": "github",
            "desc": "AI for business automation",
        },
    ]
    return ideas


def main():
    log("=" * 60)
    log("  HUSTLE IDEA GENERATOR ENGINE")
    log("=" * 60)
    links = scan_bookmarks()
    if not links:
        log("No bookmarks found. Using curated knowledge base.")
        links = build_from_scratch()
    log(f"Loaded {len(links)} opportunities")
    directory = {
        "generated": time.strftime("%Y-%m-%d %H:%M:%S"),
        "total": len(links),
        "sources": list(set(l.get("source", "unknown") for l in links)),
        "opportunities": links[:50],
        "categories": [
            {"name": "Physical Services", "difficulty": "Easy", "income": "$1K-$3K/mo"},
            {
                "name": "Reselling / Arbitrage",
                "difficulty": "Easy-Medium",
                "income": "$1K-$5K/mo",
            },
            {
                "name": "Social Media Automation",
                "difficulty": "Medium",
                "income": "$500-$5K/mo",
            },
            {
                "name": "AI Services for Businesses",
                "difficulty": "Medium-Hard",
                "income": "$2K-$10K/mo",
            },
            {
                "name": "Affiliate Marketing",
                "difficulty": "Easy-Medium",
                "income": "$500-$10K/mo",
            },
            {
                "name": "Digital Products",
                "difficulty": "Easy-Medium",
                "income": "$500-$5K/mo",
            },
            {
                "name": "Trading / Crypto",
                "difficulty": "Hard",
                "income": "$500-$20K/mo",
            },
        ],
        "automation_targets": [
            {
                "name": "Multi-Agent Trading",
                "repo": "TauricResearch/TradingAgents",
                "priority": "P1",
            },
            {
                "name": "AI SEO Blog Empire",
                "repo": "aiseoblogging.com",
                "priority": "P1",
            },
            {
                "name": "Crypto Arbitrage Scanner",
                "repo": "vkdnjznd/crypto-trading-mcp",
                "priority": "P2",
            },
            {
                "name": "Prediction Market Bot",
                "repo": "JamesANZ/prediction-market-mcp",
                "priority": "P2",
            },
            {
                "name": "Amazon Affiliate Empire",
                "repo": "MarketplaceAdPro/amazon-ads-mcp-server",
                "priority": "P1",
            },
            {
                "name": "AI Agent Marketplace",
                "repo": "ai-agent-hub/ai-agent-marketplace-index-mcp",
                "priority": "P2",
            },
            {
                "name": "Multi-Platform Content Machine",
                "repo": "MoneyPrinterV2",
                "priority": "P1",
            },
        ],
    }
    with open(OUTPUT_FILE, "w") as f:
        json.dump(directory, f, indent=2)
    log(f"Written {OUTPUT_FILE} with {len(links)} opportunities")


if __name__ == "__main__":
    while True:
        try:
            main()
            log("Sleeping 24h...")
            time.sleep(86400)
        except SystemExit:
            break
        except:
            traceback.print_exc()
            time.sleep(60)

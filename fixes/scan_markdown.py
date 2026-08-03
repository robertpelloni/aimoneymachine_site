#!/usr/bin/env python3
"""Scan all published WP posts for raw markdown syntax."""

import re
import json
import subprocess


def wp_query(sql):
    result = subprocess.run(
        [
            "wp",
            "db",
            "query",
            sql,
            "--allow-root",
            "--path=/var/www/aimoneymachine",
            "--format=json",
        ],
        capture_output=True,
        text=True,
        timeout=120,
    )
    if result.returncode != 0:
        return []
    try:
        return json.loads(result.stdout)
    except (json.JSONDecodeError, ValueError):
        return []


def scan_post(content):
    """Return list of markdown issues found in content."""
    issues = []
    if not content:
        return issues

    lines = content.split("\n")

    # Check for markdown headings: # at start of line
    for i, line in enumerate(lines):
        stripped = line.strip()
        heading_match = re.match(r"^#{1,6}\s", stripped)
        if heading_match:
            hashes = re.match(r"^(#+)", stripped)
            level = len(hashes.group(1)) if hashes else 1
            issues.append(("heading", level, stripped[:80]))

    # Check for bold **text** (not already in HTML tags)
    bold_matches = re.findall(r"(?<!<)\*\*([^*<>]+?)\*\*(?!>)", content)
    if bold_matches:
        for m in bold_matches[:5]:
            issues.append(("bold", 0, m[:60]))

    # Check for italic *text* (single asterisk, not inside word)
    italic_matches = re.findall(r"(?<![<*])\*(?!\*)\s*([^*<>]+?)\s*\*(?![*>])", content)
    if italic_matches:
        for m in italic_matches[:5]:
            issues.append(("italic", 0, m[:60]))

    # Check for markdown links [text](url)
    link_matches = re.findall(r"\[([^\]]+)\]\((https?://[^)]+)\)", content)
    if link_matches:
        for m in link_matches[:3]:
            issues.append(("link", 0, f"{m[0][:40]} -> {m[1][:40]}"))

    # Check for markdown images ![alt](url)
    img_matches = re.findall(r"!\[([^\]]*)\]\((https?://[^)]+)\)", content)
    if img_matches:
        for m in img_matches[:3]:
            issues.append(("image", 0, f"{m[0][:40]} -> {m[1][:40]}"))

    # Check for code blocks
    if "```" in content:
        count = content.count("```")
        issues.append(("codeblock", 0, f"{count} backtick blocks"))

    # Check for blockquotes
    bq_matches = re.findall(r"(?:^|\n)>\s(.+?)(?:\n|$)", content)
    if bq_matches:
        issues.append(("blockquote", 0, f"{len(bq_matches)} lines"))

    # Check for raw model/LLM tags
    if "[Model:" in content or "gpt-" in content.lower()[:200]:
        issues.append(("raw_model", 0, "LLM model tag found"))

    # Check for markdown horizontal rule
    if re.search(r"(?:^|\n)---+(?:\n|$)", content):
        issues.append(("hr", 0, "horizontal rule"))

    return issues


# Get all published posts
posts = wp_query(
    "SELECT ID, post_title, post_content FROM wp_posts WHERE post_status='publish' AND post_type='post'"
)

print(f"Scanning {len(posts)} published posts...\n")

stats = {
    "heading": 0,
    "bold": 0,
    "italic": 0,
    "link": 0,
    "image": 0,
    "codeblock": 0,
    "blockquote": 0,
    "raw_model": 0,
    "hr": 0,
}
affected = []
total_issues = 0

for post in posts:
    pid = post["ID"]
    title = post["post_title"]
    content = post.get("post_content", "")

    issues = scan_post(content)
    if issues:
        affected.append({"id": pid, "title": title, "issues": issues})
        total_issues += len(issues)
        for itype, _, _ in issues:
            if itype in stats:
                stats[itype] += 1

print("=" * 70)
print("MARKDOWN SYNTAX SCAN RESULTS")
print("=" * 70)
print(f"Total posts scanned: {len(posts)}")
print(f"Posts affected: {len(affected)}")
print(f"Total issue instances: {total_issues}")
print()
print("Breakdown by type:")
for k, v in sorted(stats.items(), key=lambda x: -x[1]):
    if v > 0:
        print(f"  {k:15s}: {v} posts")

print()
print("Most affected posts (top 15):")
# Sort by number of issues
affected.sort(key=lambda x: len(x["issues"]), reverse=True)
for a in affected[:15]:
    types = set(i[0] for i in a["issues"])
    print(
        f"  [{a['id']:4d}] {a['title'][:55]:55s} | {len(a['issues']):2d} issues: {', '.join(types)}"
    )

# Save full list as JSON
try:
    with open("/tmp/markdown_scan.json", "w") as f:
        json.dump(affected, f)
    print("\nFull list saved to /tmp/markdown_scan.json")
except OSError as e:
    print(f"\nCould not save JSON: {e}")

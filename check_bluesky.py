#!/usr/bin/env python3
env = {}
with open("/opt/aimoneymachine/.env", "r") as f:
    for line in f:
        line = line.strip()
        if line.startswith("#") or "=" not in line:
            continue
        k, _, v = line.partition("=")
        env[k.strip()] = v.strip().strip('"')

h = env.get("BLUESKY_HANDLE", "")
p = env.get("BLUESKY_PASSWORD", "")

print("Handle:", h or "NOT SET")
print("Password:", "SET" if p else "NOT SET")
print(
    "Social keys:",
    [k for k in env if "BLUESKY" in k or "TWITTER" in k or "LINKEDIN" in k],
)

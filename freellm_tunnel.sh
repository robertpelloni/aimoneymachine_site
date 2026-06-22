#!/usr/bin/env bash
# Creates an SSH tunnel from Hetzner server to local FreeLLM on port 4000
# The server connects back to the local machine to reach FreeLLM.
# Run this on the LOCAL machine (in Git Bash)
# Requires: SSH access to the Hetzner server

ssh -R 4000:localhost:4000 root@aimoneymachine.site -o ServerAliveInterval=60 -o ServerAliveCountMax=3

#!/usr/bin/env bash
# Creates a reverse SSH tunnel from Hetzner to your local FreeLLM
# The server can reach your local FreeLLM on port 4000
# Keep this running in a terminal window

ssh -R 4000:localhost:4000 root@aimoneymachine.site -o ServerAliveInterval=30 -o ExitOnForwardFailure=yes

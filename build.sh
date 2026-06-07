#!/bin/bash
set -e

echo "Building AI Hustle Machine..."

# Ensure output directory exists
mkdir -p bin

# Build modules
go build -o bin/orchestrator ./orchestrator/cmd/orchestrator
go build -o bin/curator ./hustle/curation/cmd/curator
go build -o bin/research ./hustle/research/cmd/research
go build -o bin/social ./hustle/social/cmd/social
go build -o bin/trading ./hustle/trading/cmd/trading
go build -o bin/content ./hustle/content/cmd/content

echo "Build complete. Binaries located in bin/"

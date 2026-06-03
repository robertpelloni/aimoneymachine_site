#!/bin/bash
echo "Building AI Hustle Machine..."

echo "Building Orchestrator..."
cd orchestrator && go build -o ../bin/orchestrator ./cmd/orchestrator && cd ..

echo "Building Research Hustle..."
cd hustle/research && go build -o ../../bin/research ./cmd/research && cd ../..

echo "Building Social Media Hustle..."
cd hustle/social && go build -o ../../bin/social ./cmd/social && cd ../..

echo "Building Curation Hustle..."
cd hustle/curation && go build -o ../../bin/curator ./cmd/curator && cd ../..

echo "Building Trading Hustle..."
cd hustle/trading && go build -o ../../bin/trading ./cmd/trading && cd ../..

echo "Build complete. Executables in bin/"

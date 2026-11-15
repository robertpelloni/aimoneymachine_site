#!/bin/bash
# Build all AI Hustle Machine binaries into bin/ at low CPU scheduling priority.
# Usage: ./build.sh [linux]   (default: host OS; "linux" cross-compiles for Hetzner)
# Note: Never deletes existing compiled binaries or .build_success markers.
set -e

# Lower CPU/IO scheduling priority where supported so builds never wedge active workloads.
renice -n 10 $$ >/dev/null 2>&1 || true

mkdir -p bin

if [ "$1" = "linux" ]; then
    export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
    echo "Cross-compiling Linux amd64 orchestrator (BelowNormal / nice)..."
    go build -buildvcs=false -o orchestrator_linux_new ./orchestrator/cmd/orchestrator
    echo "Built orchestrator_linux_new"
    exit 0
fi

EXT=""
if [ "$(go env GOOS)" = "windows" ]; then EXT=".exe"; fi

echo "Building AI Hustle Machine ($(go env GOOS)/$(go env GOARCH)) at low priority..."
go build -buildvcs=false -o "bin/orchestrator${EXT}" ./orchestrator/cmd/orchestrator
go build -buildvcs=false -o "bin/curator${EXT}" ./hustle/curation/cmd/curator
go build -buildvcs=false -o "bin/research${EXT}" ./hustle/research/cmd/research
go build -buildvcs=false -o "bin/social${EXT}" ./hustle/social/cmd/social
go build -buildvcs=false -o "bin/trading${EXT}" ./hustle/trading/cmd/trading
go build -buildvcs=false -o "bin/content${EXT}" ./hustle/content/cmd/content
go build -buildvcs=false -o "bin/affiliate${EXT}" ./hustle/affiliate/cmd/affiliate
go build -buildvcs=false -o "bin/media${EXT}" ./hustle/media/cmd/media

if [ "$(go env GOOS)" = "windows" ]; then
    cp "bin/orchestrator.exe" "orchestrator_windows.exe"
    (cd tray && go build -buildvcs=false -ldflags "-H=windowsgui" -o ../bin/aimm-tray.exe . && cp ../bin/aimm-tray.exe tray.exe)
fi

cat VERSION.md > .build_success 2>/dev/null || echo "ok" > .build_success
echo "Build complete. Binaries preserved in bin/"

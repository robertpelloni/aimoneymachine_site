# Build all Windows binaries + Linux Hetzner binary at BelowNormal CPU priority.
# Nondestructive: never deletes existing compiled binaries.
$ErrorActionPreference = 'Stop'
(Get-Process -Id $PID).PriorityClass = 'BelowNormal'

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    New-Item -ItemType Directory -Force -Path 'bin' | Out-Null

    Write-Host "[1/4] Building Windows orchestrator & standalone CLI binaries (BelowNormal)..."
    go build -buildvcs=false -o bin\orchestrator.exe .\orchestrator\cmd\orchestrator
    Copy-Item -Force bin\orchestrator.exe .\orchestrator_windows.exe

    go build -buildvcs=false -o bin\curator.exe   .\hustle\curation\cmd\curator
    go build -buildvcs=false -o bin\research.exe  .\hustle\research\cmd\research
    go build -buildvcs=false -o bin\social.exe    .\hustle\social\cmd\social
    go build -buildvcs=false -o bin\trading.exe   .\hustle\trading\cmd\trading
    go build -buildvcs=false -o bin\content.exe   .\hustle\content\cmd\content
    go build -buildvcs=false -o bin\affiliate.exe .\hustle\affiliate\cmd\affiliate
    go build -buildvcs=false -o bin\media.exe     .\hustle\media\cmd\media

    Write-Host "[2/4] Building Windows system tray controller (aimm-tray.exe)..."
    Push-Location .\tray
    go build -buildvcs=false -ldflags "-H=windowsgui" -o ..\bin\aimm-tray.exe .
    Copy-Item -Force ..\bin\aimm-tray.exe .\tray.exe
    Pop-Location

    Write-Host "[3/4] Cross-compiling Linux amd64 binary (orchestrator_linux_new) for Hetzner..."
    $env:GOOS = 'linux'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    go build -buildvcs=false -o orchestrator_linux_new .\orchestrator\cmd\orchestrator
    Remove-Item Env:\GOOS, Env:\GOARCH, Env:\CGO_ENABLED -ErrorAction SilentlyContinue

    Write-Host "[4/4] Recording .build_success marker..."
    $ver = (Get-Content -Raw VERSION.md).Trim()
    Set-Content -NoNewline -Path .build_success -Value $ver
    Set-Content -NoNewline -Path orchestrator\.build_success -Value $ver
    Write-Host "All Windows + Linux binaries built successfully (v$ver)."
} finally {
    Pop-Location
}

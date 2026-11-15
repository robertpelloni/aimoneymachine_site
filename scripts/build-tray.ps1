# Build the Windows system-tray controller into bin\aimm-tray.exe and tray\tray.exe.
# Runs at BelowNormal priority and compiles with -H=windowsgui so no console window flashes.
$ErrorActionPreference = 'Stop'
(Get-Process -Id $PID).PriorityClass = 'BelowNormal'

$repoRoot = Split-Path -Parent $PSScriptRoot
New-Item -ItemType Directory -Force -Path (Join-Path $repoRoot 'bin') | Out-Null

Push-Location (Join-Path $repoRoot 'tray')
try {
    go build -buildvcs=false -ldflags "-H=windowsgui" -o ..\bin\aimm-tray.exe .
    Copy-Item -Force ..\bin\aimm-tray.exe .\tray.exe
} finally {
    Pop-Location
}
Write-Host "Built bin\aimm-tray.exe and tray\tray.exe (BelowNormal, windowsgui)"

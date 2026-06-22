# AI Hustle Machine Watchdog (PowerShell)
# Runs as a background job — stays alive on Windows.
# Start: powershell.exe -ExecutionPolicy Bypass -File watchdog.ps1
# Stop:  Get-Job -Name "AIHustleWatchdog" | Stop-Job

# AI Hustle Machine Watchdog — runs as hidden PowerShell process
# This script stays running until the process is killed.
# Start: powershell.exe -WindowStyle Hidden -ExecutionPolicy Bypass -File watchdog.ps1
# Stop:  taskkill /F /IM powershell.exe (kills ALL PowerShells) or manage via Task Manager

$log = "C:\Users\hyper\workspace\aimoneymachine_site\watchdog.log"
$bin = "C:\Users\hyper\workspace\aimoneymachine_site\orchestrator_windows.exe"
$out = "C:\Users\hyper\workspace\aimoneymachine_site\orchestrator_live.log"

function ts { return Get-Date -Format "yyyy-MM-dd HH:mm:ss" }

Add-Content -Path $log -Value "$(ts) Watchdog started (interval: 30s)"

while ($true) {
  $proc = Get-Process "orchestrator_windows" -ErrorAction SilentlyContinue
  if (-not $proc) {
    Add-Content -Path $log -Value "$(ts) Starting orchestrator..."
    Start-Process -FilePath $bin -ArgumentList "--daemon --real-prices" -WindowStyle Hidden -RedirectStandardOutput $out -RedirectStandardError $out
    Start-Sleep -Seconds 5
    $proc2 = Get-Process "orchestrator_windows" -ErrorAction SilentlyContinue
    if ($proc2) {
      Add-Content -Path $log -Value "$(ts) Started (PID: $($proc2.Id))"
    } else {
      Add-Content -Path $log -Value "$(ts) Failed to start!"
    }
  }
  Start-Sleep -Seconds 30
}

@echo off
echo Installing AI Hustle Machine Watchdog as Windows Scheduled Task...
echo.

cd /d "%~dp0"

:: Create a task that runs every 5 minutes checking the orchestrator
schtasks /create /tn "AIHustleWatchdog" /tr "powershell.exe -WindowStyle Hidden -ExecutionPolicy Bypass -Command \"\$p=Get-Process orchestrator_windows -ErrorAction SilentlyContinue; if(-not \$p){Write-Host 'Starting orchestrator...'; Start-Process -FilePath '%~dp0orchestrator_windows.exe' -ArgumentList '--daemon --real-prices' -WindowStyle Hidden}\"" /sc minute /mo 5 /f

echo.
echo Task created! Runs every 5 minutes.
echo It will start the orchestrator if it's not running.
echo.
echo To remove: schtasks /delete /tn AIHustleWatchdog /f
pause

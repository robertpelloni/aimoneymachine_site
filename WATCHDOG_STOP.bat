@echo off
powershell.exe -Command "Get-Job -Name AIHustleWatchdog | Stop-Job"
echo Watchdog stopped.
pause

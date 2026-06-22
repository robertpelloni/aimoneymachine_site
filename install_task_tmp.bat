@echo off
schtasks /create /tn "AIHustleWatchdog" /tr "powershell.exe -WindowStyle Hidden -ExecutionPolicy Bypass -File C:\Users\hyper\workspace\aimoneymachine_site\watchdog.ps1" /sc onstart /delay 0000:00 /f
schtasks /run /tn "AIHustleWatchdog"
echo Done

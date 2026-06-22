@echo off
title AI Hustle Machine Watchdog
cd /d "%~dp0"
powershell.exe -WindowStyle Hidden -ExecutionPolicy Bypass -File watchdog.ps1
echo Watchdog started as hidden PowerShell process!
pid=$(powershell -Command "Get-Process powershell | Where-Object { $_.CommandLine -like '*watchdog*' } | Select-Object -ExpandProperty Id")
echo It will keep the orchestrator running 24/7.
echo To stop: Open Task Manager and end the PowerShell process running watchdog.ps1
pause

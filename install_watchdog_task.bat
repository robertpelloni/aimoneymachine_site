@echo off
echo Installing AI Hustle Machine Watchdog as Windows Scheduled Task...
echo.

cd /d "%~dp0"

:: Create an XML task definition
set TASK_NAME=AIHustleWatchdog
set SCRIPT="%CD%\watchdog.ps1"

:: Delete existing task if present
schtasks /delete /tn "%TASK_NAME%" /f 2>nul

:: Create the task - runs every 5 minutes, checks if orchestrator is running
schtasks /create /tn "%TASK_NAME%" /tr "powershell.exe -WindowStyle Hidden -ExecutionPolicy Bypass -File %SCRIPT%" /sc onstart /delay 0000:00 /f

echo.
echo Task created! It will start the watchdog automatically on boot.
echo.
echo To run it now:
echo   schtasks /run /tn "%TASK_NAME%"
echo.
echo To stop it:
echo   schtasks /end /tn "%TASK_NAME%"
echo.
pause

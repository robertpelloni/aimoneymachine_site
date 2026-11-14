# Build the tray controller (Windows .exe) into bin\.
# Below-normal priority so a build never wedges the machine.
(Get-Process -Id $PID).PriorityClass = 'BelowNormal'
Push-Location $PSScriptRoot\..\tray
go build -o ..\bin\aimm-tray.exe .
Pop-Location
Write-Host "Built bin\aimm-tray.exe"

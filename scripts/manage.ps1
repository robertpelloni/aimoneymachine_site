# Manage local AI Hustle Machine processes safely on Windows (BelowNormal priority, exact-PID tracking).
# Usage: .\scripts\manage.ps1 <status|start-local|stop-local|start-tray|stop-tray|smoke>
param(
    [ValidateSet('status','start-local','stop-local','start-tray','stop-tray','smoke')]
    [string]$Action = 'status'
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$localPidFile = Join-Path $repoRoot '.aimm-local.pid'
$trayPidFile  = Join-Path $repoRoot '.aimm-tray.pid'

switch ($Action) {
    'status' {
        try {
            $resp = Invoke-RestMethod -Uri 'http://localhost:8082/api/status' -TimeoutSec 4
            Write-Host "Orchestrator :8082 => $($resp.status) v$($resp.version) | Profit: `$$($resp.profit) | Disk: $($resp.disk_used_pct)%"
        } catch {
            Write-Host "Orchestrator :8082 => unreachable ($($_.Exception.Message))"
        }
        $trays = Get-Process -Name 'aimm-tray','tray' -ErrorAction SilentlyContinue
        if ($trays) {
            Write-Host ("Tray Icon => RUNNING (PID: " + (($trays | Select-Object -ExpandProperty Id) -join ', ') + ")")
        } else {
            Write-Host "Tray Icon => not running"
        }
    }
    'start-local' {
        $bin = Join-Path $repoRoot 'bin\orchestrator.exe'
        if (-not (Test-Path $bin)) { throw "Missing $bin - run .\scripts\build-all.ps1 first." }
        $proc = Start-Process -FilePath $bin -ArgumentList '-api','8082','-daemon','-dry-run' -WorkingDirectory $repoRoot -WindowStyle Hidden -PassThru
        $proc.PriorityClass = 'BelowNormal'
        Set-Content -Path $localPidFile -Value $proc.Id
        Write-Host "Started local orchestrator at BelowNormal priority (PID $($proc.Id)) on http://localhost:8082"
    }
    'stop-local' {
        if (Test-Path $localPidFile) {
            $targetPid = [int](Get-Content $localPidFile).Trim()
            $p = Get-Process -Id $targetPid -ErrorAction SilentlyContinue
            if ($p -and $p.ProcessName -like '*orchestrator*') {
                Stop-Process -Id $targetPid -Force
                Write-Host "Stopped local orchestrator PID $targetPid"
            }
            Remove-Item $localPidFile -Force -ErrorAction SilentlyContinue
        } else {
            Write-Host "No .aimm-local.pid file present."
        }
    }
    'start-tray' {
        $bin = Join-Path $repoRoot 'bin\aimm-tray.exe'
        if (-not (Test-Path $bin)) { throw "Missing $bin - run .\scripts\build-tray.ps1 first." }
        $proc = Start-Process -FilePath $bin -WorkingDirectory $repoRoot -PassThru
        try { $proc.PriorityClass = 'BelowNormal' } catch {}
        Set-Content -Path $trayPidFile -Value $proc.Id
        Write-Host "Started aimm-tray.exe (PID $($proc.Id))"
    }
    'stop-tray' {
        if (Test-Path $trayPidFile) {
            $targetPid = [int](Get-Content $trayPidFile).Trim()
            $p = Get-Process -Id $targetPid -ErrorAction SilentlyContinue
            if ($p -and ($p.ProcessName -eq 'aimm-tray' -or $p.ProcessName -eq 'tray')) {
                Stop-Process -Id $targetPid -Force
                Write-Host "Stopped aimm-tray PID $targetPid"
            }
            Remove-Item $trayPidFile -Force -ErrorAction SilentlyContinue
        }
    }
    'smoke' {
        python (Join-Path $PSScriptRoot 'smoke_prod.py') 'http://localhost:8082'
    }
}

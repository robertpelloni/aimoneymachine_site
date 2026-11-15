# Hetzner-live-only status & optional remote tray helper on Windows (zero local servers).
# Usage: .\scripts\manage.ps1 <status|smoke|start-tray|stop-tray>
param(
    [ValidateSet('status','smoke','start-tray','stop-tray')]
    [string]$Action = 'status'
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$trayPidFile = Join-Path $repoRoot '.aimm-tray.pid'
$liveUrl = if ($env:AIMM_API_URL) { $env:AIMM_API_URL.TrimEnd('/') } else { 'https://aimoneymachine.site' }

switch ($Action) {
    'status' {
        try {
            $headers = @{ 'User-Agent' = 'Mozilla/5.0 (compatible; AIMM-Manage/2.0.1; +https://aimoneymachine.site)' }
            $resp = Invoke-RestMethod -Uri "$liveUrl/api/status" -Headers $headers -TimeoutSec 8
            Write-Host "Hetzner Live ($liveUrl) => $($resp.status) v$($resp.version) | Profit: `$$($resp.profit) | Disk: $($resp.disk_used_pct)%"
        } catch {
            Write-Host "Hetzner Live ($liveUrl) => unreachable ($($_.Exception.Message))"
        }
        $trays = Get-Process -Name 'aimm-tray','tray' -ErrorAction SilentlyContinue
        if ($trays) {
            Write-Host ("Remote Tray Icon => RUNNING (PID: " + (($trays | Select-Object -ExpandProperty Id) -join ', ') + ")")
        } else {
            Write-Host "Remote Tray Icon => not running (zero local processes)"
        }
    }
    'start-tray' {
        $bin = Join-Path $repoRoot 'bin\aimm-tray.exe'
        if (-not (Test-Path $bin)) { throw "Missing $bin - run .\scripts\build-tray.ps1 first." }
        $proc = Start-Process -FilePath $bin -WorkingDirectory $repoRoot -PassThru
        try { $proc.PriorityClass = 'BelowNormal' } catch {}
        Set-Content -Path $trayPidFile -Value $proc.Id
        Write-Host "Started remote-only aimm-tray.exe (PID $($proc.Id)) -> $liveUrl/dashboard"
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
        python (Join-Path $PSScriptRoot 'smoke_prod.py') $liveUrl
    }
}

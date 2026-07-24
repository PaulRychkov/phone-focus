$ErrorActionPreference = "SilentlyContinue"

$key = "C:\Users\pavel\.ssh\tt-k3s"
$cloudflared = "C:\Users\pavel\cloudflared\cloudflared.exe"
$config = "C:\Users\pavel\.cloudflared\phone-focus.yml"
$env:TUNNEL_ORIGIN_CERT = "C:\Users\pavel\.cloudflared\cert.pem"
$log = "$env:LOCALAPPDATA\phone-focus-relay.log"
$sshLog = "$env:LOCALAPPDATA\phone-focus-ssh.log"
$readyUrl = "http://127.0.0.1:20241/ready"

function Log($m) {
    Add-Content -Path $log -Value "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $m"
}

function Rotate($path) {
    $f = Get-Item $path
    if ($f -and $f.Length -gt 1MB) {
        $tail = Get-Content $path -Tail 300
        Set-Content -Path $path -Value $tail
    }
}

function Test-TunnelDeep {
    try { return (Invoke-WebRequest -Uri "http://127.0.0.1:18083/healthz" -TimeoutSec 5 -UseBasicParsing).StatusCode -eq 200 }
    catch { return $false }
}

function Stop-Tunnel {
    $pidFile = "$env:LOCALAPPDATA\phone-focus-tunnel.pid"
    if (Test-Path $pidFile) {
        $tpid = [int](Get-Content $pidFile)
        $p = Get-Process -Id $tpid -ErrorAction SilentlyContinue
        if ($p -and $p.ProcessName -eq 'ssh') { $p | Stop-Process -Force }
        Remove-Item $pidFile -Force
    }
    $owners = @(netstat -ano | ForEach-Object {
        if ($_ -match 'TCP\s+127\.0\.0\.1:(18083|9095)\s.*LISTENING\s+(\d+)') { $Matches[2] }
    } | Select-Object -Unique)
    foreach ($id in $owners) {
        $p = Get-Process -Id $id -ErrorAction SilentlyContinue
        if ($p -and $p.ProcessName -eq 'ssh') { $p | Stop-Process -Force }
    }
}

function Start-Tunnel {
    Stop-Tunnel
    Start-Sleep -Seconds 1
    Rotate $sshLog
    $p = Start-Process ssh -WindowStyle Hidden -PassThru -ArgumentList @(
        "-i", $key,
        "-o", "StrictHostKeyChecking=no",
        "-o", "BatchMode=yes",
        "-o", "ServerAliveInterval=20",
        "-o", "ServerAliveCountMax=3",
        "-o", "ExitOnForwardFailure=yes",
        "-E", $sshLog,
        "-N",
        "-L", "127.0.0.1:18083:localhost:8083",
        "-L", "127.0.0.1:9095:localhost:9095",
        "tt@tt-k3s"
    )
    Set-Content -Path "$env:LOCALAPPDATA\phone-focus-tunnel.pid" -Value $p.Id
    Log "ssh: туннель запущен (pid $($p.Id))"
}

function Sync-VmIp {
    $cfg = "$env:USERPROFILE\.ssh\config"
    $cands = @(arp -a | ForEach-Object {
        if ($_ -match '^\s*(192\.168\.\d+\.\d+)\s+00-15-5d') { $Matches[1] }
    } | Select-Object -Unique)
    Log "sync-vmip: кандидаты ARP: $($cands -join ', ')"
    $vmIp = $null
    $probeOut = "$env:LOCALAPPDATA\phone-focus-probe.txt"
    foreach ($ip in $cands) {
        Remove-Item $probeOut -Force -ErrorAction SilentlyContinue
        $probe = Start-Process ssh -WindowStyle Hidden -PassThru -RedirectStandardOutput $probeOut -ArgumentList @(
            "-i", $key, "-o", "StrictHostKeyChecking=no", "-o", "BatchMode=yes", "-o", "ConnectTimeout=6", "tt@$ip", "hostname"
        )
        if (-not $probe.WaitForExit(12000)) {
            $probe.Kill()
            Log "sync-vmip: проба $ip повисла, убита"
            continue
        }
        $h = Get-Content $probeOut -ErrorAction SilentlyContinue
        if ("$h" -match "tt-k3s") { $vmIp = $ip; break }
    }
    if (-not $vmIp) { Log "sync-vmip: VM не найдена, конфиг не трогаю"; return }
    $cur = ""
    if (Test-Path $cfg) {
        $m = [regex]::Match((Get-Content $cfg -Raw), "(?ms)^Host\s+tt-k3s\b.*?HostName\s+(\S+)")
        if ($m.Success) { $cur = $m.Groups[1].Value }
    }
    if ($cur -eq $vmIp) { Log "sync-vmip: IP актуален ($vmIp)"; return }
    $block = "Host tt-k3s`n    HostName $vmIp`n    User tt`n    IdentityFile ~/.ssh/tt-k3s`n    IdentitiesOnly yes`n    StrictHostKeyChecking no`n"
    if (Test-Path $cfg) {
        $t = [regex]::Replace((Get-Content $cfg -Raw), "(?ms)^Host\s+tt-k3s\b.*?(?=^Host\s|\z)", "")
        Set-Content -Path $cfg -Value ($t.TrimEnd() + "`n`n" + $block) -Encoding ASCII
    } else {
        Set-Content -Path $cfg -Value $block -Encoding ASCII
    }
    Log "sync-vmip: IP обновлён $cur -> $vmIp"
}

function Start-CF($killFirst) {
    if ($killFirst) {
        Get-Process cloudflared | Stop-Process -Force
        Start-Sleep -Seconds 2
    }
    Start-Process $cloudflared -WindowStyle Hidden -ArgumentList @(
        "--config", $config, "tunnel", "run", "phone-focus"
    )
    Log "cloudflared запущен (killFirst=$killFirst)"
}

function Test-Ready {
    try { return (Invoke-WebRequest -Uri $readyUrl -TimeoutSec 4 -UseBasicParsing).StatusCode -eq 200 }
    catch { return $false }
}

Rotate $log
Log "=== релей запущен (pid $PID) ==="

$cfFail = 0
while ($true) {
    if (-not (Test-TunnelDeep)) {
        Log "туннель нездоров, пересоздаю"
        Sync-VmIp
        Start-Tunnel
        Start-Sleep -Seconds 6
        Log "туннель после рестарта здоров: $(Test-TunnelDeep)"
    }

    if (-not (Get-Process cloudflared)) {
        Start-CF $false
        $cfFail = 0
    } elseif (Test-Ready) {
        $cfFail = 0
    } else {
        $cfFail++
        if ($cfFail -ge 4) {
            Log "cloudflared не ready $cfFail циклов, перезапускаю"
            Start-CF $true
            $cfFail = 0
        }
    }

    Start-Sleep -Seconds 20
}

# 构建「单 exe 分发版」：内嵌 web + 语音模型 + Vosk DLL + 默认配置
# 用法: .\scripts\build-standalone.ps1

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $Root

& (Join-Path $PSScriptRoot "setup-vosk.ps1")
& (Join-Path $PSScriptRoot "prepare-bundle.ps1")

$binDir = Join-Path $Root "third_party\vosk\windows-amd64\bin"
$payloadDir = Join-Path $Root "cmd\launcher\payload"
New-Item -ItemType Directory -Force -Path $payloadDir | Out-Null
Get-ChildItem $payloadDir -File | Remove-Item -Force

$coreOut = Join-Path $payloadDir "speaknow-core.exe"
$env:CGO_ENABLED = "1"
go build -ldflags="-s -w" -o $coreOut ./cmd/server
Copy-Item "$binDir\*.dll" $payloadDir -Force

$launcherOut = Join-Path $Root "speaknow.exe"
go build -ldflags="-s -w" -o $launcherOut ./cmd/launcher

$totalMB = [math]::Round((Get-Item $launcherOut).Length / 1MB, 1)
Write-Host ""
Write-Host "Done: speaknow.exe ($totalMB MB)"
Write-Host "Distribute ONLY this file. Config is embedded; first run extracts model to .speaknow-data"
Write-Host "Open http://127.0.0.1:8080/web/"

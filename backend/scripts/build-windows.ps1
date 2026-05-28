# 构建 Windows 版：根目录生成 speaknow.exe（内嵌 speaknow-core + Vosk DLL）
# 用法: .\scripts\build-windows.ps1

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $Root

$binDir = Join-Path $Root "third_party\vosk\windows-amd64\bin"
$embedDir = Join-Path $Root "internal\voskruntime\dll"
$payloadDir = Join-Path $Root "cmd\launcher\payload"

if (-not (Test-Path (Join-Path $binDir "libvosk.dll")) -or
    -not (Test-Path (Join-Path $embedDir "libvosk.dll"))) {
    Write-Host "Vosk libs missing, running setup-vosk.ps1 ..."
    & (Join-Path $PSScriptRoot "setup-vosk.ps1")
}

Copy-Item (Join-Path $Root "configs\config.release.yaml") (Join-Path $Root "internal\assets\default.yaml") -Force

New-Item -ItemType Directory -Force -Path $payloadDir | Out-Null
Get-ChildItem $payloadDir -File | Remove-Item -Force

$coreOut = Join-Path $payloadDir "speaknow-core.exe"
$env:CGO_ENABLED = "1"
go build -o $coreOut ./cmd/server
Copy-Item "$binDir\*.dll" $payloadDir -Force

$launcherOut = Join-Path $Root "..\speaknow.exe"
go build -o $launcherOut ./cmd/launcher

$coreMB = [math]::Round((Get-Item $coreOut).Length / 1MB, 1)
$totalMB = [math]::Round((Get-Item $launcherOut).Length / 1MB, 1)
Write-Host ""
Write-Host "Built: $launcherOut ($totalMB MB, embeds core $coreMB MB + Vosk DLL)"
Write-Host "Run:  double-click speaknow.exe  (or: .\speaknow.exe -config path\to\custom.yaml)"
Write-Host "First run extracts runtime to .speaknow-data\runtime\ (no extra DLL beside speaknow.exe)."
if (Test-Path (Join-Path $Root "model\vosk-model-small-cn-0.22")) {
    Write-Host "Model: model\vosk-model-small-cn-0.22"
} else {
    Write-Host "Tip: model\vosk-model-small-cn-0.22 for offline Vosk, or build-standalone.ps1 to embed model."
}

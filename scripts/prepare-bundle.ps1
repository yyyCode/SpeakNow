# 将 web、语音模型复制到 internal/assets，供 go:embed 打进单文件 exe
# 用法: .\scripts\prepare-bundle.ps1

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$Assets = Join-Path $Root "internal\assets"

$webSrc = Join-Path $Root "web"
$modelSrc = Join-Path $Root "model\vosk-model-small-cn-0.22"
$webDst = Join-Path $Assets "web"
$modelDst = Join-Path $Assets "model\vosk-model-small-cn-0.22"

if (-not (Test-Path $webSrc)) { throw "missing $webSrc" }
if (-not (Test-Path (Join-Path $modelSrc "conf\model.conf"))) {
    throw "missing vosk model at $modelSrc"
}

if (Test-Path $webDst) { Remove-Item $webDst -Recurse -Force }
if (Test-Path (Join-Path $Assets "model")) { Remove-Item (Join-Path $Assets "model") -Recurse -Force }

New-Item -ItemType Directory -Force -Path (Split-Path $modelDst) | Out-Null
Copy-Item $webSrc $webDst -Recurse -Force
Copy-Item $modelSrc $modelDst -Recurse -Force
Copy-Item (Join-Path $Root "configs\config.release.yaml") (Join-Path $Assets "default.yaml") -Force

$mb = [math]::Round((Get-ChildItem $Assets -Recurse | Measure-Object Length -Sum).Sum / 1MB, 1)
Write-Host "Bundle prepared under internal/assets ($mb MB). Run: go build -o speaknow.exe ./cmd/server"

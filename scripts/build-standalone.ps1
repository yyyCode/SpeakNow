# 构建「单 exe 分发版」：内嵌 web + 语音模型 + Vosk DLL
# 用法: .\scripts\build-standalone.ps1

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $Root

& (Join-Path $PSScriptRoot "setup-vosk.ps1")
& (Join-Path $PSScriptRoot "prepare-bundle.ps1")

$env:CGO_ENABLED = "1"
go build -ldflags="-s -w" -o speaknow.exe ./cmd/server

$size = [math]::Round((Get-Item speaknow.exe).Length / 1MB, 1)
Write-Host ""
Write-Host "Done: speaknow.exe ($size MB)"
Write-Host "Distribute ONLY this file. First run extracts model to .speaknow-data beside the exe."
Write-Host "Open http://127.0.0.1:8080/web/"

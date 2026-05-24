# 构建 Windows 版：先确保 Vosk 库就绪，再编译并组装 dist 目录
# 用法: .\scripts\build-windows.ps1

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $Root

$binDir = Join-Path $Root "third_party\vosk\windows-amd64\bin"
$embedDir = Join-Path $Root "internal\voskruntime\dll"

# 同步 bin 与 go:embed 目录（缺任一即重新 setup）
if (-not (Test-Path (Join-Path $binDir "libvosk.dll")) -or
    -not (Test-Path (Join-Path $embedDir "libvosk.dll"))) {
    Write-Host "Vosk libs missing, running setup-vosk.ps1 ..."
    & (Join-Path $PSScriptRoot "setup-vosk.ps1")
}

$dist = Join-Path $Root "dist\speaknow"
New-Item -ItemType Directory -Force -Path $dist | Out-Null

$env:CGO_ENABLED = "1"
$exeOut = Join-Path $dist "speaknow.exe"
go build -o $exeOut ./cmd/server

Copy-Item "$binDir\*.dll" $dist -Force
Copy-Item $exeOut (Join-Path $Root "speaknow.exe") -Force
Copy-Item -Recurse -Force (Join-Path $Root "web") (Join-Path $dist "web")
if (Test-Path (Join-Path $Root "model")) {
    New-Item -ItemType Directory -Force -Path (Join-Path $dist "model") | Out-Null
    Write-Host "Tip: copy model/vosk-model-small-cn-0.22 to dist\speaknow\model\ for offline ASR"
}

Write-Host "Built: $dist"
Write-Host "Run (dist):  cd dist\speaknow ; .\speaknow.exe -config ..\..\configs\config.yaml"
Write-Host "Run (root):  .\speaknow.exe -config configs\config.yaml"

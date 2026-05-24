# 从官方 release 或项目内 zip 解压 Vosk Windows 库到 third_party/vosk
# 官方: https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-win64-0.3.45.zip
# 用法: .\scripts\setup-vosk.ps1 [-Version 0.3.45] [-ZipPath "third_party\vosk-win64.zip"]

param(
    [string]$Version = "0.3.45",
    [string]$ZipPath = ""
)

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

$Base = Join-Path $Root "third_party\vosk\windows-amd64"
$Include = Join-Path $Base "include"
$Lib = Join-Path $Base "lib"
$Bin = Join-Path $Base "bin"
$EmbedDLL = Join-Path $Root "internal\voskruntime\dll"

New-Item -ItemType Directory -Force -Path $Include, $Lib, $Bin, $EmbedDLL | Out-Null

$ZipName = "vosk-win64-$Version.zip"
if (-not $ZipPath) {
    $ZipPath = Join-Path $Root "third_party\$ZipName"
}
if (-not (Test-Path $ZipPath)) {
    $ZipPath = Join-Path $env:TEMP $ZipName
    $Url = "https://github.com/alphacep/vosk-api/releases/download/v$Version/$ZipName"
    Write-Host "Downloading $Url ..."
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath -UseBasicParsing
} else {
    Write-Host "Using local zip: $ZipPath"
}

$ExtractDir = Join-Path $env:TEMP "vosk-win64-extract-$Version"
if (Test-Path $ExtractDir) { Remove-Item $ExtractDir -Recurse -Force }
Expand-Archive -Path $ZipPath -DestinationPath $ExtractDir -Force

$SrcDir = Get-ChildItem $ExtractDir -Recurse -File -Filter "vosk_api.h" | Select-Object -First 1
if ($SrcDir) { $SrcDir = $SrcDir.DirectoryName } else { $SrcDir = $ExtractDir }

Copy-Item (Join-Path $SrcDir "vosk_api.h") $Include -Force
if (Test-Path (Join-Path $SrcDir "libvosk.lib")) {
    Copy-Item (Join-Path $SrcDir "libvosk.lib") $Lib -Force
}
Get-ChildItem $SrcDir -Filter "*.dll" | ForEach-Object {
    Copy-Item $_.FullName $Bin -Force
    Copy-Item $_.FullName $EmbedDLL -Force
}
# MinGW 链接需要 lib 目录下同时有 .lib 与 .dll
if (Test-Path (Join-Path $Bin "libvosk.dll")) {
    Copy-Item (Join-Path $Bin "libvosk.dll") $Lib -Force
}

Write-Host "OK -> $Base"
Write-Host "OK -> $EmbedDLL (for go:embed build)"
Get-ChildItem $Bin | Select-Object Name, Length

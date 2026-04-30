# icoo_assistant build script
param(
    [string]$Output = "bin/assistant.exe",
    [string]$Version = "",
    [switch]$Clean,
    [switch]$Install
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

if ($Version -eq "") {
    $Version = (Get-Date -Format "0.1.yyMMdd.HHmm")
}
$LDFLAGS = "-s -w -X main.Version=$Version"

Write-Host "=== icoo_assistant build ===" -ForegroundColor Cyan
Write-Host "Version : $Version"
Write-Host "Output  : $Output"
Write-Host "LDFLAGS : $LDFLAGS"
Write-Host ""

if ($Clean) {
    Write-Host "[clean] removing bin/" -ForegroundColor Yellow
    Remove-Item -Recurse -Force bin/ -ErrorAction SilentlyContinue
}

Write-Host "[build] go build -ldflags `"$LDFLAGS`" -o $Output ./cmd/assistant" -ForegroundColor Green
$outDir = Split-Path $Output -Parent
if ($outDir -ne "") {
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null
}

go build -ldflags "$LDFLAGS" -o $Output ./cmd/assistant

if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] build failed" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "[OK] $Output" -ForegroundColor Green

if ($Install) {
    $target = Join-Path $env:GOPATH "bin/assistant.exe"
    Copy-Item -Force $Output $target
    Write-Host "[install] $target" -ForegroundColor Green
}

Write-Host ""
Write-Host "Run: $Output check" -ForegroundColor Cyan

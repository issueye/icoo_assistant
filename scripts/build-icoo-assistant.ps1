param(
    [string]$TargetOS = "windows",
    [string]$TargetArch = "amd64",
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptDir
$projectRoot = Join-Path $repoRoot "icoo_assistant"
$distRoot = Join-Path $repoRoot "dist"
$versionFile = Join-Path $projectRoot "cmd\assistant\version.go"
$envExample = Join-Path $projectRoot ".env.example"

if (-not (Test-Path -LiteralPath $projectRoot)) {
    throw "project root not found: $projectRoot"
}

if (-not (Test-Path -LiteralPath $versionFile)) {
    throw "version file not found: $versionFile"
}

$versionContent = Get-Content -LiteralPath $versionFile -Raw
$versionMatch = [regex]::Match($versionContent, 'const\s+Version\s*=\s*"([^"]+)"')
if (-not $versionMatch.Success) {
    throw "failed to parse version from $versionFile"
}

$version = $versionMatch.Groups[1].Value
$packageName = "icoo_assistant-v$version-$TargetOS-$TargetArch"
$outputDir = Join-Path $distRoot $packageName
$binaryName = if ($TargetOS -eq "windows") { "assistant.exe" } else { "assistant" }
$binaryPath = Join-Path $outputDir $binaryName
$zipPath = Join-Path $distRoot ($packageName + ".zip")
$envPath = Join-Path $outputDir ".env"

Write-Host "==> building $packageName"

$pushedLocation = $false
try {
    Push-Location $projectRoot
    $pushedLocation = $true

    if (-not $SkipTests) {
        Write-Host "==> running tests"
        & go test ./... | Out-Host
        if ($LASTEXITCODE -ne 0) {
            throw "go test failed"
        }
    }

    if (-not (Test-Path -LiteralPath $distRoot)) {
        New-Item -ItemType Directory -Path $distRoot | Out-Null
    }

    if (-not (Test-Path -LiteralPath $outputDir)) {
        New-Item -ItemType Directory -Path $outputDir | Out-Null
    }

    Write-Host "==> compiling binary"
    $env:GOOS = $TargetOS
    $env:GOARCH = $TargetArch
    if (Test-Path -LiteralPath $binaryPath) {
        Remove-Item -LiteralPath $binaryPath -Force
    }
    & go build -o $binaryPath ./cmd/assistant | Out-Host
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed"
    }

    if (Test-Path -LiteralPath $envExample) {
        Copy-Item -LiteralPath $envExample -Destination $envPath -Force
    }

    Write-Host "==> creating zip archive"
    if (Test-Path -LiteralPath $zipPath) {
        Remove-Item -LiteralPath $zipPath -Force
    }
    Compress-Archive -LiteralPath $outputDir -DestinationPath $zipPath
} finally {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    if ($pushedLocation) {
        Pop-Location
    }
}

Write-Host "==> done"
Write-Host "binary: $binaryPath"
Write-Host "archive: $zipPath"

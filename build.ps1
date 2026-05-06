$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$outputDir = Join-Path $scriptDir "icoo_runtime"
$outputPath = Join-Path $outputDir "icoo.exe"
$configSourcePath = Join-Path $scriptDir "config.toml.example"
$configTargetPath = Join-Path $outputDir "config.toml"
$runtimePackagePath = "./icoo_runtime/cmd/assistant"
$rootPackagePath = "./cmd/assistant"

$packagePath = $runtimePackagePath
if (-not (Test-Path (Join-Path $scriptDir "icoo_runtime\\cmd\\assistant"))) {
    $packagePath = $rootPackagePath
}

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

Push-Location $scriptDir
try {
    go build -o $outputPath $packagePath
    if (Test-Path $configSourcePath) {
        Copy-Item -Force $configSourcePath $configTargetPath
    }
}
finally {
    Pop-Location
}

Write-Host "Built $outputPath from $packagePath"
if (Test-Path $configTargetPath) {
    Write-Host "Copied $configTargetPath from $configSourcePath"
}

param(
  [switch]$SkipTests,
  [switch]$SkipFrontendInstall,
  [switch]$WailsBuild
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$FrontendDir = Join-Path $ProjectRoot "frontend"
$BuildDir = Join-Path $ProjectRoot "build"
$OutputDir = Join-Path $BuildDir "bin"
$OutputFile = Join-Path $OutputDir "icoo_proxy.exe"

function Write-Step {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Message
  )

  Write-Host ""
  Write-Host "==> $Message" -ForegroundColor Cyan
}

function Assert-Command {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Name
  )

  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "Required command not found: $Name"
  }
}

function Invoke-Checked {
  param(
    [Parameter(Mandatory = $true)]
    [scriptblock]$Script
  )

  & $Script
  if ($LASTEXITCODE -ne 0) {
    throw "Command failed with exit code $LASTEXITCODE."
  }
}

Assert-Command "go"
Assert-Command "npm"
if ($WailsBuild) {
  Assert-Command "wails"
}

if (-not (Test-Path $OutputDir)) {
  New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

if ($SkipFrontendInstall) {
  Write-Step "Skipping frontend install by request"
} else {
  $HasNodeModules = Test-Path (Join-Path $FrontendDir "node_modules")
  if (-not $HasNodeModules) {
    Write-Step "Installing frontend dependencies"
    Push-Location $FrontendDir
    try {
      if (Test-Path (Join-Path $FrontendDir "package-lock.json")) {
        Invoke-Checked { npm ci }
      } else {
        Invoke-Checked { npm install }
      }
    } finally {
      Pop-Location
    }
  } else {
    Write-Step "Skipping frontend install because node_modules already exists"
  }
}

if (-not $SkipTests) {
  Write-Step "Running Go tests"
  Push-Location $ProjectRoot
  try {
    Invoke-Checked { go test ./... }
  } finally {
    Pop-Location
  }
}

Write-Step "Building frontend assets"
Push-Location $FrontendDir
try {
  Invoke-Checked { npm run build }
} finally {
  Pop-Location
}

Write-Step "Building proxy executable"
Push-Location $ProjectRoot
try {
  Invoke-Checked { go build -o $OutputFile . }
} finally {
  Pop-Location
}

if ($WailsBuild) {
  Write-Step "Building Wails desktop application"
  Push-Location $ProjectRoot
  try {
    Invoke-Checked { wails build }
  } finally {
    Pop-Location
  }
}

Write-Step "Build completed"
Write-Host "Output: $OutputFile" -ForegroundColor Green
if ($WailsBuild) {
  Write-Host "Wails output: $OutputDir" -ForegroundColor Green
}

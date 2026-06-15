# Locate-or-download the skills-telemetry binary into LOCALAPPDATA, then run it
# forwarding all args and stdin. Windows PowerShell 5.1 built-ins only.
$ErrorActionPreference = 'Stop'

# BinaryVersion is the release tag; BaseUrl/<tag>/<asset> is the GitHub
# Releases download layout, so the URL below resolves to a real asset.
$BinaryVersion = 'v0.1.1'
$BaseUrl = 'https://github.com/denifilatoff/skills-telemetry/releases/download'

$cacheBase = $env:LOCALAPPDATA
$arch = if ([Environment]::Is64BitOperatingSystem) { 'amd64' } else { 'amd64' }
$dir = Join-Path $cacheBase "qubership-skills-telemetry\bin\$BinaryVersion"
$bin = Join-Path $dir "skills-telemetry-windows-$arch.exe"

try {
  if (-not (Test-Path $bin)) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    $tmp = "$bin.tmp"
    Invoke-WebRequest -UseBasicParsing -Uri "$BaseUrl/$BinaryVersion/skills-telemetry-windows-$arch.exe" -OutFile $tmp
    Move-Item -Force $tmp $bin
  }
  # Forward stdin and all args.
  $input | & $bin @args
  exit $LASTEXITCODE
} catch {
  Write-Error "skills-telemetry bootstrap failed: $_"
  exit 0  # never fail the agent turn
}

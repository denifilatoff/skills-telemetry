# Locate-or-download the skills-telemetry binary into LOCALAPPDATA, then run it
# forwarding all args and stdin. Windows PowerShell 5.1 built-ins only.
#
# Two callers share this script:
#   - the hook, as `bootstrap.ps1 ingest --agent=...` — must never fail the
#     agent turn, so bootstrap's own errors exit 0;
#   - the provisioning one-liner, as `... } provision` — interactive, so
#     bootstrap's own errors must surface with a non-zero exit.
$ErrorActionPreference = 'Stop'

# BinaryVersion is the release tag; BaseUrl/<tag>/<asset> is the GitHub
# Releases download layout, so the URL below resolves to a real asset.
$BinaryVersion = 'v0.4.0'
$BaseUrl = 'https://github.com/denifilatoff/skills-telemetry/releases/download'

$cmd = if ($args.Count -gt 0) { $args[0] } else { '' }

$cacheBase = $env:LOCALAPPDATA
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
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
  # The hook (ingest) must never fail the agent turn; interactive commands surface it.
  if ($cmd -eq 'ingest') { exit 0 } else { exit 1 }
}

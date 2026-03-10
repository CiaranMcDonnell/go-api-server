# Usage: .\benchmark\scripts\run.ps1 <scenario> [profile] [-Dashboard]
#   scenario: health | auth-flow | login-sustained
#   profile:  smoke | load | stress | spike | breakpoint  (default: load)
#
# Examples:
#   .\benchmark\scripts\run.ps1 health smoke
#   .\benchmark\scripts\run.ps1 auth-flow stress
#   .\benchmark\scripts\run.ps1 auth-flow load -Dashboard

param(
    [Parameter(Mandatory=$true)][string]$Scenario,
    [string]$Profile = "load",
    [switch]$Dashboard
)

$BaseUrl = if ($env:BASE_URL) { $env:BASE_URL } else { "http://localhost:8080" }
$ResultsDir = "benchmark\results"

if (!(Test-Path $ResultsDir)) { New-Item -ItemType Directory -Path $ResultsDir | Out-Null }

$Timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$OutputFile = "$ResultsDir\${Scenario}_${Profile}_${Timestamp}.json"

Write-Host "=== Benchmark: $Scenario | Profile: $Profile ==="
Write-Host "    Target: $BaseUrl"
Write-Host "    Output: $OutputFile"
if ($Dashboard) { Write-Host "    Dashboard: http://localhost:5665" }
Write-Host ""

$outArgs = @("--out", "json=$OutputFile")
if ($Dashboard) { $outArgs += @("--out", "web-dashboard") }

k6 run `
    -e PROFILE="$Profile" `
    -e BASE_URL="$BaseUrl" `
    @outArgs `
    "benchmark/scenarios/${Scenario}.js"

# Runs all scenarios sequentially with the given profile.
# Usage: .\benchmark\scripts\run-all.ps1 [profile]
#
# Examples:
#   .\benchmark\scripts\run-all.ps1 smoke
#   .\benchmark\scripts\run-all.ps1 stress

param(
    [string]$Profile = "load"
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "=== Running all benchmarks with profile: $Profile ==="
Write-Host ""

foreach ($Scenario in @("health", "auth-flow", "login-sustained")) {
    & "$ScriptDir\run.ps1" -Scenario $Scenario -Profile $Profile
    Write-Host ""
    Write-Host "--- Cooldown (10s) ---"
    Start-Sleep -Seconds 10
}

Write-Host "=== All benchmarks complete ==="

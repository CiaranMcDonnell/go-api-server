# Extracts key metrics from a k6 JSON output file and appends to results.csv
#
# Usage:
#   .\benchmark\scripts\extract.ps1 benchmark\results\auth-flow_smoke_20260310_053508.json
#   .\benchmark\scripts\extract.ps1 benchmark\results\auth-flow_smoke_20260310_053508.json -Label "baseline"
#
# Or extract all results in the folder:
#   Get-ChildItem benchmark\results\*.json | ForEach-Object { .\benchmark\scripts\extract.ps1 $_.FullName }

param(
    [Parameter(Mandatory=$true)][string]$File,
    [string]$Label = ""
)

$CsvPath = "benchmark\results.csv"

# Parse the k6 JSON output (each line is a separate JSON object)
$lines = Get-Content $File

# Extract summary metrics from "Point" type entries
$metrics = @{}
foreach ($line in $lines) {
    try {
        $obj = $line | ConvertFrom-Json
        if ($obj.type -eq "Point" -and $obj.metric) {
            $name = $obj.metric
            $value = $obj.data.value
            if (-not $metrics.ContainsKey($name)) {
                $metrics[$name] = [System.Collections.Generic.List[double]]::new()
            }
            $metrics[$name].Add($value)
        }
    } catch { continue }
}

# Calculate percentiles
function Get-Percentile($values, $p) {
    if ($values.Count -eq 0) { return 0 }
    $sorted = $values | Sort-Object
    $index = [math]::Ceiling($p / 100.0 * $sorted.Count) - 1
    if ($index -lt 0) { $index = 0 }
    return [math]::Round($sorted[$index], 2)
}

function Get-Avg($values) {
    if ($values.Count -eq 0) { return 0 }
    return [math]::Round(($values | Measure-Object -Average).Average, 2)
}

# Extract filename parts for scenario/profile
$basename = [System.IO.Path]::GetFileNameWithoutExtension($File)
$parts = $basename -split "_"
$scenario = $parts[0..($parts.Count-3)] -join "_"
$profile = $parts[$parts.Count-2]
$timestamp = $parts[$parts.Count-1]

# Build row
$row = [PSCustomObject]@{
    timestamp     = $timestamp
    scenario      = $scenario
    profile       = $profile
    label         = $Label
    reqs_total    = if ($metrics.ContainsKey("http_reqs")) { $metrics["http_reqs"].Count } else { 0 }
    duration_avg  = if ($metrics.ContainsKey("http_req_duration")) { Get-Avg $metrics["http_req_duration"] } else { 0 }
    duration_p95  = if ($metrics.ContainsKey("http_req_duration")) { Get-Percentile $metrics["http_req_duration"] 95 } else { 0 }
    duration_p99  = if ($metrics.ContainsKey("http_req_duration")) { Get-Percentile $metrics["http_req_duration"] 99 } else { 0 }
    errors        = if ($metrics.ContainsKey("errors")) { $metrics["errors"].Count } else { 0 }
}

# Append to CSV
$exists = Test-Path $CsvPath
$row | Export-Csv -Path $CsvPath -Append -NoTypeInformation -Force

if ($exists) {
    Write-Host "Appended to $CsvPath"
} else {
    Write-Host "Created $CsvPath"
}

Write-Host ""
Write-Host "  scenario:    $scenario"
Write-Host "  profile:     $profile"
Write-Host "  label:       $Label"
Write-Host "  total reqs:  $($row.reqs_total)"
Write-Host "  avg (ms):    $($row.duration_avg)"
Write-Host "  p95 (ms):    $($row.duration_p95)"
Write-Host "  p99 (ms):    $($row.duration_p99)"
Write-Host "  errors:      $($row.errors)"

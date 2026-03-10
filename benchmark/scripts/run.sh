#!/usr/bin/env bash
set -euo pipefail

# Usage: ./benchmark/scripts/run.sh <scenario> [profile]
#   scenario: health | auth-flow | login-sustained
#   profile:  smoke | load | stress | spike  (default: load)
#
# Examples:
#   ./benchmark/scripts/run.sh health smoke
#   ./benchmark/scripts/run.sh auth-flow stress
#   ./benchmark/scripts/run.sh login-sustained load

# Find k6 binary
K6="$(command -v k6 2>/dev/null || echo "/c/Program Files/k6/k6.exe")"
if [ ! -x "$K6" ] && ! command -v k6 &>/dev/null; then
  echo "Error: k6 not found. Install with: winget install k6" >&2
  exit 1
fi

SCENARIO="${1:?Usage: run.sh <scenario> [profile]}"
PROFILE="${2:-load}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
RESULTS_DIR="benchmark/results"

mkdir -p "$RESULTS_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_FILE="$RESULTS_DIR/${SCENARIO}_${PROFILE}_${TIMESTAMP}.json"

echo "=== Benchmark: ${SCENARIO} | Profile: ${PROFILE} ==="
echo "    Target: ${BASE_URL}"
echo "    Output: ${OUTPUT_FILE}"
echo ""

"$K6" run \
  -e PROFILE="$PROFILE" \
  -e BASE_URL="$BASE_URL" \
  --out json="$OUTPUT_FILE" \
  "benchmark/scenarios/${SCENARIO}.js"

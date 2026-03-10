#!/usr/bin/env bash
set -euo pipefail

# Runs all scenarios sequentially with the given profile.
# Usage: ./benchmark/scripts/run-all.sh [profile]

PROFILE="${1:-load}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Running all benchmarks with profile: ${PROFILE} ==="
echo ""

for scenario in health auth-flow login-sustained; do
  "$SCRIPT_DIR/run.sh" "$scenario" "$PROFILE"
  echo ""
  echo "--- Cooldown (10s) ---"
  sleep 10
done

echo "=== All benchmarks complete ==="

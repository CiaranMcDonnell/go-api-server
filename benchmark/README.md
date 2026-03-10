# Benchmarks

Load tests using [k6](https://k6.io) to measure API throughput, latency, and breaking points.

## Setup

Install k6:
```
winget install k6
```

Start the server:
```
docker compose up --build
```

## Quick Start

```bash
# Run a single benchmark
node benchmark/bench.mjs run health smoke
node benchmark/bench.mjs run auth-flow stress --label opt-v3

# Run all scenarios with a profile
node benchmark/bench.mjs run-all stress --label opt-v3

# Extract existing k6 JSON files to CSV
node benchmark/bench.mjs extract "benchmark/results/*_stress_*.json" --label opt-v3

# Compare last two runs of a scenario/profile
node benchmark/bench.mjs compare auth-flow stress
```

Each `run` automatically: executes k6 → parses JSON output → appends to `results.csv` → prints a comparison against the previous run.

## Scenarios

| Scenario | File | What it tests |
|---|---|---|
| `health` | `scenarios/health.js` | Raw framework throughput, audit middleware overhead |
| `auth-flow` | `scenarios/auth-flow.js` | Full lifecycle: register → login → /me → logout |
| `login-sustained` | `scenarios/login-sustained.js` | Sustained login load (argon2id + JWT isolation) |

## Profiles

| Profile | Pattern | Use case |
|---|---|---|
| `smoke` | 5 VUs, 50s | Sanity check — does it work? |
| `load` | 50 VUs, 2m | Normal production load |
| `stress` | Ramp to 300 VUs, 6m | Find the breaking point |
| `spike` | Jump to 500 VUs | Sudden traffic burst |

## Results & Dashboard

Results are tracked in `results.csv` and documented in the [JOURNAL.md](JOURNAL.md).

Open `dashboard.html` in a browser and load `results.csv` to visualize latency, throughput, and errors across runs.

## Key Metrics

- **duration_avg** — Average request latency
- **duration_p95** — 95th percentile latency (target: <500ms)
- **duration_p99** — 99th percentile latency (target: <1000ms)
- **rps** — Requests per second throughput
- **errors** — Total failed checks

## Structure

```
benchmark/
├── bench.mjs               # CLI — run, extract, compare (Node.js)
├── config.js               # Shared profiles, thresholds, base URL
├── helpers.js              # Shared utilities (cookies, unique emails)
├── dashboard.html          # Chart.js visualisation (open in browser)
├── results.csv             # Accumulated metrics (committed)
├── JOURNAL.md              # Optimisation changelog
├── scenarios/
│   ├── health.js           # Baseline throughput
│   ├── auth-flow.js        # Full auth lifecycle
│   └── login-sustained.js  # Sustained login hammering
└── results/                # Raw k6 JSON output (gitignored)
```

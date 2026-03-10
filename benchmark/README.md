# Benchmarks

Load tests using [k6](https://k6.io) to measure API throughput, latency, and breaking points.

## Setup

Install k6:
```powershell
winget install k6
```

Start the server:
```bash
docker compose up --build
```

## Scenarios

| Scenario | File | What it tests |
|---|---|---|
| `health` | `scenarios/health.js` | Raw framework throughput, audit middleware overhead |
| `auth-flow` | `scenarios/auth-flow.js` | Full lifecycle: register → login → /me → logout |
| `login-sustained` | `scenarios/login-sustained.js` | Sustained login load (bcrypt + JWT isolation) |

## Profiles

| Profile | Pattern | Use case |
|---|---|---|
| `smoke` | 5 VUs, 50s | Sanity check — does it work? |
| `load` | 50 VUs, 2m | Normal production load |
| `stress` | Ramp to 300 VUs, 6m | Find the breaking point |
| `spike` | Jump to 500 VUs | Sudden traffic burst |

## Usage

Run a benchmark:
```powershell
.\benchmark\scripts\run.ps1 health smoke
.\benchmark\scripts\run.ps1 auth-flow load
```

Extract results to CSV:
```powershell
go run benchmark/cmd/benchtools/main.go extract benchmark/results/<file>.json --label "baseline"
```

Run k6 directly:
```bash
k6 run -e PROFILE=stress benchmark/scenarios/auth-flow.js
```

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
├── config.js                # Shared profiles, thresholds, base URL
├── helpers.js               # Shared utilities (cookies, unique emails)
├── dashboard.html           # Chart.js visualisation (open in browser)
├── results.csv              # Accumulated metrics (committed)
├── JOURNAL.md               # Optimisation notes and observations
├── cmd/benchtools/main.go   # Go tool to extract k6 results → CSV
├── scenarios/
│   ├── health.js            # Baseline throughput
│   ├── auth-flow.js         # Full auth lifecycle
│   └── login-sustained.js   # Sustained login hammering
├── scripts/
│   ├── run.ps1              # Run a single scenario (PowerShell)
│   ├── run-all.ps1          # Run all scenarios (PowerShell)
│   ├── run.sh               # Run a single scenario (bash)
│   └── run-all.sh           # Run all scenarios (bash)
└── results/                 # Raw k6 JSON output (gitignored)
```

# Benchmark Changelog

Log of changes made to the API and their impact on performance. Raw data lives in `results.csv`, visualisations in `dashboard.html`.

**Hardware:** AMD Ryzen 9800X3D (8c/16t, 5.2GHz boost), liquid cooled, Docker Desktop

---

### Baseline — 2026-03-10

**Label:** `baseline`

**Config:** bcrypt cost 12, 25 max DB conns, unbounded audit goroutines, no caching, 15s server timeouts

**Bottlenecks identified:**
1. Unbounded audit goroutines — every request spawns one, no worker pool
2. DB connection pool too small (25 max) — exhausted under load
3. No caching — JWT re-parsed and /me hits DB every time
4. Server timeouts at 15s — too tight for stress scenarios
5. Bcrypt is the CPU ceiling — ~160ms per hash, starves all other endpoints under load

---

### Round 1: Infrastructure optimisations — 2026-03-10

**Label:** `opt-v1`

**Changes:**
- Audit worker pool (10 workers, 1000 buffered queue) replacing unbounded goroutines
- Request body capture capped at 4KB, response body buffering removed
- DB pool: 25→50 max conns, 5→10 min conns, idle time 30min→5min
- Server read/write timeouts: 15s→30s
- In-memory TTL cache for JWT tokens and user lookups (5min TTL)
- Audit query pagination (default LIMIT 100)
- Audit user_id type: VARCHAR→INTEGER migration, direct JOIN

**Impact:** /me p95 at stress dropped 1,277ms → 356ms (-72%). Auth-flow p50 improved 694ms → 316ms (-54%). Login unchanged — bcrypt is CPU-bound.

---

### Round 2: Argon2id — 2026-03-10

**Label:** `argon2id`

**Changes:**
- Replaced bcrypt (cost 12, ~160ms/hash) with argon2id (m=47104, t=1, p=1, ~12ms/hash)
- OWASP recommended parameters, memory-hard security
- Clean swap — same function signatures, no migration needed

**Impact:** Auth-flow throughput doubled (55K→134K reqs, +142%). Login-sustained tripled (29K→81K, +179%). Auth-flow p50 at stress: 694ms → 95ms (-86%). The CPU ceiling from hashing is gone.

**Remaining bottleneck:** p95 tail (1-2s) at 300 VUs is now DB/connection contention. Next levers: connection pooling, read replicas, horizontal scaling.

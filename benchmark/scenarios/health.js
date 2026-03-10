// Baseline throughput test — no auth, no DB writes, minimal logic.
// Use this to measure raw framework overhead + audit middleware cost.
//
// Usage:
//   k6 run benchmark/scenarios/health.js
//   k6 run -e PROFILE=stress benchmark/scenarios/health.js

import http from "k6/http";
import { check, sleep } from "k6";
import { BASE_URL, profiles, thresholds } from "../config.js";

const profile = __ENV.PROFILE || "load";

export const options = {
  stages: profiles[profile],
  thresholds,
};

export default function () {
  const res = http.get(`${BASE_URL}/health`);

  check(res, {
    "status 200": (r) => r.status === 200,
    "has status field": (r) => JSON.parse(r.body).status !== undefined,
  });

  sleep(0.1);
}

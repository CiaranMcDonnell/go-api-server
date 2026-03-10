// Sustained login load — pre-registers users in setup, then hammers login only.
// Isolates bcrypt + JWT generation performance from registration overhead.
//
// Usage:
//   k6 run benchmark/scenarios/login-sustained.js
//   k6 run -e PROFILE=stress -e USERS=200 benchmark/scenarios/login-sustained.js

import http from "k6/http";
import { check, sleep } from "k6";
import { SharedArray } from "k6/data";
import { Trend } from "k6/metrics";
import { BASE_URL, profiles, thresholds } from "../config.js";
import { errors, jsonHeaders } from "../helpers.js";

const profile = __ENV.PROFILE || "load";
const userCount = parseInt(__ENV.USERS || "100");

const loginDuration = new Trend("login_duration");

export const options = {
  stages: profiles[profile],
  thresholds: {
    ...thresholds,
    login_duration: ["p(95)<1000", "p(99)<2000"],
  },
};

// Pre-generate credentials — shared across all VUs (read-only)
const users = new SharedArray("users", function () {
  const arr = [];
  for (let i = 0; i < userCount; i++) {
    arr.push({
      email: `sustained+${i}@test.com`,
      password: "SustainedTest1234!",
      name: "Sustained User",
    });
  }
  return arr;
});

// Setup: register all test users before the test begins.
// Idempotent — skips users that already exist from previous runs.
export function setup() {
  let created = 0;
  let existing = 0;

  for (const user of users) {
    const res = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify(user),
      { headers: jsonHeaders },
    );

    if (res.status === 201) {
      created++;
    } else {
      // User likely already exists — verify by logging in
      const loginRes = http.post(
        `${BASE_URL}/api/v1/auth/login`,
        JSON.stringify({ email: user.email, password: user.password }),
        { headers: jsonHeaders },
      );
      if (loginRes.status === 200) {
        existing++;
      } else {
        console.log(`Setup: failed to register or login ${user.email}: register=${res.status} login=${loginRes.status}`);
      }
    }
  }

  console.log(`Setup: ${created} created, ${existing} existing, ${users.length} total`);
  return { ready: created + existing };
}

export default function () {
  // Pick a random pre-registered user
  const user = users[Math.floor(Math.random() * users.length)];

  const res = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: user.email, password: user.password }),
    { headers: jsonHeaders },
  );

  loginDuration.add(res.timings.duration);

  if (!check(res, { "login 200": (r) => r.status === 200 })) {
    errors.add(1);
  }

  sleep(0.1);
}

// Full auth lifecycle: register → login → /me → logout.
// Tests bcrypt, JWT, DB writes, cookie handling, and audit logging together.
//
// Usage:
//   k6 run benchmark/scenarios/auth-flow.js
//   k6 run -e PROFILE=stress benchmark/scenarios/auth-flow.js

import http from "k6/http";
import { check, group, sleep } from "k6";
import { Trend } from "k6/metrics";
import { BASE_URL, profiles, thresholds } from "../config.js";
import { errors, jsonHeaders, extractCookie, uniqueEmail } from "../helpers.js";

const profile = __ENV.PROFILE || "load";

const registerDuration = new Trend("register_duration");
const loginDuration = new Trend("login_duration");
const meDuration = new Trend("me_duration");

export const options = {
  stages: profiles[profile],
  thresholds: {
    ...thresholds,
    register_duration: ["p(95)<2000"],
    login_duration: ["p(95)<1000"],
    me_duration: ["p(95)<200"],
  },
};

export default function () {
  const email = uniqueEmail("bench");
  const password = "BenchTest1234!";

  // --- Register ---
  group("register", () => {
    const res = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({ name: "Bench User", email, password }),
      { headers: jsonHeaders },
    );
    registerDuration.add(res.timings.duration);

    if (
      !check(res, { "register 201": (r) => r.status === 201 })
    ) {
      errors.add(1);
      return;
    }
  });

  sleep(0.3);

  // --- Login ---
  let cookie = null;

  group("login", () => {
    const res = http.post(
      `${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({ email, password }),
      { headers: jsonHeaders },
    );
    loginDuration.add(res.timings.duration);

    if (!check(res, { "login 200": (r) => r.status === 200 })) {
      errors.add(1);
      return;
    }

    cookie = extractCookie(res, "authToken");
    check(res, { "login has cookie": () => cookie !== null });
  });

  if (!cookie) return;
  sleep(0.2);

  // --- Get current user ---
  group("me", () => {
    const res = http.get(`${BASE_URL}/api/v1/auth/me`, {
      headers: { Cookie: `authToken=${cookie}` },
    });
    meDuration.add(res.timings.duration);

    if (!check(res, { "me 200": (r) => r.status === 200 })) {
      errors.add(1);
    }
  });

  sleep(0.2);

  // --- Logout ---
  group("logout", () => {
    const res = http.post(`${BASE_URL}/api/v1/auth/logout`, null, {
      headers: { Cookie: `authToken=${cookie}` },
    });
    check(res, { "logout 200": (r) => r.status === 200 });
  });

  sleep(0.3);
}

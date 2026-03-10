// Items CRUD + paginated listing benchmark.
// Registers a user, creates items, then hammers list/get/update/delete.
//
// Usage:
//   k6 run benchmark/scenarios/items-crud.js
//   k6 run -e PROFILE=stress benchmark/scenarios/items-crud.js

import http from "k6/http";
import { check, group, sleep } from "k6";
import { Trend } from "k6/metrics";
import { BASE_URL, profiles, thresholds } from "../config.js";
import { errors, jsonHeaders, extractCookie, uniqueEmail } from "../helpers.js";

const profile = __ENV.PROFILE || "load";

const createDuration = new Trend("create_item_duration");
const listDuration = new Trend("list_items_duration");
const getDuration = new Trend("get_item_duration");
const updateDuration = new Trend("update_item_duration");
const deleteDuration = new Trend("delete_item_duration");

export const options = {
  stages: profiles[profile],
  thresholds: {
    ...thresholds,
    create_item_duration: ["p(95)<300"],
    list_items_duration: ["p(95)<200"],
    get_item_duration: ["p(95)<150"],
    update_item_duration: ["p(95)<300"],
    delete_item_duration: ["p(95)<200"],
  },
};

function authHeaders(cookie) {
  return {
    "Content-Type": "application/json",
    Cookie: `authToken=${cookie}`,
  };
}

export function setup() {
  const email = `items-bench+${Date.now()}@test.com`;
  const password = "BenchTest1234!";

  http.post(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({ name: "Items Bench", email, password }),
    { headers: jsonHeaders },
  );

  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    { headers: jsonHeaders },
  );

  const cookie = extractCookie(loginRes, "authToken");
  if (!cookie) throw new Error("setup: login failed");

  // seed some items for pagination testing
  for (let i = 0; i < 25; i++) {
    http.post(
      `${BASE_URL}/api/v1/items`,
      JSON.stringify({ name: `Seed Item ${i}`, description: `Seeded for pagination test ${i}` }),
      { headers: authHeaders(cookie) },
    );
  }

  return { cookie };
}

export default function (data) {
  const headers = authHeaders(data.cookie);

  // --- Create ---
  let itemId = null;
  group("create item", () => {
    const res = http.post(
      `${BASE_URL}/api/v1/items`,
      JSON.stringify({ name: `Bench ${Date.now()}`, description: "load test item" }),
      { headers },
    );
    createDuration.add(res.timings.duration);

    if (!check(res, { "create 201": (r) => r.status === 201 })) {
      errors.add(1);
      return;
    }
    itemId = res.json("item.id");
  });

  if (!itemId) return;
  sleep(0.1);

  // --- List page 1 ---
  group("list items page 1", () => {
    const res = http.get(`${BASE_URL}/api/v1/items?page=1&per_page=10`, { headers });
    listDuration.add(res.timings.duration);
    const body = res.json();

    check(res, {
      "list 200": (r) => r.status === 200,
      "has items": () => body.items && body.items.length > 0,
      "has pagination": () => body.pagination && body.pagination.page === 1,
      "has per_page": () => body.pagination && body.pagination.per_page === 10,
    }) || errors.add(1);
  });

  sleep(0.1);

  // --- List page 2 ---
  group("list items page 2", () => {
    const res = http.get(`${BASE_URL}/api/v1/items?page=2&per_page=10`, { headers });
    listDuration.add(res.timings.duration);
    const body = res.json();

    check(res, {
      "page2 200": (r) => r.status === 200,
      "page2 has_prev": () => body.pagination && body.pagination.has_prev_page === true,
    }) || errors.add(1);
  });

  sleep(0.1);

  // --- Get ---
  group("get item", () => {
    const res = http.get(`${BASE_URL}/api/v1/items/${itemId}`, {
      headers,
      tags: { name: "GET /api/v1/items/:id" },
    });
    getDuration.add(res.timings.duration);

    check(res, {
      "get 200": (r) => r.status === 200,
      "correct id": (r) => r.json("item.id") === itemId,
    }) || errors.add(1);
  });

  sleep(0.1);

  // --- Update ---
  group("update item", () => {
    const res = http.put(
      `${BASE_URL}/api/v1/items/${itemId}`,
      JSON.stringify({ name: "Updated by k6" }),
      { headers, tags: { name: "PUT /api/v1/items/:id" } },
    );
    updateDuration.add(res.timings.duration);

    check(res, {
      "update 200": (r) => r.status === 200,
      "name updated": (r) => r.json("item.name") === "Updated by k6",
    }) || errors.add(1);
  });

  sleep(0.1);

  // --- Delete ---
  group("delete item", () => {
    const res = http.del(`${BASE_URL}/api/v1/items/${itemId}`, null, {
      headers,
      tags: { name: "DELETE /api/v1/items/:id" },
    });
    deleteDuration.add(res.timings.duration);

    check(res, { "delete 200": (r) => r.status === 200 }) || errors.add(1);
  });

  sleep(0.3);
}

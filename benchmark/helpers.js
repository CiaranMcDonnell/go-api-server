import { Counter } from "k6/metrics";

export const errors = new Counter("errors");

export const jsonHeaders = { "Content-Type": "application/json" };

export function extractCookie(res, name) {
  const header = res.headers["Set-Cookie"];
  if (!header) return null;
  const match = header.match(new RegExp(`${name}=([^;]+)`));
  return match ? match[1] : null;
}

export function uniqueEmail(prefix) {
  return `${prefix}+${__VU}-${__ITER}-${Date.now()}@test.com`;
}

export const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export const profiles = {
  smoke: [
    { duration: "10s", target: 5 },
    { duration: "30s", target: 5 },
    { duration: "10s", target: 0 },
  ],
  load: [
    { duration: "30s", target: 50 },
    { duration: "1m", target: 50 },
    { duration: "30s", target: 0 },
  ],
  stress: [
    { duration: "30s", target: 50 },
    { duration: "1m", target: 100 },
    { duration: "1m", target: 200 },
    { duration: "1m", target: 300 },
    { duration: "2m", target: 300 },
    { duration: "30s", target: 0 },
  ],
  spike: [
    { duration: "10s", target: 10 },
    { duration: "10s", target: 500 },
    { duration: "30s", target: 500 },
    { duration: "10s", target: 10 },
    { duration: "30s", target: 0 },
  ],
  breakpoint: [
    { duration: "30s", target: 100 },
    { duration: "1m", target: 300 },
    { duration: "1m", target: 500 },
    { duration: "1m", target: 750 },
    { duration: "1m", target: 1000 },
    { duration: "2m", target: 1000 },
    { duration: "30s", target: 0 },
  ],
};

export const thresholds = {
  http_req_duration: ["p(95)<500", "p(99)<1000"],
  http_req_failed: ["rate<0.01"],
};

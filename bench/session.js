import http from "k6/http";
import { check } from "k6";
import exec from "k6/execution";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const TOKEN_COUNT = Number(__ENV.TOKEN_COUNT || 1000);

if (!Number.isInteger(TOKEN_COUNT) || TOKEN_COUNT < 1) {
  throw new Error("TOKEN_COUNT must be a positive integer");
}

export const options = {
  discardResponseBodies: true,
  summaryTrendStats: ["avg", "med", "p(90)", "p(95)", "p(99)", "max"],
  thresholds: {
    checks: ["rate==1"],
    http_req_failed: ["rate==0"],
  },
};

export default function () {
  const tokenNumber = (exec.scenario.iterationInTest % TOKEN_COUNT) + 1;
  const token = `benchmark-token-${tokenNumber}`;

  const res = http.get(`${BASE_URL}/me`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    tags: {
      name: "GET /me",
    },
  });
  check(res, { "status is 200": (response) => response.status === 200 });
}

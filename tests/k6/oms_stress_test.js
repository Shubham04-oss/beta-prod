import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

export function handleSummary(data) {
  return {
    "summary.html": htmlReport(data),
  };
}

export const options = {
  scenarios: {
    // 1. Extreme Spike on single variant to test DB locks
    flash_sale: {
      executor: 'ramping-arrival-rate',
      startRate: 50,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 500,
      stages: [
        { target: 200, duration: '2s' }, // Ramp up to 200 requests/sec
        { target: 500, duration: '5s' }, // Spike to 500 requests/sec
        { target: 0, duration: '2s' },   // Cool down
      ],
      env: { SCENARIO: 'flash_sale' },
    },
    // 2. Edge Cases (Invalid variants, negative quantities, massive numbers)
    edge_cases: {
      executor: 'constant-vus',
      vus: 20,
      duration: '5s',
      env: { SCENARIO: 'edge_cases' },
    }
  },
  thresholds: {
    http_req_failed: ['rate<0.1'], // Some fail intentionally due to edge cases/OOS, but API shouldn't crash (500)
  },
};

const API_URL = __ENV.API_URL || 'http://localhost:8080/v1/oms/orders';
const VALID_VARIANT_ID = __ENV.VARIANT_ID || '00000000-0000-0000-0000-000000000000';

export default function () {
  const scenario = __ENV.SCENARIO;
  let payload = {};

  if (scenario === 'flash_sale') {
    payload = {
      customer_id: uuidv4(),
      idempotency_key: uuidv4(),
      items: [
        {
          variant_id: VALID_VARIANT_ID,
          quantity: 1, // Standard flash sale buy
          unit_price: 19.99,
        },
      ],
    };
  } else {
    // Edge Cases: Pick randomly
    const rand = Math.random();
    if (rand < 0.33) {
      // Negative quantity
      payload = {
        customer_id: uuidv4(),
        idempotency_key: uuidv4(),
        items: [{ variant_id: VALID_VARIANT_ID, quantity: -5, unit_price: 10.0 }],
      };
    } else if (rand < 0.66) {
      // Invalid variant
      payload = {
        customer_id: uuidv4(),
        idempotency_key: uuidv4(),
        items: [{ variant_id: uuidv4(), quantity: 1, unit_price: 10.0 }],
      };
    } else {
      // Massive integer overflow attempt
      payload = {
        customer_id: uuidv4(),
        idempotency_key: uuidv4(),
        items: [{ variant_id: VALID_VARIANT_ID, quantity: 2147483647, unit_price: 1.0 }],
      };
    }
  }

  const params = {
    headers: {
      'Content-Type': 'application/json',
      // Using a test Tenant ID if required by middleware. Wait, is our API authenticated?
      // For now we assume no auth or we just pass a fake header if needed.
    },
  };

  const res = http.post(API_URL, JSON.stringify(payload), params);

  // We only check that the API didn't completely die (i.e. status 500 usually means crash)
  // Status 200 is success, Status 400 is caught edge case. Both are "healthy" API behavior.
  check(res, {
    'API is alive (not 500)': (r) => r.status !== 500 && r.status !== 502 && r.status !== 503,
  });
}

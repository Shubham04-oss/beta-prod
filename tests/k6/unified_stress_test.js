import http from 'k6/http';
import { check, sleep } from 'k6';
import crypto from 'k6/crypto';

// Configuration
const API_URL = __ENV.API_URL || 'http://localhost:8080';
const WEBHOOK_SECRET = __ENV.UNIFIED_WEBHOOK_SECRET || 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ3b3Jrc3BhY2VfaWQiOiI2YTMwNDMyZDI1MDc0YmExZmE5NDBjYWUifQ.Ao3QSsnk_oMrWBi7nZc5Ukpp_7xzRbWBVpUJgg2yCgs';

// Test options
export const options = {
    scenarios: {
        webhook_spike: {
            executor: 'ramping-arrival-rate',
            startRate: 10,
            timeUnit: '1s',
            preAllocatedVUs: 100,
            maxVUs: 1000,
            stages: [
                { duration: '30s', target: 200 }, // Ramp up to 200 webhooks per second
                { duration: '1m', target: 200 },  // Sustain 200 webhooks per second
                { duration: '30s', target: 0 },   // Ramp down
            ],
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.01'], // less than 1% errors allowed
        http_req_duration: ['p(95)<500'], // 95% of webhook ingestions should be <500ms
    },
};

export default function () {
    const orderId = `synth-order-${__VU}-${__ITER}-${Date.now()}`;
    
    // Use valid UUIDs so `uuid.MustParse` in the Go backend does not panic
    const isFlashSale = Math.random() < 0.5; // 50% chance to target the exact same flash-sale SKU
    const flashSaleUUID = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'; // A deterministic valid UUID for the flash sale
    // Simple random UUID generator for k6 without external libs
    function uuidv4() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }
    const variantId = isFlashSale ? flashSaleUUID : uuidv4();

    const payload = JSON.stringify({
        event: "accounting.order.created",
        connection_id: "6a30e21564edf6eb93793d60", // Real Shopify Sandbox Connection ID
        workspace_id: "6a30432d25074ba1fa940cae",
        data: {
            id: orderId,
            total_amount: Math.floor(Math.random() * 1000) + 10,
            currency: "USD",
            line_items: [
                {
                    item_id: variantId,
                    quantity: 1,
                    unit_amount: 10.00
                }
            ],
            customer: {
                id: `cust-${__VU}`,
                email: `loadtester${__VU}@synq.example.com`
            }
        }
    });

    // Create HMAC SHA256 Signature to authenticate the webhook just like Unified.to does
    const signature = crypto.hmac('sha256', WEBHOOK_SECRET, payload, 'hex');

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Unified-Signature': signature,
        },
    };

    const res = http.post(`${API_URL}/v1/unified/webhooks`, payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200 || r.status === 202,
    });
}

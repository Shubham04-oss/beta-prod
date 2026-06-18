import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

// A simple fake string generator instead of pulling in faker.js to keep dependencies light
function randomString(length) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

export function handleSummary(data) {
    return {
        "summary_e2e.html": htmlReport(data),
    };
}

export const options = {
    scenarios: {
        e2e_journey: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '10s', target: 20 }, // Ramp up to 20 VUs
                { duration: '30s', target: 20 }, // Hold
                { duration: '10s', target: 0 },  // Ramp down
            ],
        },
        flash_sale_hot_key: {
            executor: 'shared-iterations',
            vus: 50,
            iterations: 200,
            startTime: '20s', // Starts middle of the e2e journey
        }
    },
    thresholds: {
        'http_req_failed': ['rate<0.05'], // Max 5% failure (we expect some 409s during flash sale)
        'http_req_duration': ['p(95)<500'], // 95% of requests must complete below 500ms
    },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

// A shared global variant ID for the flash sale scenario
const HOT_KEY_VARIANT_ID = "00000000-0000-0000-0000-000000000000"; 
const HOT_KEY_TENANT_ID = "698ec5cf-5b2e-4e95-a0d0-611d0a1ff30d"; // Dummy tenant used in main.go
const HOT_KEY_LOCATION_ID = "11111111-1111-1111-1111-111111111111";

export function setup() {
    // Setup the Hot Key Variant for the flash sale
    // We assume the dummy tenant is injected by the middleware for all requests in our local dev setup
    const productPayload = JSON.stringify({
        title: "Flash Sale TV",
        description: "Massive discount",
        category: "electronics"
    });

    const headers = { 'Content-Type': 'application/json' };
    const pRes = http.post(`${BASE_URL}/api/v1/pim/products`, productPayload, { headers });
    
    let productId = uuidv4();
    if (pRes.status === 200 || pRes.status === 201) {
        productId = pRes.json().id;
    }

    const vRes = http.post(`${BASE_URL}/api/v1/pim/products/${productId}/variants`, JSON.stringify({
        sku: "HOT-KEY-TV",
        barcode: "123456789",
        price: 99.99,
        currency: "USD"
    }), { headers });

    let variantId = HOT_KEY_VARIANT_ID;
    if (vRes.status === 200 || vRes.status === 201) {
        variantId = vRes.json().id;
    }

    // Inject limited inventory for the hot key
    http.post(`${BASE_URL}/api/v1/inventory/adjust`, JSON.stringify({
        variant_id: variantId,
        location_id: HOT_KEY_LOCATION_ID,
        quantity_delta: 50 // Only 50 items available for the 200 iterations flash sale!
    }), { headers });

    return { hotKeyVariantId: variantId };
}

export default function (data) {
    const headers = { 'Content-Type': 'application/json' };
    
    // Scenario 2: Flash Sale Hot Key Contention
    if (__VU >= 20) { // Using higher VU IDs for the flash sale scenario
        const orderPayload = JSON.stringify({
            idempotency_key: uuidv4(),
            items: [{
                variant_id: data.hotKeyVariantId,
                location_id: HOT_KEY_LOCATION_ID,
                quantity: 1,
                unit_price: 99.99
            }]
        });

        let res = http.post(`${BASE_URL}/api/v1/oms/orders`, orderPayload, { headers });
        // We EXPECT some of these to fail with 409 Conflict once the 50 items are sold out
        check(res, {
            'order succeeded or gracefully denied (no 500s)': (r) => r.status === 200 || r.status === 409 || r.status === 400,
        });
        return;
    }

    // Scenario 1: E2E Full Lifecycle
    group('1. Product Creation', function () {
        const payload = JSON.stringify({
            title: `Product ${randomString(10)}`, // Edge case: Random weird string
            description: `Desc ${randomString(50)}`,
            category: "test"
        });
        const pRes = http.post(`${BASE_URL}/api/v1/pim/products`, payload, { headers });
        if (pRes.status !== 200 && pRes.status !== 201) {
            console.log(`Product creation failed: ${pRes.status} ${pRes.body}`);
        }
        check(pRes, { 'product created': (r) => r.status === 200 || r.status === 201 });
        
        let productId = uuidv4();
        if (pRes.status === 200 || pRes.status === 201) productId = pRes.json().id;

        // Create Variant
        const vPayload = JSON.stringify({
            sku: `SKU-${randomString(8)}`,
            barcode: randomString(12),
            price: Math.random() * 100,
            currency: "USD"
        });
        const vRes = http.post(`${BASE_URL}/api/v1/pim/products/${productId}/variants`, vPayload, { headers });
        check(vRes, { 'variant created': (r) => r.status === 200 || r.status === 201 });

        let variantId = uuidv4();
        if (vRes.status === 200 || vRes.status === 201) variantId = vRes.json().id;

        // Adjust Inventory
        const locId = uuidv4(); // Random location
        const iPayload = JSON.stringify({
            variant_id: variantId,
            location_id: locId,
            quantity_delta: 1000 // Plenty of stock
        });
        const iRes = http.post(`${BASE_URL}/api/v1/inventory/adjust`, iPayload, { headers });
        check(iRes, { 'inventory adjusted': (r) => r.status === 200 || r.status === 201 });

        // Create Orders
        group('2. Order Checkouts', function() {
            // Place 5 orders for this product
            for (let i = 0; i < 5; i++) {
                const oPayload = JSON.stringify({
                    idempotency_key: uuidv4(),
                    items: [{
                        variant_id: variantId,
                        location_id: locId,
                        quantity: 2,
                        unit_price: 10.00
                    }]
                });
                const oRes = http.post(`${BASE_URL}/api/v1/oms/orders`, oPayload, { headers });
                check(oRes, { 'order successful': (r) => r.status === 200 || r.status === 201 });
                sleep(0.5); // Think time between orders
            }
        });
    });
}

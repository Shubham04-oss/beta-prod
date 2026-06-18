import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';

export let options = {
    vus: 100,
    duration: '60s',
};

// Only run the setup logic once to determine hostnames if needed
export function setup() {
    return {
        base_url: 'http://localhost:8080',
        firebase_url: 'http://shubhams-mac-mini.local:9099/identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=fake-api-key'
    };
}

export default function (data) {
    const uniqueId = randomString(8);
    const email = `stress_user_${uniqueId}@test.com`;
    const password = "password123";

    // 1. Onboard (Creates Tenant, User, Org, and registers in Firebase Emulator)
    let onboardRes = http.post(`${data.base_url}/api/v1/onboard`, JSON.stringify({
        org_name: `Stress Org ${uniqueId}`,
        tenant_name: `Stress Tenant ${uniqueId}`,
        admin_email: email,
        admin_password: password
    }), { headers: { 'Content-Type': 'application/json' } });

    check(onboardRes, {
        'onboard status is 201': (r) => r.status === 201,
    });

    if (onboardRes.status !== 201) {
        return; // Skip rest if onboard failed
    }

    // 2. Authenticate against Firebase Emulator to get JWT
    let authRes = http.post(data.firebase_url, JSON.stringify({
        email: email,
        password: password,
        returnSecureToken: true
    }), { headers: { 'Content-Type': 'application/json' } });

    check(authRes, {
        'auth status is 200': (r) => r.status === 200,
    });

    let idToken = "";
    if (authRes.status === 200) {
        let authData = authRes.json();
        idToken = authData.idToken;
    } else {
        return;
    }

    const authHeaders = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${idToken}`
    };

    // 3. Hammer PIM Endpoints
    // Create Product
    for (let i = 0; i < 5; i++) {
        let prodRes = http.post(`${data.base_url}/api/v1/pim/products`, JSON.stringify({
            title: `Stress Product ${uniqueId}-${i}`,
            description: "Heavy load testing product",
            category: "LoadTest"
        }), { headers: authHeaders });

        check(prodRes, {
            'create product 200': (r) => r.status === 200,
        });

        // Create Brand
        let brandRes = http.post(`${data.base_url}/api/v1/pim/brands`, JSON.stringify({
            name: `Stress Brand ${uniqueId}-${i}`,
            description: "Brand created during stress test"
        }), { headers: authHeaders });
        
        check(brandRes, {
            'create brand 200': (r) => r.status === 200,
        });
    }

    // Optional short sleep to simulate user think time and prevent total socket exhaustion
    sleep(1);
}

import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

export function handleSummary(data) {
  return {
    "scratch/summary.html": htmlReport(data),
    stdout: textSummary(data, { indent: " ", enableColors: true }),
  };
}

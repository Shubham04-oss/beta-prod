import http from 'k6/http';
import { check } from 'k6';
import crypto from 'k6/crypto';

export let options = {
    vus: 50,
    duration: '30s',
};

export default function () {
    const url = 'http://localhost:8080/unified/webhook';
    const payload = JSON.stringify({
        connection_id: "fake_shopify_123", 
        event: "item.created",
        data: {
            id: `test_item_${__ITER}`
        }
    });

    const secret = 'dummy_secret';
    const signature = crypto.hmac('sha256', secret, payload, 'hex');

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Unified-Signature': signature,
        },
    };

    const res = http.post(url, payload, params);

    check(res, {
        'is status 200': (r) => r.status === 200,
    });
}

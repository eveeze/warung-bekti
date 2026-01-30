// K6 Quick Stress Test - Runs in ~3 minutes
// Use this for quick validation before running full load test
// Run: k6 run tests/load/quick_test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1`;

export const options = {
    scenarios: {
        quick_ramp: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 50 },   // Ramp to 50
                { duration: '1m', target: 100 },   // Ramp to 100
                { duration: '1m', target: 100 },   // Hold
                { duration: '30s', target: 0 },    // Ramp down
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<2000'],
        http_req_failed: ['rate<0.1'],
    },
};

export function setup() {
    // Login to get token
    const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
        email: 'admin@warung.com',
        password: 'password',
    }), { headers: { 'Content-Type': 'application/json' } });
    
    let token = '';
    if (loginRes.status === 200) {
        token = JSON.parse(loginRes.body).data.access_token;
    }
    
    return { token };
}

export default function(data) {
    const { token } = data;
    const headers = { 'Authorization': `Bearer ${token}` };
    
    // Test 1: Product list (most common operation)
    const listRes = http.get(`${API_URL}/products?page=1&per_page=20`, { headers });
    check(listRes, { 'product list OK': (r) => r.status === 200 });
    errorRate.add(listRes.status !== 200);
    
    sleep(0.5);
    
    // Test 2: Product search
    const searchRes = http.get(`${API_URL}/products?search=mie`, { headers });
    check(searchRes, { 'search OK': (r) => r.status === 200 });
    errorRate.add(searchRes.status !== 200);
    
    sleep(0.5);
    
    // Test 3: Dashboard
    const dashRes = http.get(`${API_URL}/reports/dashboard`, { headers });
    check(dashRes, { 'dashboard OK': (r) => r.status === 200 });
    errorRate.add(dashRes.status !== 200);
    
    sleep(1);
}

// K6 Database Stress Test
// Tests database under heavy concurrent load
// Run: k6 run tests/load/db_stress_test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const errorRate = new Rate('errors');
const queryDuration = new Trend('query_duration');
const writeSuccess = new Counter('write_success');
const readSuccess = new Counter('read_success');

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1`;

export const options = {
    scenarios: {
        // Heavy read workload (70% of users)
        read_heavy: {
            executor: 'constant-vus',
            vus: 70,
            duration: '5m',
            exec: 'readOperations',
        },
        // Write workload (30% of users)
        write_heavy: {
            executor: 'constant-vus',
            vus: 30,
            duration: '5m',
            exec: 'writeOperations',
        },
    },
    thresholds: {
        'query_duration': ['p(95)<1000', 'p(99)<3000'],
        'http_req_failed': ['rate<0.05'],
        'errors': ['rate<0.1'],
    },
};

let authToken = '';
let productIds = [];

export function setup() {
    // Login
    const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
        email: 'admin@warung.com',
        password: 'password',
    }), { headers: { 'Content-Type': 'application/json' } });
    
    if (loginRes.status === 200) {
        authToken = JSON.parse(loginRes.body).data.access_token;
    }
    
    // Get product IDs
    const productsRes = http.get(`${API_URL}/products?per_page=100`, {
        headers: { 'Authorization': `Bearer ${authToken}` },
    });
    
    if (productsRes.status === 200) {
        productIds = JSON.parse(productsRes.body).data.map(p => p.id);
    }
    
    return { authToken, productIds };
}

// Read-heavy operations
export function readOperations(data) {
    const headers = { 'Authorization': `Bearer ${data.authToken}` };
    
    // Random page
    const page = randomIntBetween(1, 50);
    
    // List products (most common)
    const start1 = Date.now();
    const listRes = http.get(`${API_URL}/products?page=${page}&per_page=20`, { headers });
    queryDuration.add(Date.now() - start1);
    
    if (check(listRes, { 'list ok': (r) => r.status === 200 })) {
        readSuccess.add(1);
    } else {
        errorRate.add(1);
    }
    
    sleep(0.2);
    
    // Search products
    const searchTerms = ['mie', 'sabun', 'rokok', 'gula', 'beras', 'aqua', 'coca'];
    const start2 = Date.now();
    const searchRes = http.get(`${API_URL}/products?search=${randomItem(searchTerms)}`, { headers });
    queryDuration.add(Date.now() - start2);
    
    if (check(searchRes, { 'search ok': (r) => r.status === 200 })) {
        readSuccess.add(1);
    } else {
        errorRate.add(1);
    }
    
    sleep(0.2);
    
    // Get single product (with pricing tiers)
    if (data.productIds.length > 0) {
        const productId = randomItem(data.productIds);
        const start3 = Date.now();
        const getRes = http.get(`${API_URL}/products/${productId}`, { headers });
        queryDuration.add(Date.now() - start3);
        
        if (check(getRes, { 'get ok': (r) => r.status === 200 || r.status === 404 })) {
            readSuccess.add(1);
        } else {
            errorRate.add(1);
        }
    }
    
    sleep(0.3);
    
    // Dashboard (aggregate query)
    const start4 = Date.now();
    const dashRes = http.get(`${API_URL}/reports/dashboard`, { headers });
    queryDuration.add(Date.now() - start4);
    
    if (check(dashRes, { 'dashboard ok': (r) => r.status === 200 })) {
        readSuccess.add(1);
    } else {
        errorRate.add(1);
    }
    
    sleep(0.5);
}

// Write operations
export function writeOperations(data) {
    const headers = { 
        'Authorization': `Bearer ${data.authToken}`,
        'Content-Type': 'application/json',
    };
    
    // Cart calculation (read + compute, common before checkout)
    if (data.productIds.length >= 3) {
        const items = [];
        for (let i = 0; i < 3; i++) {
            items.push({
                product_id: data.productIds[randomIntBetween(0, data.productIds.length - 1)],
                quantity: randomIntBetween(1, 5),
            });
        }
        
        const start1 = Date.now();
        const calcRes = http.post(`${API_URL}/transactions/calculate`, 
            JSON.stringify({ items }), { headers });
        queryDuration.add(Date.now() - start1);
        
        if (check(calcRes, { 'calc ok': (r) => r.status === 200 })) {
            writeSuccess.add(1);
        } else {
            errorRate.add(1);
        }
    }
    
    sleep(0.5);
    
    // Create customer (write)
    const customerData = {
        name: `Stress Test Customer ${Date.now()}`,
        phone: `08${randomIntBetween(1000000000, 9999999999)}`,
    };
    
    const start2 = Date.now();
    const custRes = http.post(`${API_URL}/customers`, JSON.stringify(customerData), { headers });
    queryDuration.add(Date.now() - start2);
    
    if (check(custRes, { 'customer create ok': (r) => r.status === 201 || r.status === 409 })) {
        writeSuccess.add(1);
    } else {
        errorRate.add(1);
    }
    
    sleep(1);
}

export function teardown(data) {
    console.log('Database stress test completed');
}

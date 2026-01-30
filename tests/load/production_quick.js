// K6 Quick Production Test - 5 minute validation
// Same patterns as production_test.js but shorter duration
// Run: k6 run tests/load/production_quick.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Metrics
const errorRate = new Rate('error_rate');
const successRate = new Rate('success_rate');
const checkoutDuration = new Trend('checkout_duration', true);
const productListDuration = new Trend('product_list_duration', true);
const transactionCount = new Counter('transactions_created');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1`;

// Quick test - 5 minutes total
export const options = {
    scenarios: {
        quick_production: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 50 },   // Warm up
                { duration: '1m', target: 150 },   // Ramp up
                { duration: '2m', target: 200 },   // Hold at peak
                { duration: '1m', target: 100 },   // Wind down
                { duration: '30s', target: 0 },    // Cool down
            ],
            gracefulRampDown: '10s',
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<2000'],
        http_req_failed: ['rate<0.05'],
        error_rate: ['rate<0.1'],
        'checkout_duration': ['p(95)<3000'],
    },
};

// Search terms
const SEARCH_TERMS = [
    'indomie', 'aqua', 'rokok', 'sabun', 'gula', 'beras', 
    'minyak', 'kopi', 'teh', 'susu', 'roti', 'snack',
];

const PAYMENT_WEIGHTS = { cash: 65, transfer: 15, qris: 15, kasbon: 5 };
const USER_WEIGHTS = { cashier: 60, browser: 25, manager: 10, inventory: 5 };

function weightedRandom(weights) {
    const entries = Object.entries(weights);
    const total = entries.reduce((sum, [_, w]) => sum + w, 0);
    let random = Math.random() * total;
    for (const [value, weight] of entries) {
        random -= weight;
        if (random <= 0) return value;
    }
    return entries[0][0];
}

export function setup() {
    console.log('🚀 Quick Production Test Setup...');
    
    const adminLogin = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
        email: 'admin@warung.com',
        password: 'password',
    }), { headers: { 'Content-Type': 'application/json' } });
    
    const token = adminLogin.status === 200 
        ? JSON.parse(adminLogin.body).data.access_token 
        : '';
    
    if (!token) {
        console.error('❌ Login failed!');
        return { token: '', products: [], customers: [] };
    }
    console.log('✅ Login successful');
    
    const productsRes = http.get(`${API_URL}/products?per_page=100`, {
        headers: { 'Authorization': `Bearer ${token}` },
    });
    const products = productsRes.status === 200 
        ? (JSON.parse(productsRes.body).data || []).map(p => ({ id: p.id, name: p.name }))
        : [];
    console.log(`📦 Loaded ${products.length} products`);
    
    const customersRes = http.get(`${API_URL}/customers?per_page=50`, {
        headers: { 'Authorization': `Bearer ${token}` },
    });
    const customers = customersRes.status === 200 
        ? (JSON.parse(customersRes.body).data || []).map(c => ({ id: c.id, name: c.name }))
        : [];
    console.log(`👥 Loaded ${customers.length} customers\n`);
    
    return { token, products, customers };
}

export default function(data) {
    const { token, products, customers } = data;
    if (!token) return;
    
    const headers = { 
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
    };
    
    const userType = weightedRandom(USER_WEIGHTS);
    
    switch (userType) {
        case 'cashier':
            runCashierFlow(headers, products, customers);
            break;
        case 'browser':
            runBrowserFlow(headers, products);
            break;
        case 'manager':
            runManagerFlow(headers);
            break;
        case 'inventory':
            runInventoryFlow(headers);
            break;
    }
    
    sleep(randomIntBetween(1, 2));
}

function runCashierFlow(headers, products, customers) {
    group('Cashier', () => {
        // List products
        const start1 = Date.now();
        const listRes = http.get(`${API_URL}/products?per_page=20`, { headers });
        productListDuration.add(Date.now() - start1);
        recordResult(listRes, 200, 'product_list');
        
        // Search
        if (Math.random() < 0.4) {
            const searchRes = http.get(
                `${API_URL}/products?search=${randomItem(SEARCH_TERMS)}`, 
                { headers }
            );
            recordResult(searchRes, 200, 'search');
        }
        
        sleep(0.3);
        
        // Checkout flow
        if (products.length >= 2) {
            const items = createCart(products, randomIntBetween(1, 4));
            
            // Calculate
            const calcRes = http.post(
                `${API_URL}/transactions/calculate`, 
                JSON.stringify({ items }), 
                { headers }
            );
            recordResult(calcRes, 200, 'calculate', [400]);
            
            // Complete transaction (70% of cashiers)
            if (calcRes.status === 200 && Math.random() < 0.7) {
                const calcData = JSON.parse(calcRes.body);
                const total = calcData.data?.total_amount || 50000;
                
                const paymentMethod = weightedRandom(PAYMENT_WEIGHTS);
                const txPayload = {
                    items,
                    payment_method: paymentMethod === 'kasbon' ? 'kasbon' : paymentMethod,
                    amount_paid: paymentMethod === 'kasbon' ? 0 : total + 5000,
                    cashier_name: 'Kasir Test',
                };
                
                if (paymentMethod === 'kasbon' && customers.length > 0) {
                    txPayload.customer_id = randomItem(customers).id;
                }
                
                const start = Date.now();
                const txRes = http.post(
                    `${API_URL}/transactions`, 
                    JSON.stringify(txPayload), 
                    { headers }
                );
                checkoutDuration.add(Date.now() - start);
                
                if (txRes.status === 201) {
                    transactionCount.add(1);
                    successRate.add(true);
                } else {
                    recordResult(txRes, 201, 'checkout', [400]);
                }
            }
        }
    });
}

function runBrowserFlow(headers, products) {
    group('Browser', () => {
        const listRes = http.get(`${API_URL}/products?per_page=20`, { headers });
        recordResult(listRes, 200, 'product_list');
        
        if (Math.random() < 0.6) {
            const searchRes = http.get(
                `${API_URL}/products?search=${randomItem(SEARCH_TERMS)}`, 
                { headers }
            );
            recordResult(searchRes, 200, 'search');
        }
        
        if (products.length > 0 && Math.random() < 0.3) {
            const product = randomItem(products);
            const detailRes = http.get(`${API_URL}/products/${product.id}`, { headers });
            recordResult(detailRes, 200, 'detail', [404]);
        }
    });
}

function runManagerFlow(headers) {
    group('Manager', () => {
        const dashRes = http.get(`${API_URL}/reports/dashboard`, { headers });
        recordResult(dashRes, 200, 'dashboard');
        
        if (Math.random() < 0.5) {
            const today = new Date().toISOString().split('T')[0];
            const dailyRes = http.get(`${API_URL}/reports/daily?date=${today}`, { headers });
            recordResult(dailyRes, 200, 'daily_report');
        }
        
        if (Math.random() < 0.3) {
            const kasbonRes = http.get(`${API_URL}/reports/kasbon`, { headers });
            recordResult(kasbonRes, 200, 'kasbon_report');
        }
    });
}

function runInventoryFlow(headers) {
    group('Inventory', () => {
        const lowStockRes = http.get(`${API_URL}/products/low-stock`, { headers });
        recordResult(lowStockRes, 200, 'low_stock');
        
        const listRes = http.get(`${API_URL}/products?per_page=50`, { headers });
        recordResult(listRes, 200, 'product_list');
    });
}

function createCart(products, numItems) {
    const items = [];
    const used = new Set();
    
    for (let i = 0; i < numItems && i < products.length; i++) {
        let product;
        let attempts = 0;
        do {
            product = randomItem(products);
            attempts++;
        } while (used.has(product.id) && attempts < 10);
        
        if (!used.has(product.id)) {
            used.add(product.id);
            items.push({ product_id: product.id, quantity: randomIntBetween(1, 3) });
        }
    }
    return items;
}

function recordResult(res, expected, name, acceptable = []) {
    const ok = res.status === expected || acceptable.includes(res.status);
    check(res, { [`${name} OK`]: () => ok });
    
    if (ok) {
        successRate.add(true);
    } else {
        errorRate.add(true);
    }
}

export function teardown() {
    console.log('\n✅ Quick Production Test Complete!');
}

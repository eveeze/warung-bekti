// K6 Load Testing Script for Warung Backend
// Simulates 1000 concurrent users with realistic warung operations
// Run: k6 run tests/load/load_test.js

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics
const errorRate = new Rate('errors');
const productListDuration = new Trend('product_list_duration');
const checkoutDuration = new Trend('checkout_duration');
const loginDuration = new Trend('login_duration');
const searchDuration = new Trend('search_duration');
const transactionCount = new Counter('transactions_created');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1`;

// Test options - simulating 1000 concurrent users
export const options = {
    scenarios: {
        // Ramp up to 1000 users over 5 minutes, hold for 10 minutes
        load_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '2m', target: 100 },   // Warm up
                { duration: '3m', target: 500 },   // Ramp up to 500
                { duration: '5m', target: 1000 },  // Ramp up to 1000
                { duration: '10m', target: 1000 }, // Hold at 1000
                { duration: '2m', target: 0 },     // Ramp down
            ],
            gracefulRampDown: '30s',
        },
        // Spike test - sudden burst of users
        spike_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '10s', target: 500 },  // Sudden spike
                { duration: '1m', target: 500 },   // Hold spike
                { duration: '10s', target: 0 },    // Drop
            ],
            startTime: '22m', // Start after main load test
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<2000', 'p(99)<5000'], // 95% under 2s, 99% under 5s
        http_req_failed: ['rate<0.05'],                   // Less than 5% errors
        errors: ['rate<0.1'],                             // Less than 10% business errors
        'product_list_duration': ['p(95)<1000'],          // Product list under 1s
        'checkout_duration': ['p(95)<3000'],              // Checkout under 3s
    },
};

// Shared test data
let authTokens = {};
let productIds = [];
let customerIds = [];

// Setup function - runs once before load test
export function setup() {
    console.log('Setting up load test...');
    
    // Login as admin to get token
    const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
        email: 'admin@warung.com',
        password: 'password',
    }), {
        headers: { 'Content-Type': 'application/json' },
    });
    
    let adminToken = '';
    if (loginRes.status === 200) {
        const body = JSON.parse(loginRes.body);
        adminToken = body.data.access_token;
        console.log('Admin login successful');
    } else {
        console.log('Admin login failed, proceeding without auth');
    }
    
    // Fetch some products for testing
    const productsRes = http.get(`${API_URL}/products?per_page=100`, {
        headers: { 'Authorization': `Bearer ${adminToken}` },
    });
    
    let products = [];
    if (productsRes.status === 200) {
        const body = JSON.parse(productsRes.body);
        products = body.data || [];
        console.log(`Loaded ${products.length} products for testing`);
    }
    
    // Fetch some customers for testing
    const customersRes = http.get(`${API_URL}/customers?per_page=50`, {
        headers: { 'Authorization': `Bearer ${adminToken}` },
    });
    
    let customers = [];
    if (customersRes.status === 200) {
        const body = JSON.parse(customersRes.body);
        customers = body.data || [];
        console.log(`Loaded ${customers.length} customers for testing`);
    }
    
    return {
        adminToken,
        products: products.map(p => ({ id: p.id, name: p.name, barcode: p.barcode })),
        customers: customers.map(c => ({ id: c.id, name: c.name })),
    };
}

// Main test function - runs for each virtual user
export default function(data) {
    const { adminToken, products, customers } = data;
    
    // Simulate different user behaviors
    const userType = randomItem(['browser', 'buyer', 'cashier', 'manager']);
    
    group('User Authentication', () => {
        if (Math.random() < 0.3) { // 30% of users try to login
            testLogin();
        }
    });
    
    group('Product Browsing', () => {
        testProductList(adminToken);
        testProductSearch(adminToken, products);
        
        if (products.length > 0) {
            testProductDetail(adminToken, products);
        }
    });
    
    if (userType === 'buyer' || userType === 'cashier') {
        group('Cart and Checkout', () => {
            if (products.length > 0) {
                testCartCalculation(adminToken, products);
                
                // Only 10% actually complete checkout to avoid too many transactions
                if (Math.random() < 0.1) {
                    testCheckout(adminToken, products, customers);
                }
            }
        });
    }
    
    if (userType === 'manager') {
        group('Reports and Dashboard', () => {
            testDashboard(adminToken);
            testLowStock(adminToken);
        });
    }
    
    group('Customer Operations', () => {
        if (Math.random() < 0.2) { // 20% query customers
            testCustomerList(adminToken);
            
            if (customers.length > 0 && Math.random() < 0.1) {
                testKasbonSummary(adminToken, customers);
            }
        }
    });
    
    // Think time between operations
    sleep(randomIntBetween(1, 3));
}

// Test functions

function testLogin() {
    const start = Date.now();
    const res = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
        email: 'kasir@warung.com',
        password: 'password123',
    }), {
        headers: { 'Content-Type': 'application/json' },
        tags: { name: 'Login' },
    });
    
    loginDuration.add(Date.now() - start);
    
    const success = check(res, {
        'login status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    });
    
    errorRate.add(!success);
}

function testProductList(token) {
    const start = Date.now();
    const page = randomIntBetween(1, 10);
    const perPage = randomItem([10, 20, 50]);
    
    const res = http.get(`${API_URL}/products?page=${page}&per_page=${perPage}`, {
        headers: { 'Authorization': `Bearer ${token}` },
        tags: { name: 'ProductList' },
    });
    
    productListDuration.add(Date.now() - start);
    
    const success = check(res, {
        'product list status is 200': (r) => r.status === 200,
        'product list has data': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.data !== undefined;
            } catch (e) {
                return false;
            }
        },
    });
    
    errorRate.add(!success);
}

function testProductSearch(token, products) {
    const start = Date.now();
    const searchTerms = ['indomie', 'mie', 'sabun', 'gula', 'beras', 'rokok', 'aqua', 'teh'];
    const search = randomItem(searchTerms);
    
    const res = http.get(`${API_URL}/products?search=${search}`, {
        headers: { 'Authorization': `Bearer ${token}` },
        tags: { name: 'ProductSearch' },
    });
    
    searchDuration.add(Date.now() - start);
    
    const success = check(res, {
        'search status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
}

function testProductDetail(token, products) {
    if (products.length === 0) return;
    
    const product = randomItem(products);
    const res = http.get(`${API_URL}/products/${product.id}`, {
        headers: { 'Authorization': `Bearer ${token}` },
        tags: { name: 'ProductDetail' },
    });
    
    const success = check(res, {
        'product detail status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    });
    
    errorRate.add(!success && res.status !== 404);
}

function testCartCalculation(token, products) {
    if (products.length < 2) return;
    
    // Create random cart with 1-5 items
    const numItems = randomIntBetween(1, 5);
    const items = [];
    const usedProducts = new Set();
    
    for (let i = 0; i < numItems && i < products.length; i++) {
        let product;
        do {
            product = randomItem(products);
        } while (usedProducts.has(product.id) && usedProducts.size < products.length);
        
        usedProducts.add(product.id);
        items.push({
            product_id: product.id,
            quantity: randomIntBetween(1, 10),
        });
    }
    
    const res = http.post(`${API_URL}/transactions/calculate`, JSON.stringify({ items }), {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
        tags: { name: 'CartCalculate' },
    });
    
    const success = check(res, {
        'cart calculation status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success && res.status !== 400);
}

function testCheckout(token, products, customers) {
    if (products.length < 1) return;
    
    const start = Date.now();
    
    // Create random cart
    const numItems = randomIntBetween(1, 3);
    const items = [];
    const usedProducts = new Set();
    
    for (let i = 0; i < numItems && i < products.length; i++) {
        let product;
        do {
            product = randomItem(products);
        } while (usedProducts.has(product.id) && usedProducts.size < products.length);
        
        usedProducts.add(product.id);
        items.push({
            product_id: product.id,
            quantity: randomIntBetween(1, 3),
        });
    }
    
    const paymentMethods = ['cash', 'transfer', 'qris'];
    const payload = {
        items,
        payment_method: randomItem(paymentMethods),
        amount_paid: 1000000, // Enough for most transactions
        cashier_name: 'Load Test User',
    };
    
    // Sometimes add customer for kasbon test
    if (customers.length > 0 && Math.random() < 0.2) {
        payload.customer_id = randomItem(customers).id;
    }
    
    const res = http.post(`${API_URL}/transactions`, JSON.stringify(payload), {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
        },
        tags: { name: 'Checkout' },
    });
    
    checkoutDuration.add(Date.now() - start);
    
    const success = check(res, {
        'checkout status is 201 or 400': (r) => r.status === 201 || r.status === 400,
    });
    
    if (res.status === 201) {
        transactionCount.add(1);
    }
    
    errorRate.add(!success && res.status !== 400);
}

function testDashboard(token) {
    const res = http.get(`${API_URL}/reports/dashboard`, {
        headers: { 'Authorization': `Bearer ${token}` },
        tags: { name: 'Dashboard' },
    });
    
    const success = check(res, {
        'dashboard status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
}

function testLowStock(token) {
    const res = http.get(`${API_URL}/products/low-stock`, {
        headers: { 'Authorization': `Bearer ${token}` },
        tags: { name: 'LowStock' },
    });
    
    const success = check(res, {
        'low stock status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
}

function testCustomerList(token) {
    const res = http.get(`${API_URL}/customers?per_page=20`, {
        headers: { 'Authorization': `Bearer ${token}` },
        tags: { name: 'CustomerList' },
    });
    
    const success = check(res, {
        'customer list status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
}

function testKasbonSummary(token, customers) {
    if (customers.length === 0) return;
    
    const customer = randomItem(customers);
    const res = http.get(`${API_URL}/kasbon/customers/${customer.id}/summary`, {
        headers: { 'Authorization': `Bearer ${token}` },
        tags: { name: 'KasbonSummary' },
    });
    
    const success = check(res, {
        'kasbon summary status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    });
    
    errorRate.add(!success && res.status !== 404);
}

// Teardown function
export function teardown(data) {
    console.log('Load test completed');
    console.log(`Total transactions created: ${transactionCount}`);
}

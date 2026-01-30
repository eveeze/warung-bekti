// K6 Production-Realistic Load Test for Warung Backend
// Simulates real warung traffic patterns with multiple user roles
// Run: k6 run tests/load/production_test.js
//
// Traffic Distribution (based on real warung patterns):
// - 60% Cashier operations (checkout, browse products)
// - 25% Customer browsing (view products, search)
// - 10% Manager operations (reports, dashboard)
// - 5% Inventory operations (stock check, updates)

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomIntBetween, randomItem } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics for detailed monitoring
const errorRate = new Rate('error_rate');
const successRate = new Rate('success_rate');

// Duration metrics per operation type
const loginDuration = new Trend('login_duration', true);
const productListDuration = new Trend('product_list_duration', true);
const productSearchDuration = new Trend('product_search_duration', true);
const productDetailDuration = new Trend('product_detail_duration', true);
const cartCalculateDuration = new Trend('cart_calculate_duration', true);
const checkoutDuration = new Trend('checkout_duration', true);
const dashboardDuration = new Trend('dashboard_duration', true);
const customerListDuration = new Trend('customer_list_duration', true);
const kasbonDuration = new Trend('kasbon_duration', true);
const categoryDuration = new Trend('category_duration', true);

// Counters
const transactionCount = new Counter('transactions_created');
const cashTransactions = new Counter('cash_transactions');
const transferTransactions = new Counter('transfer_transactions');
const qrisTransactions = new Counter('qris_transactions');
const kasbonTransactions = new Counter('kasbon_transactions');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1`;

// Test configuration - Production simulation
export const options = {
    scenarios: {
        // Phase 1: Morning rush (opening hours 06:00-09:00)
        morning_rush: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '1m', target: 50 },    // Warm up
                { duration: '2m', target: 200 },   // Morning customers coming in
                { duration: '3m', target: 300 },   // Peak morning
                { duration: '1m', target: 100 },   // Slow down
            ],
            gracefulRampDown: '30s',
        },
        // Phase 2: Steady daytime (09:00-17:00)
        steady_daytime: {
            executor: 'constant-vus',
            vus: 150,
            duration: '5m',
            startTime: '7m',
        },
        // Phase 3: Evening rush (17:00-21:00)
        evening_rush: {
            executor: 'ramping-vus',
            startVUs: 150,
            stages: [
                { duration: '2m', target: 400 },   // People coming home
                { duration: '3m', target: 500 },   // Peak evening
                { duration: '2m', target: 200 },   // Winding down
                { duration: '1m', target: 0 },     // Closing
            ],
            startTime: '12m',
            gracefulRampDown: '30s',
        },
        // Spike test - Flash sale / promo situation
        spike_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '10s', target: 300 },  // Sudden spike
                { duration: '30s', target: 500 },  // Hold spike
                { duration: '10s', target: 100 },  // Drop
            ],
            startTime: '20m',
        },
    },
    thresholds: {
        // Overall thresholds
        http_req_duration: ['p(95)<2000', 'p(99)<5000'],
        http_req_failed: ['rate<0.05'],
        error_rate: ['rate<0.1'],
        
        // Per-operation thresholds
        'product_list_duration': ['p(95)<1000'],
        'product_search_duration': ['p(95)<1500'],
        'cart_calculate_duration': ['p(95)<1000'],
        'checkout_duration': ['p(95)<3000'],
        'dashboard_duration': ['p(95)<2000'],
    },
};

// Realistic Indonesian product search terms
const SEARCH_TERMS = [
    'indomie', 'mie sedap', 'aqua', 'le minerale', 'teh botol', 'teh pucuk',
    'rokok', 'sampoerna', 'gudang garam', 'marlboro', 'surya',
    'sabun', 'rinso', 'daia', 'attack', 'molto', 'downy',
    'gula', 'beras', 'minyak', 'tepung', 'kecap', 'saos',
    'pulsa', 'token', 'gas', 'lpg', 'galon',
    'roti', 'biskuit', 'snack', 'chitato', 'oreo',
    'susu', 'dancow', 'indomilk', 'ultra', 'frisian',
    'kopi', 'kapal api', 'abc', 'good day', 'luwak',
];

// Payment method distribution (realistic for Indonesian warung)
const PAYMENT_WEIGHTS = {
    cash: 65,      // 65% cash
    transfer: 15,  // 15% bank transfer
    qris: 15,      // 15% QRIS
    kasbon: 5,     // 5% credit (kasbon)
};

// User type weights (realistic distribution)
const USER_WEIGHTS = {
    cashier: 60,    // Most operations are by cashier
    browser: 25,    // Customers just browsing
    manager: 10,    // Checking reports
    inventory: 5,   // Stock operations
};

// Helper function for weighted random selection
function weightedRandom(weights) {
    const entries = Object.entries(weights);
    const total = entries.reduce((sum, [_, weight]) => sum + weight, 0);
    let random = Math.random() * total;
    
    for (const [value, weight] of entries) {
        random -= weight;
        if (random <= 0) return value;
    }
    return entries[0][0];
}

// Setup function - runs once before load test
export function setup() {
    console.log('🚀 Setting up Production Load Test...');
    console.log(`   Target: ${BASE_URL}`);
    
    const tokens = {};
    const testData = {};
    
    // Login as Admin
    const adminLogin = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
        email: 'admin@warung.com',
        password: 'password',
    }), { headers: { 'Content-Type': 'application/json' } });
    
    if (adminLogin.status === 200) {
        tokens.admin = JSON.parse(adminLogin.body).data.access_token;
        console.log('✅ Admin login successful');
    } else {
        console.error('❌ Admin login failed');
    }
    
    // Login as Cashier
    const cashierLogin = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
        email: 'cashier@warung.com',
        password: 'password',
    }), { headers: { 'Content-Type': 'application/json' } });
    
    if (cashierLogin.status === 200) {
        tokens.cashier = JSON.parse(cashierLogin.body).data.access_token;
        console.log('✅ Cashier login successful');
    } else {
        // Fallback to admin token
        tokens.cashier = tokens.admin;
        console.log('⚠️ Cashier login failed, using admin token');
    }
    
    // Login as Inventory
    const inventoryLogin = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
        email: 'inventory@warung.com',
        password: 'password',
    }), { headers: { 'Content-Type': 'application/json' } });
    
    if (inventoryLogin.status === 200) {
        tokens.inventory = JSON.parse(inventoryLogin.body).data.access_token;
        console.log('✅ Inventory login successful');
    } else {
        tokens.inventory = tokens.admin;
        console.log('⚠️ Inventory login failed, using admin token');
    }
    
    // Fetch products
    const productsRes = http.get(`${API_URL}/products?per_page=200`, {
        headers: { 'Authorization': `Bearer ${tokens.admin}` },
    });
    
    if (productsRes.status === 200) {
        const body = JSON.parse(productsRes.body);
        testData.products = (body.data || []).map(p => ({
            id: p.id,
            name: p.name,
            barcode: p.barcode,
            base_price: p.base_price,
            current_stock: p.current_stock,
        }));
        console.log(`📦 Loaded ${testData.products.length} products`);
    } else {
        testData.products = [];
    }
    
    // Fetch customers
    const customersRes = http.get(`${API_URL}/customers?per_page=100`, {
        headers: { 'Authorization': `Bearer ${tokens.admin}` },
    });
    
    if (customersRes.status === 200) {
        const body = JSON.parse(customersRes.body);
        testData.customers = (body.data || []).map(c => ({
            id: c.id,
            name: c.name,
            credit_limit: c.credit_limit,
            current_debt: c.current_debt,
        }));
        console.log(`👥 Loaded ${testData.customers.length} customers`);
    } else {
        testData.customers = [];
    }
    
    // Fetch categories
    const categoriesRes = http.get(`${API_URL}/categories`, {
        headers: { 'Authorization': `Bearer ${tokens.admin}` },
    });
    
    if (categoriesRes.status === 200) {
        const body = JSON.parse(categoriesRes.body);
        testData.categories = (body.data || []).map(c => ({ id: c.id, name: c.name }));
        console.log(`📁 Loaded ${testData.categories.length} categories`);
    } else {
        testData.categories = [];
    }
    
    console.log('✅ Setup complete!\n');
    
    return { tokens, ...testData };
}

// Main test function
export default function(data) {
    const { tokens, products, customers, categories } = data;
    
    // Select user type based on weighted distribution
    const userType = weightedRandom(USER_WEIGHTS);
    
    // Get appropriate token for user type
    const token = userType === 'manager' ? tokens.admin : 
                  userType === 'inventory' ? tokens.inventory : 
                  tokens.cashier;
    
    const headers = { 
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
    };
    
    // Execute operations based on user type
    switch (userType) {
        case 'cashier':
            runCashierFlow(headers, products, customers);
            break;
        case 'browser':
            runBrowserFlow(headers, products, categories);
            break;
        case 'manager':
            runManagerFlow(headers, customers);
            break;
        case 'inventory':
            runInventoryFlow(headers, products);
            break;
    }
    
    // Think time - simulate human delay
    sleep(randomIntBetween(1, 3));
}

// === CASHIER FLOW (60% of traffic) ===
// Realistic cashier: scan products, add to cart, checkout
function runCashierFlow(headers, products, customers) {
    group('Cashier Operations', () => {
        // 1. Browse products (customer asking for something)
        const page = randomIntBetween(1, 5);
        const listRes = timedRequest(
            () => http.get(`${API_URL}/products?page=${page}&per_page=20`, { headers }),
            productListDuration
        );
        checkAndRecord(listRes, 200, 'product_list');
        
        sleep(0.3);
        
        // 2. Sometimes search for specific product (40% of time)
        if (Math.random() < 0.4) {
            const searchTerm = randomItem(SEARCH_TERMS);
            const searchRes = timedRequest(
                () => http.get(`${API_URL}/products?search=${encodeURIComponent(searchTerm)}`, { headers }),
                productSearchDuration
            );
            checkAndRecord(searchRes, 200, 'product_search');
            sleep(0.2);
        }
        
        // 3. Create cart and calculate
        if (products.length >= 2) {
            const numItems = randomIntBetween(1, 5);
            const cartItems = createRandomCart(products, numItems);
            
            const calcRes = timedRequest(
                () => http.post(`${API_URL}/transactions/calculate`, 
                    JSON.stringify({ items: cartItems }), 
                    { headers }
                ),
                cartCalculateDuration
            );
            checkAndRecord(calcRes, 200, 'cart_calculate', [400]);
            
            sleep(0.5);
            
            // 4. Complete checkout (80% of cashier flows end in checkout)
            if (Math.random() < 0.8 && calcRes.status === 200) {
                const calcData = JSON.parse(calcRes.body);
                const total = calcData.data?.total_amount || 50000;
                
                const paymentMethod = weightedRandom(PAYMENT_WEIGHTS);
                
                const txPayload = {
                    items: cartItems,
                    payment_method: paymentMethod === 'kasbon' ? 'kasbon' : paymentMethod,
                    cashier_name: 'Kasir ' + randomIntBetween(1, 3),
                };
                
                // Set payment details based on method
                if (paymentMethod === 'cash') {
                    txPayload.amount_paid = total + randomIntBetween(0, 10000);
                    cashTransactions.add(1);
                } else if (paymentMethod === 'transfer' || paymentMethod === 'qris') {
                    txPayload.amount_paid = total;
                    if (paymentMethod === 'transfer') transferTransactions.add(1);
                    else qrisTransactions.add(1);
                } else if (paymentMethod === 'kasbon' && customers.length > 0) {
                    // Select customer with credit available
                    const customer = randomItem(customers);
                    txPayload.customer_id = customer.id;
                    txPayload.amount_paid = 0;
                    kasbonTransactions.add(1);
                }
                
                const txRes = timedRequest(
                    () => http.post(`${API_URL}/transactions`, 
                        JSON.stringify(txPayload), 
                        { headers }
                    ),
                    checkoutDuration
                );
                
                if (txRes.status === 201) {
                    transactionCount.add(1);
                    successRate.add(true);
                } else {
                    // 400 is expected for validation errors
                    checkAndRecord(txRes, 201, 'checkout', [400]);
                }
            }
        }
    });
}

// === BROWSER FLOW (25% of traffic) ===
// Customer browsing products on mobile app or asking cashier
function runBrowserFlow(headers, products, categories) {
    group('Customer Browsing', () => {
        // 1. View product list
        const listRes = timedRequest(
            () => http.get(`${API_URL}/products?per_page=20`, { headers }),
            productListDuration
        );
        checkAndRecord(listRes, 200, 'product_list');
        
        sleep(0.5);
        
        // 2. Browse by category (if categories exist)
        if (categories.length > 0 && Math.random() < 0.3) {
            const category = randomItem(categories);
            const catRes = timedRequest(
                () => http.get(`${API_URL}/products?category_id=${category.id}`, { headers }),
                categoryDuration
            );
            checkAndRecord(catRes, 200, 'category_filter');
            sleep(0.3);
        }
        
        // 3. Search for products (60% of browsers search)
        if (Math.random() < 0.6) {
            const searchTerm = randomItem(SEARCH_TERMS);
            const searchRes = timedRequest(
                () => http.get(`${API_URL}/products?search=${encodeURIComponent(searchTerm)}`, { headers }),
                productSearchDuration
            );
            checkAndRecord(searchRes, 200, 'product_search');
            sleep(0.3);
        }
        
        // 4. View product detail (40% of browsers look at details)
        if (products.length > 0 && Math.random() < 0.4) {
            const product = randomItem(products);
            const detailRes = timedRequest(
                () => http.get(`${API_URL}/products/${product.id}`, { headers }),
                productDetailDuration
            );
            checkAndRecord(detailRes, 200, 'product_detail', [404]);
        }
        
        // 5. Pagination (20% browse more pages)
        if (Math.random() < 0.2) {
            const page = randomIntBetween(2, 10);
            const pageRes = http.get(`${API_URL}/products?page=${page}&per_page=20`, { headers });
            checkAndRecord(pageRes, 200, 'pagination');
        }
    });
}

// === MANAGER FLOW (10% of traffic) ===
// Owner/manager checking reports and dashboard
function runManagerFlow(headers, customers) {
    group('Manager Operations', () => {
        // 1. Dashboard (always)
        const dashRes = timedRequest(
            () => http.get(`${API_URL}/reports/dashboard`, { headers }),
            dashboardDuration
        );
        checkAndRecord(dashRes, 200, 'dashboard');
        
        sleep(0.5);
        
        // 2. Daily report (60% of managers)
        if (Math.random() < 0.6) {
            const today = new Date().toISOString().split('T')[0];
            const dailyRes = http.get(`${API_URL}/reports/daily?date=${today}`, { headers });
            checkAndRecord(dailyRes, 200, 'daily_report');
            sleep(0.3);
        }
        
        // 3. Kasbon report (40% of managers)
        if (Math.random() < 0.4) {
            const kasbonRes = timedRequest(
                () => http.get(`${API_URL}/reports/kasbon`, { headers }),
                kasbonDuration
            );
            checkAndRecord(kasbonRes, 200, 'kasbon_report');
            sleep(0.3);
        }
        
        // 4. Customer list with debt (30% of managers)
        if (Math.random() < 0.3) {
            const debtRes = timedRequest(
                () => http.get(`${API_URL}/customers/with-debt`, { headers }),
                customerListDuration
            );
            checkAndRecord(debtRes, 200, 'customers_with_debt');
        }
        
        // 5. View specific customer kasbon (if has customers)
        if (customers.length > 0 && Math.random() < 0.2) {
            const customer = randomItem(customers);
            const custKasbonRes = http.get(
                `${API_URL}/kasbon/customers/${customer.id}/summary`, 
                { headers }
            );
            checkAndRecord(custKasbonRes, 200, 'customer_kasbon', [404]);
        }
        
        // 6. Transaction list (50% of managers)
        if (Math.random() < 0.5) {
            const txListRes = http.get(`${API_URL}/transactions?per_page=20`, { headers });
            checkAndRecord(txListRes, 200, 'transaction_list');
        }
    });
}

// === INVENTORY FLOW (5% of traffic) ===
// Staff checking and updating stock
function runInventoryFlow(headers, products) {
    group('Inventory Operations', () => {
        // 1. Check low stock products (always)
        const lowStockRes = http.get(`${API_URL}/products/low-stock`, { headers });
        checkAndRecord(lowStockRes, 200, 'low_stock');
        
        sleep(0.3);
        
        // 2. Product list with stock info
        const listRes = timedRequest(
            () => http.get(`${API_URL}/products?per_page=50`, { headers }),
            productListDuration
        );
        checkAndRecord(listRes, 200, 'product_list');
        
        // 3. Check specific product details
        if (products.length > 0 && Math.random() < 0.5) {
            const product = randomItem(products);
            const detailRes = timedRequest(
                () => http.get(`${API_URL}/products/${product.id}`, { headers }),
                productDetailDuration
            );
            checkAndRecord(detailRes, 200, 'product_detail', [404]);
        }
        
        // 4. Search for specific product to restock
        if (Math.random() < 0.4) {
            const searchTerm = randomItem(['gas', 'galon', 'beras', 'gula', 'minyak']);
            const searchRes = http.get(
                `${API_URL}/products?search=${encodeURIComponent(searchTerm)}`, 
                { headers }
            );
            checkAndRecord(searchRes, 200, 'product_search');
        }
    });
}

// === HELPER FUNCTIONS ===

function timedRequest(requestFn, trendMetric) {
    const start = Date.now();
    const res = requestFn();
    trendMetric.add(Date.now() - start);
    return res;
}

function createRandomCart(products, numItems) {
    const items = [];
    const usedProducts = new Set();
    
    for (let i = 0; i < numItems && i < products.length; i++) {
        let product;
        let attempts = 0;
        
        do {
            product = randomItem(products);
            attempts++;
        } while (usedProducts.has(product.id) && attempts < 10);
        
        if (!usedProducts.has(product.id)) {
            usedProducts.add(product.id);
            items.push({
                product_id: product.id,
                quantity: randomIntBetween(1, 5),
            });
        }
    }
    
    return items;
}

function checkAndRecord(response, expectedStatus, operationName, acceptableErrors = []) {
    const isSuccess = response.status === expectedStatus || 
                      acceptableErrors.includes(response.status);
    
    check(response, {
        [`${operationName} status OK`]: (r) => isSuccess,
    });
    
    if (!isSuccess) {
        errorRate.add(true);
        console.log(`❌ ${operationName} failed: ${response.status} - ${response.body?.substring(0, 100)}`);
    } else {
        successRate.add(true);
    }
    
    return isSuccess;
}

// Teardown
export function teardown(data) {
    console.log('\n📊 Load Test Complete!');
    console.log('=' .repeat(50));
    console.log('Transaction Summary:');
}

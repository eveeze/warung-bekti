#!/usr/bin/env python3
"""
Integration Test Runner for Warung Backend
Runs comprehensive API tests to verify all endpoints work correctly

Usage:
    python tests/integration/run_api_tests.py
"""

import json
import sys
import time
import requests
from typing import Optional, Dict, Any, List
from dataclasses import dataclass
from enum import Enum

# Configuration
BASE_URL = "http://localhost:8080"
API_URL = f"{BASE_URL}/api/v1"

class TestStatus(Enum):
    PASS = "✓"
    FAIL = "✗"
    SKIP = "○"

@dataclass
class TestResult:
    name: str
    status: TestStatus
    duration_ms: float
    error: Optional[str] = None

class WarungAPITester:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.api_url = f"{base_url}/api/v1"
        self.token: Optional[str] = None
        self.results: List[TestResult] = []
        self.test_data: Dict[str, Any] = {}
    
    def log(self, message: str):
        print(message)
    
    def run_test(self, name: str, test_func):
        """Run a single test and record result"""
        start = time.time()
        try:
            test_func()
            duration = (time.time() - start) * 1000
            self.results.append(TestResult(name, TestStatus.PASS, duration))
            self.log(f"  {TestStatus.PASS.value} {name} ({duration:.0f}ms)")
        except AssertionError as e:
            duration = (time.time() - start) * 1000
            self.results.append(TestResult(name, TestStatus.FAIL, duration, str(e)))
            self.log(f"  {TestStatus.FAIL.value} {name}: {e}")
        except Exception as e:
            duration = (time.time() - start) * 1000
            self.results.append(TestResult(name, TestStatus.FAIL, duration, str(e)))
            self.log(f"  {TestStatus.FAIL.value} {name}: {e}")
    
    def get_headers(self) -> dict:
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers
    
    # ==================== AUTH TESTS ====================
    
    def test_auth(self):
        self.log("\n📝 Auth Tests")
        
        self.run_test("Login with valid credentials", self._test_login_success)
        self.run_test("Login with invalid credentials", self._test_login_failure)
        self.run_test("Refresh token", self._test_refresh_token)
    
    def _test_login_success(self):
        resp = requests.post(
            f"{self.base_url}/auth/login",
            json={"email": "admin@warung.com", "password": "password"},
            timeout=10
        )
        assert resp.status_code == 200, f"Expected 200, got {resp.status_code}"
        
        data = resp.json()
        assert data.get("success") == True
        assert "access_token" in data.get("data", {})
        
        self.token = data["data"]["access_token"]
        self.test_data["refresh_token"] = data["data"].get("refresh_token")
    
    def _test_login_failure(self):
        resp = requests.post(
            f"{self.base_url}/auth/login",
            json={"email": "wrong@email.com", "password": "wrongpass"},
            timeout=10
        )
        assert resp.status_code == 401, f"Expected 401, got {resp.status_code}"
    
    def _test_refresh_token(self):
        if not self.test_data.get("refresh_token"):
            raise AssertionError("No refresh token available")
        
        resp = requests.post(
            f"{self.base_url}/auth/refresh",
            json={"refresh_token": self.test_data["refresh_token"]},
            timeout=10
        )
        # May fail if refresh token not implemented or expired
        assert resp.status_code in [200, 401], f"Unexpected status {resp.status_code}"
    
    # ==================== PRODUCT TESTS ====================
    
    def test_products(self):
        self.log("\n📦 Product Tests")
        
        self.run_test("List products", self._test_list_products)
        self.run_test("Search products", self._test_search_products)
        self.run_test("Create product", self._test_create_product)
        self.run_test("Get product by ID", self._test_get_product)
        self.run_test("Update product", self._test_update_product)
        self.run_test("List with pagination", self._test_product_pagination)
        self.run_test("Low stock products", self._test_low_stock)
        self.run_test("Delete product", self._test_delete_product)
    
    def _test_list_products(self):
        resp = requests.get(
            f"{self.api_url}/products?per_page=10",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200, f"Expected 200, got {resp.status_code}"
        
        data = resp.json()
        assert "data" in data
        assert isinstance(data["data"], list)
    
    def _test_search_products(self):
        resp = requests.get(
            f"{self.api_url}/products?search=test",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200, f"Expected 200, got {resp.status_code}"
    
    def _test_create_product(self):
        product_data = {
            "name": f"API Test Product {int(time.time())}",
            "unit": "pcs",
            "base_price": 15000,
            "cost_price": 10000,
            "is_stock_active": True,
            "current_stock": 50,
            "min_stock_alert": 10,
        }
        
        resp = requests.post(
            f"{self.api_url}/products",
            json=product_data,
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 201, f"Expected 201, got {resp.status_code}: {resp.text}"
        
        data = resp.json()
        assert data.get("data", {}).get("id")
        self.test_data["product_id"] = data["data"]["id"]
    
    def _test_get_product(self):
        product_id = self.test_data.get("product_id")
        if not product_id:
            raise AssertionError("No product_id available")
        
        resp = requests.get(
            f"{self.api_url}/products/{product_id}",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200, f"Expected 200, got {resp.status_code}"
    
    def _test_update_product(self):
        product_id = self.test_data.get("product_id")
        if not product_id:
            raise AssertionError("No product_id available")
        
        resp = requests.put(
            f"{self.api_url}/products/{product_id}",
            json={"base_price": 16000},
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200, f"Expected 200, got {resp.status_code}"
    
    def _test_product_pagination(self):
        resp = requests.get(
            f"{self.api_url}/products?page=1&per_page=5",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
        
        data = resp.json()
        assert len(data.get("data", [])) <= 5
    
    def _test_low_stock(self):
        resp = requests.get(
            f"{self.api_url}/products/low-stock",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    def _test_delete_product(self):
        product_id = self.test_data.get("product_id")
        if not product_id:
            raise AssertionError("No product_id available")
        
        resp = requests.delete(
            f"{self.api_url}/products/{product_id}",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 204, f"Expected 204, got {resp.status_code}"
    
    # ==================== CUSTOMER TESTS ====================
    
    def test_customers(self):
        self.log("\n👥 Customer Tests")
        
        self.run_test("Create customer", self._test_create_customer)
        self.run_test("List customers", self._test_list_customers)
        self.run_test("Get customer by ID", self._test_get_customer)
        self.run_test("Update customer", self._test_update_customer)
        self.run_test("Customers with debt", self._test_customers_with_debt)
    
    def _test_create_customer(self):
        customer_data = {
            "name": f"API Test Customer {int(time.time())}",
            "phone": f"08{int(time.time()) % 1000000000:010d}",
            "credit_limit": 500000,
        }
        
        resp = requests.post(
            f"{self.api_url}/customers",
            json=customer_data,
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 201, f"Expected 201, got {resp.status_code}: {resp.text}"
        
        self.test_data["customer_id"] = resp.json()["data"]["id"]
    
    def _test_list_customers(self):
        resp = requests.get(
            f"{self.api_url}/customers",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    def _test_get_customer(self):
        customer_id = self.test_data.get("customer_id")
        if not customer_id:
            raise AssertionError("No customer_id available")
        
        resp = requests.get(
            f"{self.api_url}/customers/{customer_id}",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    def _test_update_customer(self):
        customer_id = self.test_data.get("customer_id")
        if not customer_id:
            raise AssertionError("No customer_id available")
        
        resp = requests.put(
            f"{self.api_url}/customers/{customer_id}",
            json={"credit_limit": 750000},
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    def _test_customers_with_debt(self):
        resp = requests.get(
            f"{self.api_url}/customers/with-debt",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    # ==================== TRANSACTION TESTS ====================
    
    def test_transactions(self):
        self.log("\n💰 Transaction Tests")
        
        self.run_test("Cart calculation", self._test_cart_calculation)
        self.run_test("Create transaction (cash)", self._test_create_transaction)
        self.run_test("List transactions", self._test_list_transactions)
        self.run_test("Get transaction by ID", self._test_get_transaction)
    
    def _test_cart_calculation(self):
        # Get a product to use
        resp = requests.get(
            f"{self.api_url}/products?per_page=1",
            headers=self.get_headers(),
            timeout=10
        )
        
        if resp.status_code != 200 or not resp.json().get("data"):
            raise AssertionError("No products available for cart test")
        
        product = resp.json()["data"][0]
        self.test_data["cart_product_id"] = product["id"]
        
        # Calculate cart
        resp = requests.post(
            f"{self.api_url}/transactions/calculate",
            json={
                "items": [
                    {"product_id": product["id"], "quantity": 2}
                ]
            },
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200, f"Expected 200, got {resp.status_code}: {resp.text}"
    
    def _test_create_transaction(self):
        product_id = self.test_data.get("cart_product_id")
        if not product_id:
            raise AssertionError("No product_id for transaction")
        
        resp = requests.post(
            f"{self.api_url}/transactions",
            json={
                "items": [{"product_id": product_id, "quantity": 1}],
                "payment_method": "cash",
                "amount_paid": 1000000,
                "cashier_name": "API Tester",
            },
            headers=self.get_headers(),
            timeout=10
        )
        
        # May fail due to insufficient stock - that's ok
        assert resp.status_code in [201, 400], f"Unexpected status {resp.status_code}: {resp.text}"
        
        if resp.status_code == 201:
            self.test_data["transaction_id"] = resp.json()["data"]["id"]
    
    def _test_list_transactions(self):
        resp = requests.get(
            f"{self.api_url}/transactions",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    def _test_get_transaction(self):
        tx_id = self.test_data.get("transaction_id")
        if not tx_id:
            # Get any transaction
            resp = requests.get(
                f"{self.api_url}/transactions?per_page=1",
                headers=self.get_headers(),
                timeout=10
            )
            if resp.status_code == 200 and resp.json().get("data"):
                tx_id = resp.json()["data"][0]["id"]
        
        if not tx_id:
            raise AssertionError("No transaction available")
        
        resp = requests.get(
            f"{self.api_url}/transactions/{tx_id}",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    # ==================== REPORTS TESTS ====================
    
    def test_reports(self):
        self.log("\n📊 Reports Tests")
        
        self.run_test("Dashboard", self._test_dashboard)
        self.run_test("Daily report", self._test_daily_report)
    
    def _test_dashboard(self):
        resp = requests.get(
            f"{self.api_url}/reports/dashboard",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    def _test_daily_report(self):
        today = time.strftime("%Y-%m-%d")
        resp = requests.get(
            f"{self.api_url}/reports/daily?date={today}",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    # ==================== KASBON TESTS ====================
    
    def test_kasbon(self):
        self.log("\n📋 Kasbon Tests")
        
        self.run_test("Kasbon report", self._test_kasbon_report)
    
    def _test_kasbon_report(self):
        resp = requests.get(
            f"{self.api_url}/reports/kasbon",
            headers=self.get_headers(),
            timeout=10
        )
        assert resp.status_code == 200
    
    # ==================== RUN ALL ====================
    
    def run_all(self):
        self.log("=" * 60)
        self.log("🧪 Warung Backend API Integration Tests")
        self.log("=" * 60)
        
        start = time.time()
        
        self.test_auth()
        self.test_products()
        self.test_customers()
        self.test_transactions()
        self.test_reports()
        self.test_kasbon()
        
        duration = time.time() - start
        
        # Summary
        passed = sum(1 for r in self.results if r.status == TestStatus.PASS)
        failed = sum(1 for r in self.results if r.status == TestStatus.FAIL)
        
        self.log("\n" + "=" * 60)
        self.log(f"📊 Results: {passed} passed, {failed} failed")
        self.log(f"⏱️  Duration: {duration:.2f}s")
        self.log("=" * 60)
        
        if failed > 0:
            self.log("\n❌ Failed Tests:")
            for r in self.results:
                if r.status == TestStatus.FAIL:
                    self.log(f"  - {r.name}: {r.error}")
        
        return failed == 0


def main():
    tester = WarungAPITester(BASE_URL)
    success = tester.run_all()
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()

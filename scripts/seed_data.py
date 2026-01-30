#!/usr/bin/env python3
"""
Seed Script for Warung Backend
Creates 2000 products with realistic Indonesian warung data
Also creates customers and categories for testing

Usage:
    python scripts/seed_data.py --products 2000 --customers 100
"""

import argparse
import json
import random
import string
import requests
from concurrent.futures import ThreadPoolExecutor, as_completed

# Configuration
BASE_URL = "http://localhost:8080"
API_URL = f"{BASE_URL}/api/v1"

# Realistic Indonesian product data
PRODUCT_TEMPLATES = [
    # Makanan Instan
    {"prefix": "Indomie Goreng", "category": "Makanan Instan", "base_price_range": (3500, 4500), "unit": "pcs"},
    {"prefix": "Indomie Kuah", "category": "Makanan Instan", "base_price_range": (3000, 4000), "unit": "pcs"},
    {"prefix": "Mie Sedaap Goreng", "category": "Makanan Instan", "base_price_range": (3000, 4000), "unit": "pcs"},
    {"prefix": "Mie Sedaap Kuah", "category": "Makanan Instan", "base_price_range": (2800, 3500), "unit": "pcs"},
    {"prefix": "Pop Mie", "category": "Makanan Instan", "base_price_range": (5000, 7000), "unit": "pcs"},
    {"prefix": "Sarimi", "category": "Makanan Instan", "base_price_range": (2500, 3500), "unit": "pcs"},
    
    # Minuman
    {"prefix": "Aqua", "category": "Minuman", "base_price_range": (3000, 6000), "unit": "botol"},
    {"prefix": "Le Minerale", "category": "Minuman", "base_price_range": (3000, 5000), "unit": "botol"},
    {"prefix": "Teh Botol Sosro", "category": "Minuman", "base_price_range": (5000, 7000), "unit": "botol"},
    {"prefix": "Teh Pucuk", "category": "Minuman", "base_price_range": (4000, 6000), "unit": "botol"},
    {"prefix": "Pocari Sweat", "category": "Minuman", "base_price_range": (7000, 10000), "unit": "botol"},
    {"prefix": "Coca Cola", "category": "Minuman", "base_price_range": (7000, 15000), "unit": "botol"},
    {"prefix": "Sprite", "category": "Minuman", "base_price_range": (7000, 15000), "unit": "botol"},
    {"prefix": "Fanta", "category": "Minuman", "base_price_range": (7000, 15000), "unit": "botol"},
    {"prefix": "Mizone", "category": "Minuman", "base_price_range": (5000, 7000), "unit": "botol"},
    {"prefix": "Good Day", "category": "Minuman", "base_price_range": (5000, 8000), "unit": "botol"},
    {"prefix": "Kopi ABC", "category": "Minuman", "base_price_range": (2000, 3000), "unit": "sachet"},
    {"prefix": "Kopi Kapal Api", "category": "Minuman", "base_price_range": (2000, 3000), "unit": "sachet"},
    {"prefix": "Teh Sariwangi", "category": "Minuman", "base_price_range": (500, 1500), "unit": "sachet"},
    
    # Makanan Ringan
    {"prefix": "Chitato", "category": "Makanan Ringan", "base_price_range": (10000, 15000), "unit": "pcs"},
    {"prefix": "Lays", "category": "Makanan Ringan", "base_price_range": (10000, 15000), "unit": "pcs"},
    {"prefix": "Taro", "category": "Makanan Ringan", "base_price_range": (2000, 5000), "unit": "pcs"},
    {"prefix": "Qtela", "category": "Makanan Ringan", "base_price_range": (7000, 12000), "unit": "pcs"},
    {"prefix": "Oreo", "category": "Makanan Ringan", "base_price_range": (5000, 15000), "unit": "pcs"},
    {"prefix": "Roma Kelapa", "category": "Makanan Ringan", "base_price_range": (3000, 8000), "unit": "pcs"},
    {"prefix": "Biskuat", "category": "Makanan Ringan", "base_price_range": (2000, 5000), "unit": "pcs"},
    {"prefix": "Richeese", "category": "Makanan Ringan", "base_price_range": (2000, 5000), "unit": "pcs"},
    
    # Rokok
    {"prefix": "Gudang Garam", "category": "Rokok", "base_price_range": (25000, 30000), "unit": "bungkus"},
    {"prefix": "Djarum Super", "category": "Rokok", "base_price_range": (20000, 25000), "unit": "bungkus"},
    {"prefix": "Sampoerna Mild", "category": "Rokok", "base_price_range": (28000, 35000), "unit": "bungkus"},
    {"prefix": "Surya Pro", "category": "Rokok", "base_price_range": (18000, 22000), "unit": "bungkus"},
    {"prefix": "LA Lights", "category": "Rokok", "base_price_range": (20000, 25000), "unit": "bungkus"},
    {"prefix": "Marlboro", "category": "Rokok", "base_price_range": (35000, 45000), "unit": "bungkus"},
    
    # Bahan Pokok
    {"prefix": "Beras Premium", "category": "Bahan Pokok", "base_price_range": (14000, 18000), "unit": "kg"},
    {"prefix": "Beras Medium", "category": "Bahan Pokok", "base_price_range": (11000, 14000), "unit": "kg"},
    {"prefix": "Gula Pasir", "category": "Bahan Pokok", "base_price_range": (14000, 17000), "unit": "kg"},
    {"prefix": "Minyak Goreng Bimoli", "category": "Bahan Pokok", "base_price_range": (18000, 25000), "unit": "liter"},
    {"prefix": "Minyak Goreng Tropical", "category": "Bahan Pokok", "base_price_range": (16000, 22000), "unit": "liter"},
    {"prefix": "Tepung Terigu Segitiga", "category": "Bahan Pokok", "base_price_range": (11000, 14000), "unit": "kg"},
    {"prefix": "Garam Kasar", "category": "Bahan Pokok", "base_price_range": (3000, 5000), "unit": "kg"},
    {"prefix": "Telur Ayam", "category": "Bahan Pokok", "base_price_range": (2500, 3500), "unit": "butir"},
    
    # Bumbu Dapur
    {"prefix": "Kecap ABC", "category": "Bumbu Dapur", "base_price_range": (10000, 18000), "unit": "botol"},
    {"prefix": "Kecap Bango", "category": "Bumbu Dapur", "base_price_range": (12000, 22000), "unit": "botol"},
    {"prefix": "Saos Sambal ABC", "category": "Bumbu Dapur", "base_price_range": (8000, 15000), "unit": "botol"},
    {"prefix": "Saos Tomat", "category": "Bumbu Dapur", "base_price_range": (8000, 12000), "unit": "botol"},
    {"prefix": "Royco", "category": "Bumbu Dapur", "base_price_range": (2000, 5000), "unit": "sachet"},
    {"prefix": "Masako", "category": "Bumbu Dapur", "base_price_range": (2000, 5000), "unit": "sachet"},
    {"prefix": "Bawang Goreng", "category": "Bumbu Dapur", "base_price_range": (5000, 15000), "unit": "pcs"},
    
    # Sabun & Toiletries
    {"prefix": "Sabun Lifebuoy", "category": "Toiletries", "base_price_range": (3000, 10000), "unit": "pcs"},
    {"prefix": "Sabun Dettol", "category": "Toiletries", "base_price_range": (8000, 15000), "unit": "pcs"},
    {"prefix": "Sabun Lux", "category": "Toiletries", "base_price_range": (5000, 12000), "unit": "pcs"},
    {"prefix": "Shampoo Pantene", "category": "Toiletries", "base_price_range": (1500, 30000), "unit": "sachet"},
    {"prefix": "Shampoo Clear", "category": "Toiletries", "base_price_range": (1500, 30000), "unit": "sachet"},
    {"prefix": "Shampoo Sunsilk", "category": "Toiletries", "base_price_range": (1500, 30000), "unit": "sachet"},
    {"prefix": "Pasta Gigi Pepsodent", "category": "Toiletries", "base_price_range": (5000, 20000), "unit": "pcs"},
    {"prefix": "Sabun Cuci Rinso", "category": "Toiletries", "base_price_range": (2000, 25000), "unit": "sachet"},
    {"prefix": "Sabun Cuci Daia", "category": "Toiletries", "base_price_range": (2000, 20000), "unit": "sachet"},
    {"prefix": "Pewangi Molto", "category": "Toiletries", "base_price_range": (1500, 15000), "unit": "sachet"},
    
    # Susu
    {"prefix": "Susu Indomilk", "category": "Susu", "base_price_range": (8000, 18000), "unit": "kotak"},
    {"prefix": "Susu Ultra", "category": "Susu", "base_price_range": (7000, 15000), "unit": "kotak"},
    {"prefix": "Susu Frisian Flag", "category": "Susu", "base_price_range": (6000, 20000), "unit": "kotak"},
    {"prefix": "Susu Bear Brand", "category": "Susu", "base_price_range": (10000, 15000), "unit": "kaleng"},
    {"prefix": "SKM Frisian Flag", "category": "Susu", "base_price_range": (8000, 15000), "unit": "kaleng"},
    
    # Gas & Galon
    {"prefix": "Gas LPG 3kg", "category": "Gas & Galon", "base_price_range": (22000, 25000), "unit": "tabung"},
    {"prefix": "Gas LPG 12kg", "category": "Gas & Galon", "base_price_range": (150000, 180000), "unit": "tabung"},
    {"prefix": "Galon Aqua", "category": "Gas & Galon", "base_price_range": (18000, 22000), "unit": "galon"},
    {"prefix": "Galon Cleo", "category": "Gas & Galon", "base_price_range": (15000, 20000), "unit": "galon"},
]

VARIANTS = [
    "Original", "Jumbo", "Mini", "250ml", "500ml", "600ml", "1L", "1.5L", "2L",
    "Pedas", "Spicy", "Rendang", "Ayam Bawang", "Soto", "Kari", "Geprek",
    "Reguler", "Extra", "Double", "Triple", "Isi 12", "Isi 16", "Isi 20",
    "Sachet", "Botol Kecil", "Botol Besar", "Refill", "Pouch",
    "Putih", "Merah", "Kuning", "Hijau", "Biru",
]

CUSTOMER_NAMES = [
    "Bu Tejo", "Pak Bambang", "Bu Sri", "Pak Karno", "Bu Dewi", "Pak Agus",
    "Bu Ratna", "Pak Joko", "Bu Endang", "Pak Wahyu", "Bu Siti", "Pak Hendra",
    "Warung Makan Sederhana", "Warung Nasi Padang", "Warung Kopi",
    "Toko Kelontong Makmur", "Kios Pak Udin", "Warung Bu Yati",
    "Depot Es Teh", "Kantin Sekolah", "Koperasi Desa", "Toko Sembako Jaya",
]


class WarungSeeder:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.api_url = f"{base_url}/api/v1"
        self.token = None
        self.categories = {}
        self.products_created = 0
        self.customers_created = 0
        
    def login(self, email: str, password: str) -> bool:
        """Login and get auth token"""
        try:
            resp = requests.post(
                f"{self.base_url}/auth/login",
                json={"email": email, "password": password},
                timeout=10
            )
            if resp.status_code == 200:
                data = resp.json()
                self.token = data["data"]["access_token"]
                print(f"✓ Logged in as {email}")
                return True
            else:
                print(f"✗ Login failed: {resp.status_code}")
                return False
        except Exception as e:
            print(f"✗ Login error: {e}")
            return False
    
    def get_headers(self) -> dict:
        return {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json"
        }
    
    def create_categories(self):
        """Create categories from product templates"""
        category_names = set(t["category"] for t in PRODUCT_TEMPLATES)
        
        for name in category_names:
            try:
                resp = requests.post(
                    f"{self.api_url}/categories",
                    json={"name": name},
                    headers=self.get_headers(),
                    timeout=10
                )
                if resp.status_code in [200, 201]:
                    data = resp.json()
                    self.categories[name] = data["data"]["id"]
                    print(f"  ✓ Category: {name}")
                elif resp.status_code == 409:  # Already exists
                    # Fetch existing categories
                    pass
            except Exception as e:
                print(f"  ✗ Category {name}: {e}")
        
        # Fetch all existing categories
        try:
            resp = requests.get(f"{self.api_url}/categories", headers=self.get_headers())
            if resp.status_code == 200:
                cats = resp.json().get("data", [])
                for cat in cats:
                    self.categories[cat["name"]] = cat["id"]
        except:
            pass
        
        print(f"✓ {len(self.categories)} categories ready")
    
    def generate_barcode(self) -> str:
        """Generate random EAN-13 barcode"""
        return "888" + "".join(random.choices(string.digits, k=10))
    
    def generate_sku(self, prefix: str) -> str:
        """Generate SKU from prefix"""
        code = "".join(c[0] for c in prefix.split()[:3]).upper()
        return f"{code}-{random.randint(1000, 9999)}"
    
    def create_product(self, template: dict, variant: str) -> bool:
        """Create a single product"""
        name = f"{template['prefix']} {variant}"
        category_id = self.categories.get(template["category"])
        
        base_price = random.randint(*template["base_price_range"])
        # Round to nearest 500
        base_price = round(base_price / 500) * 500
        cost_price = int(base_price * 0.75)  # 25% margin
        
        product_data = {
            "name": name,
            "barcode": self.generate_barcode(),
            "sku": self.generate_sku(name),
            "unit": template["unit"],
            "base_price": base_price,
            "cost_price": cost_price,
            "is_stock_active": True,
            "current_stock": random.randint(10, 200),
            "min_stock_alert": random.randint(5, 20),
        }
        
        if category_id:
            product_data["category_id"] = category_id
        
        # Add pricing tiers for wholesale
        if random.random() < 0.4:  # 40% products have tiers
            product_data["pricing_tiers"] = [
                {"name": "Grosir 10+", "min_quantity": 10, "price": int(base_price * 0.95)},
                {"name": "Grosir 50+", "min_quantity": 50, "price": int(base_price * 0.90)},
            ]
        
        try:
            resp = requests.post(
                f"{self.api_url}/products",
                json=product_data,
                headers=self.get_headers(),
                timeout=15
            )
            if resp.status_code in [200, 201]:
                self.products_created += 1
                return True
            else:
                return False
        except Exception as e:
            return False
    
    def create_products(self, count: int, threads: int = 10):
        """Create products in parallel"""
        print(f"\n📦 Creating {count} products...")
        
        # Generate product list
        products_to_create = []
        templates = PRODUCT_TEMPLATES.copy()
        
        while len(products_to_create) < count:
            template = random.choice(templates)
            variant = random.choice(VARIANTS)
            products_to_create.append((template, variant))
        
        # Create in parallel
        completed = 0
        with ThreadPoolExecutor(max_workers=threads) as executor:
            futures = [
                executor.submit(self.create_product, template, variant)
                for template, variant in products_to_create
            ]
            
            for future in as_completed(futures):
                completed += 1
                if completed % 100 == 0:
                    print(f"  Progress: {completed}/{count} ({self.products_created} created)")
        
        print(f"✓ Created {self.products_created} products")
    
    def create_customer(self, name: str, index: int) -> bool:
        """Create a single customer"""
        phone = f"08{random.randint(1, 9)}{random.randint(10000000, 99999999)}"
        
        customer_data = {
            "name": f"{name} {index}" if index > 0 else name,
            "phone": phone,
            "address": f"Jl. Contoh No. {random.randint(1, 100)}, Jakarta",
            "credit_limit": random.choice([0, 100000, 200000, 500000, 1000000]),
        }
        
        try:
            resp = requests.post(
                f"{self.api_url}/customers",
                json=customer_data,
                headers=self.get_headers(),
                timeout=10
            )
            if resp.status_code in [200, 201]:
                self.customers_created += 1
                return True
            return False
        except:
            return False
    
    def create_customers(self, count: int):
        """Create customers"""
        print(f"\n👥 Creating {count} customers...")
        
        for i in range(count):
            name = random.choice(CUSTOMER_NAMES)
            self.create_customer(name, i)
            
            if (i + 1) % 20 == 0:
                print(f"  Progress: {i + 1}/{count}")
        
        print(f"✓ Created {self.customers_created} customers")
    
    def run(self, products: int = 2000, customers: int = 100):
        """Run the seeding process"""
        print("=" * 50)
        print("🌱 Warung Backend Data Seeder")
        print("=" * 50)
        
        # Login
        if not self.login("admin@warung.com", "password"):
            print("Failed to login. Make sure backend is running and admin user exists.")
            return False
        
        # Create categories
        print("\n📁 Creating categories...")
        self.create_categories()
        
        # Create products
        self.create_products(products)
        
        # Create customers
        self.create_customers(customers)
        
        # Summary
        print("\n" + "=" * 50)
        print("📊 Seeding Complete!")
        print(f"   Products: {self.products_created}")
        print(f"   Customers: {self.customers_created}")
        print(f"   Categories: {len(self.categories)}")
        print("=" * 50)
        
        return True


def main():
    parser = argparse.ArgumentParser(description="Seed Warung Backend with test data")
    parser.add_argument("--url", default="http://localhost:8080", help="Backend URL")
    parser.add_argument("--products", type=int, default=2000, help="Number of products")
    parser.add_argument("--customers", type=int, default=100, help="Number of customers")
    
    args = parser.parse_args()
    
    seeder = WarungSeeder(args.url)
    success = seeder.run(products=args.products, customers=args.customers)
    
    exit(0 if success else 1)


if __name__ == "__main__":
    main()

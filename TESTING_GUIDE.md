# Performance & Load Testing Guide

## Ikhtisar

Dokumen ini menjelaskan cara menjalankan berbagai jenis test untuk memastikan backend warung dapat menangani:
- **1000 pengguna** secara bersamaan
- **2000+ produk** dalam database
- Performa optimal di bawah beban tinggi

---

## 📋 Prasyarat

### 1. Install k6 (Load Testing Tool)

```bash
# Ubuntu/Debian
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# macOS
brew install k6

# Atau download binary: https://k6.io/docs/getting-started/installation/
```

### 2. Python 3 untuk Script

```bash
pip install requests
```

### 3. Jalankan Backend & Database

```bash
# Terminal 1: Database
docker-compose up -d postgres redis

# Terminal 2: Backend
make run
```

---

## 🌱 Step 1: Seed Data (2000 Produk)

```bash
# Seed 2000 products dan 100 customers
python scripts/seed_data.py --products 2000 --customers 100

# Custom URL jika backend bukan di localhost
python scripts/seed_data.py --url http://192.168.1.100:8080 --products 2000
```

**Output yang diharapkan:**
```
🌱 Warung Backend Data Seeder
✓ Logged in as admin@warung.com
✓ 10 categories ready
  Progress: 2000/2000 (1987 created)
✓ Created 1987 products
✓ Created 98 customers
```

---

## 🧪 Step 2: Quick Stress Test (3 menit)

Test cepat untuk validasi sebelum load test penuh:

```bash
k6 run tests/load/quick_test.js
```

**Target:**
- 100 VUs (Virtual Users)
- Response time p95 < 2s
- Error rate < 10%

---

## 🚀 Step 3: Full Load Test (1000 Users)

```bash
# Test utama - 22 menit total
k6 run tests/load/load_test.js

# Dengan custom URL
BASE_URL=http://192.168.1.100:8080 k6 run tests/load/load_test.js

# Output ke JSON untuk analisis
k6 run --out json=results.json tests/load/load_test.js
```

**Skenario yang ditest:**
1. **Ramp Up**: 0 → 1000 users dalam 10 menit
2. **Sustained Load**: 1000 users selama 10 menit
3. **Spike Test**: Lonjakan mendadak 500 users

**Target Performa:**
| Metrik | Target |
|--------|--------|
| Response time p95 | < 2 detik |
| Response time p99 | < 5 detik |
| Error rate | < 5% |
| Product list | < 1 detik (p95) |
| Checkout | < 3 detik (p95) |

---

## 💾 Step 4: Database Stress Test

Test khusus untuk database dengan workload berat:

```bash
k6 run tests/load/db_stress_test.js
```

**Workload Mix:**
- 70% read operations (list, search, get)
- 30% write operations (create customer, calculate cart)

---

## ✅ Step 5: API Integration Tests

Test semua endpoint berfungsi dengan benar:

```bash
python tests/integration/run_api_tests.py
```

**Test yang dijalankan:**
- Auth (login, refresh token)
- Products (CRUD, search, pagination)
- Customers (CRUD, debt operations)
- Transactions (calculate, checkout)
- Reports (dashboard, daily)
- Kasbon (report)

---

## 📊 Step 6: Go Benchmark Tests

Test performa query database secara langsung:

```bash
# Pindah ke direktori test
cd tests/benchmark

# Jalankan benchmark
go test -bench=. -benchmem

# Jalankan dengan durasi lebih lama
go test -bench=. -benchtime=10s -benchmem
```

**Output contoh:**
```
BenchmarkProductRepository_List-8          5000        250000 ns/op      15000 B/op       200 allocs/op
BenchmarkProductRepository_Search-8        3000        400000 ns/op      20000 B/op       250 allocs/op
```

---

## 📈 Interpretasi Hasil

### K6 Output

```
     ✓ product list status is 200

     checks.........................: 99.85% ✓ 125847    ✗ 189
     data_received..................: 850 MB  6.5 MB/s
     data_sent......................: 45 MB   350 kB/s
     http_req_duration..............: avg=125ms   min=15ms  med=95ms  max=4.5s  p(95)=350ms  p(99)=1.2s
     http_reqs......................: 126036  965/s
     vus............................: 1000    min=0      max=1000
```

### Kriteria PASS:
- ✅ `checks > 95%` — Validasi respons berhasil
- ✅ `http_req_duration p(95) < 2s` — Respons cepat
- ✅ `http_req_failed < 5%` — Error rate rendah
- ✅ `http_reqs > 500/s` — Throughput tinggi

### Kriteria WARNING:
- ⚠️ `p(95) > 2s` — Perlu optimasi
- ⚠️ `http_req_failed > 5%` — Ada masalah

### Kriteria FAIL:
- ❌ `p(95) > 5s` — Terlalu lambat
- ❌ `http_req_failed > 20%` — Banyak error
- ❌ Backend crash — Perlu perbaikan

---

## 🔧 Troubleshooting

### "Connection refused" / "EOF"
- Backend tidak running atau overloaded
- Tambah resource: `ulimit -n 65535`

### "Too many open files"
```bash
# Tingkatkan limit
ulimit -n 65535
```

### Respons Lambat
1. Pastikan database indexes sudah benar
2. Cek query yang berat di logs
3. Aktifkan Redis cache

### Out of Memory
```bash
# Kurangi jumlah VUs
k6 run -u 500 tests/load/load_test.js
```

---

## 📁 Struktur File Test

```
tests/
├── load/
│   ├── load_test.js        # Full load test 1000 users
│   ├── quick_test.js       # Quick 3-minute stress test
│   └── db_stress_test.js   # Database-focused stress test
├── benchmark/
│   └── repository_test.go  # Go benchmark tests
└── integration/
    ├── service_test.go     # Service layer unit tests
    └── run_api_tests.py    # API integration tests

scripts/
└── seed_data.py            # Seed 2000 products
```

---

## Quick Commands

```bash
# Seed data
python scripts/seed_data.py --products 2000

# Quick stress test
k6 run tests/load/quick_test.js

# Full load test
k6 run tests/load/load_test.js

# API integration tests
python tests/integration/run_api_tests.py

# Go benchmarks
go test -bench=. ./tests/benchmark/
```

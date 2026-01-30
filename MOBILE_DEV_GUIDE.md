# Panduan Pengembangan Aplikasi Mobile Warung Kelontong (WarungOS)

## Ringkasan Backend

```
Base URL: http://your-server:8080
API Version: /api/v1
Auth: Bearer Token (JWT)
```

### 📊 Performance (Hasil Load Test)
| Metric | Value |
|--------|-------|
| Throughput | **~200 requests/second** |
| Response Time p95 | **<12ms** |
| Max Concurrent Users | **200+ kasir** |
| Rate Limit | **1000 requests/minute** |

### 🔐 Test Credentials (Development)
| Role | Email | Password |
|------|-------|----------|
| Admin | `admin@warung.com` | `password` |
| Cashier | `cashier@warung.com` | `password` |
| Inventory | `inventory@warung.com` | `password` |

---

## 1. Autentikasi (Authentication Flow)

### 1.1 Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "kasir@warung.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": { "id": "uuid", "name": "Kasir 1", "role": "cashier" },
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

### 1.2 Refresh Token
```http
POST /auth/refresh
{
  "refresh_token": "eyJhbGc..."
}
```

### 1.3 Register User (Admin Only)

| Role | Akses |
|------|-------|
| `admin` | Semua fitur |
| `cashier` | POS, transaksi, kasbon, pelanggan |
| `inventory` | Stok, restock, stock opname |

---

## 2. Flow Aplikasi POS (Point of Sale)

### 2.1 Scan Produk (Barcode)
```http
GET /api/v1/products/search?barcode=8886008111105
Authorization: Bearer {token}
```

### 2.2 List Produk (Search/Browse)
```http
GET /api/v1/products?search=indomie&page=1&per_page=20
```

### 2.3 Hitung Keranjang (Preview)
```http
POST /api/v1/transactions/calculate
{
  "items": [
    { "product_id": "uuid-1", "quantity": 10 },
    { "product_id": "uuid-2", "quantity": 2 }
  ]
}
```

**Response:**
```json
{
  "items": [
    {
      "product_id": "uuid-1",
      "product_name": "Indomie Goreng",
      "quantity": 10,
      "unit_price": 3000,
      "tier_name": "Grosir 10+",
      "subtotal": 30000,
      "is_available": true
    }
  ],
  "subtotal": 35000
}
```

### 2.4 Checkout (Buat Transaksi)
```http
POST /api/v1/transactions
{
  "items": [
    { "product_id": "uuid", "quantity": 10 }
  ],
  "payment_method": "cash",
  "amount_paid": 100000,
  "customer_id": null,
  "cashier_name": "Kasir 1"
}
```

**Payment Methods:** `cash`, `kasbon`, `qris`, `transfer`

---

## 3. Manajemen Pelanggan & Kasbon

### 3.1 Daftar Pelanggan dengan Hutang
```http
GET /api/v1/customers/with-debt
```

### 3.2 Detail Kasbon Pelanggan
```http
GET /api/v1/kasbon/customers/{customer_id}/summary
```

**Response:**
```json
{
  "customer_name": "Bu Tejo",
  "current_balance": 150000,
  "credit_limit": 500000,
  "remaining_credit": 350000,
  "total_debt": 200000,
  "total_payment": 50000
}
```

### 3.3 Catat Pembayaran Kasbon
```http
POST /api/v1/kasbon/customers/{customer_id}/payments
{
  "amount": 50000,
  "notes": "Bayar cicilan"
}
```

### 3.4 Download Tagihan (PDF)
```http
GET /api/v1/kasbon/customers/{customer_id}/billing/pdf
```

---

## 4. Manajemen Stok & Inventory

### 4.1 Cek Stok Menipis
```http
GET /api/v1/products/low-stock
```

### 4.2 Restock (Tambah Stok)
```http
POST /api/v1/inventory/restock
{
  "product_id": "uuid",
  "quantity": 100,
  "cost_per_unit": 3000,
  "notes": "Beli dari Toko ABC"
}
```

### 4.3 Stock Opname (Audit Stok)

```http
# 1. Mulai sesi
POST /api/v1/stock-opname/sessions
{ "notes": "Stock opname bulanan" }

# 2. Catat hitungan
POST /api/v1/stock-opname/sessions/{session_id}/items
{ "product_id": "uuid", "physical_count": 95 }

# 3. Finalisasi
POST /api/v1/stock-opname/sessions/{session_id}/finalize
```

### 4.4 Download PDF Restock List
```http
GET /api/v1/inventory/restock-list/pdf
```

---

## 5. Laporan & Dashboard

### 5.1 Dashboard Harian
```http
GET /api/v1/reports/dashboard
```

**Response:**
```json
{
  "today_sales": 1500000,
  "today_transactions": 45,
  "today_profit": 225000,
  "outstanding_kasbon": 750000,
  "low_stock_count": 12
}
```

### 5.2 Laporan Penjualan Harian
```http
GET /api/v1/reports/daily?date=2026-01-30
```

---

## 6. Fitur Khusus Warung

### 6.1 Konsinyasi (Titip Jual)
```http
# List penitip
GET /api/v1/consignors

# Buat penitip baru
POST /api/v1/consignors
{
  "name": "Bu Tejo Kue Basah",
  "commission_rate": 10.0,
  "phone": "081234567890"
}
```

### 6.2 Gas & Galon (Refillable)
```http
# Cek stok tabung
GET /api/v1/refillables

# Adjust stok (tukar kosong ↔ isi)
POST /api/v1/refillables/adjust
{
  "product_id": "uuid-gas-isi",
  "adjustment_type": "exchange",
  "quantity": 5
}
```

### 6.3 Cart Tertunda (Hold/Resume)
```http
# Simpan cart sementara
POST /api/v1/pos/held-carts
{ "customer_name": "Warteg Pak Min", "items": [...] }

# Ambil kembali
POST /api/v1/pos/held-carts/{id}/resume
```

---

## 7. Response Format

### Success
```json
{
  "success": true,
  "message": "Description",
  "data": { ... }
}
```

### Error
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Name is required",
    "details": { "name": "Name is required" }
  }
}
```

### HTTP Status Codes
| Code | Meaning | Action di Mobile |
|------|---------|------------------|
| `200` | Success | Tampilkan data |
| `201` | Created | Transaksi berhasil |
| `400` | Bad Request | Validasi error, tampilkan pesan |
| `401` | Unauthorized | Redirect ke login |
| `403` | Forbidden | Role tidak punya akses |
| `404` | Not Found | Item tidak ditemukan |
| `429` | Rate Limited | Retry setelah 1 detik |
| `500` | Server Error | Tampilkan "Coba lagi" |

### Error Handling (Mobile Best Practice)
```javascript
// Contoh error handling
async function apiCall(endpoint) {
  try {
    const response = await fetch(API_URL + endpoint, { headers });
    
    if (response.status === 401) {
      // Token expired, refresh atau logout
      await refreshToken();
      return apiCall(endpoint); // Retry
    }
    
    if (response.status === 429) {
      // Rate limited, tunggu lalu retry
      await sleep(1000);
      return apiCall(endpoint);
    }
    
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error?.message || 'Terjadi kesalahan');
    }
    
    return response.json();
  } catch (e) {
    // Network error - tampilkan offline mode
    showOfflineMessage();
  }
}
```

### Pagination
```json
{
  "data": [...],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total_items": 150,
    "total_pages": 8
  }
}
```

---

## 8. Kategori Produk

### 8.1 List Semua Kategori
```http
GET /api/v1/categories
```

**Response:**
```json
{
  "data": [
    { "id": "uuid", "name": "Makanan", "description": "Makanan ringan" },
    { "id": "uuid", "name": "Minuman", "description": "Minuman dingin/panas" },
    { "id": "uuid", "name": "Rokok", "description": "Rokok dan vape" }
  ]
}
```

### 8.2 Filter Produk by Kategori
```http
GET /api/v1/products?category_id={category_uuid}&page=1&per_page=20
```

---

## 9. Prioritas Implementasi Mobile

### Phase 1: MVP (2-3 minggu)
1. ✅ Login/Logout
2. ✅ Scan barcode → tambah ke cart
3. ✅ Checkout cash
4. ✅ List produk dengan search

### Phase 2: Kasbon (1-2 minggu)
1. ✅ Checkout dengan kasbon
2. ✅ List pelanggan hutang
3. ✅ Catat pembayaran

### Phase 3: Inventory (1-2 minggu)
1. ✅ Low stock alerts
2. ✅ Restock produk
3. ✅ Stock opname

### Phase 4: Reports (1 minggu)
1. ✅ Dashboard harian
2. ✅ Laporan penjualan
3. ✅ Export PDF

---

## 10. Tips UI/UX untuk Warung

| Screen | Tips |
|--------|------|
| **POS** | Font besar, tombol besar untuk touch. Tampilkan gambar produk |
| **Keranjang** | Swipe untuk hapus item. +/- button untuk quantity |
| **Kasbon** | Warna merah untuk hutang. Hijau untuk lunas |
| **Low Stock** | Badge notifikasi di home screen |
| **Barcode** | Suara "beep" saat sukses scan |

---

## 11. Offline Support (Recommended)

### Data yang Harus Di-cache Lokal:
| Data | Alasan |
|------|--------|
| Produk (list + harga) | Bisa scan & hitung offline |
| Kategori | Filter produk offline |
| Pelanggan | Pilih pelanggan kasbon offline |
| Pending transactions | Queue transaksi saat offline |

### Sync Strategy:
```
1. Pull produk & pelanggan saat app dibuka
2. Simpan transaksi di SQLite/Room saat offline
3. Sync pending transactions saat online
4. Background sync setiap 5 menit
```

### Conflict Resolution:
- Server always wins untuk data master (produk, harga)
- Queue transactions di local, kirim saat online
- Show "pending sync" indicator di UI

---

## 12. Testing Endpoints

### Postman Collection
```
/warung-backend.postman_collection.json
```
Import ke Postman/Insomnia untuk testing semua endpoint.

### API Integration Test
```bash
# Run all API tests
python tests/integration/run_api_tests.py

# Run load test (5 min)
make test-prod-quick
```

### Health Check
```http
GET /health
```
Gunakan untuk check koneksi di mobile app.

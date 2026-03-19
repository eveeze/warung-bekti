# Panduan API Lengkap - Warung Backend

**Total Endpoints: 71**

```
Base URL: http://your-server:8080
API Prefix: /api/v1
Auth: Bearer Token (JWT)
Header: Authorization: Bearer <access_token>
```

---

## 📋 Daftar Isi

1. [Health Check](#1-health-check-3-endpoints)
2. [Authentication](#2-authentication-3-endpoints)
3. [Products](#3-products-11-endpoints)
4. [Customers](#4-customers-6-endpoints)
5. [Kasbon](#5-kasbon-4-endpoints)
6. [Transactions](#6-transactions-5-endpoints)
7. [Inventory](#7-inventory-6-endpoints)
8. [Reports](#8-reports-4-endpoints)
9. [Payments](#9-payments-4-endpoints)
10. [Stock Opname](#10-stock-opname-9-endpoints)
11. [Cash Flow](#11-cash-flow-6-endpoints)
12. [POS Features](#12-pos-features-6-endpoints)
13. [Consignment](#13-consignment-3-endpoints)
14. [Refillables](#14-refillables-2-endpoints)
15. [Image Handling & Development Setup](#15-image-handling--development-setup)

---

## 1. Health Check (3 endpoints)

### 1.1 Health

```
GET /health
```

No auth required.

### 1.2 Ready

```
GET /ready
```

No auth required.

### 1.3 Live

```
GET /live
```

No auth required.

---

## 2. Authentication (3 endpoints)

### 2.1 Login

```
POST /auth/login
```

**Request Body:**

```json
{
  "email": "admin@warung.com",
  "password": "password"
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "user": {
      "id": "uuid",
      "name": "Admin",
      "email": "admin@warung.com",
      "role": "admin"
    }
  }
}
```

### 2.2 Register

```
POST /auth/register
```

**Request Body:**

```json
{
  "name": "Nama User",
  "email": "user@warung.com",
  "password": "password123",
  "role": "cashier"
}
```

> Roles: `cashier`, `inventory` (admin tidak bisa didaftarkan via API)

### 2.3 Refresh Token

```
POST /auth/refresh
```

**Request Body:**

```json
{
  "refresh_token": "eyJhbGc..."
}
```

---

## 3. Products (11 endpoints)

### 3.1 List Products

```
GET /api/v1/products
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `search` | string | Cari by name/sku |
| `category_id` | uuid | Filter kategori |
| `page` | int | Default: 1 |
| `per_page` | int | Default: 20 |
| `sort_by` | string | name, base_price, created_at |
| `sort_order` | string | asc, desc |
| `low_stock` | bool | true untuk filter low stock |

### 3.2 Create Product

```
POST /api/v1/products
```

**Role:** Admin only
**Content-Type:** `multipart/form-data`

> **📝 Note untuk Mobile Dev:**
> Request ini **WAJIB** menggunakan `multipart/form-data` karena ada upload file.
>
> - Field `data`: Berisi JSON String dari detail produk. Jangan kirim JSON raw body terpisah.
> - Field `image`: Berisi binary file gambar (JPG/PNG).

**Form Fields:**

- `data`: JSON string (Contoh: `{"name": "Produk A", ...}`)
- `image`: File (optional, max 5MB, JPG/PNG only)

**JSON (`data` field content):**

```json
{
  "name": "Indomie Goreng",
  "barcode": "8886008101025",
  "sku": "IND-GRG-001",
  "description": "Mie instan goreng",
  "category_id": "uuid-kategori",
  "unit": "pcs",
  "base_price": 3500,
  "cost_price": 3100,
  "is_stock_active": true,
  "current_stock": 100,
  "min_stock_alert": 10,
  "max_stock": 500,
  "is_refillable": false,
  "pricing_tiers": [
    {
      "name": "Grosir 10+",
      "min_quantity": 10,
      "max_quantity": 39,
      "price": 3300
    },
    {
      "name": "Karton 40+",
      "min_quantity": 40,
      "price": 3150
    }
  ]
}
```

### 3.3 Get Product by Barcode

```
GET /api/v1/products/search?barcode=8886008101025
```

### 3.4 Get Low Stock Products

```
GET /api/v1/products/low-stock
```

### 3.5 Get Product Detail

```
GET /api/v1/products/{id}
```

### 3.6 Update Product

```
PUT /api/v1/products/{id}
```

**Role:** Admin only
**Content-Type:** `multipart/form-data` atau `application/json`

**Form Fields (jika ada gambar):**

- `data`: JSON string
- `image`: File

**JSON Body (jika tidak ada gambar, semua field optional):**

```json
{
  "name": "Indomie Goreng Special",
  "base_price": 3600,
  "cost_price": 3200,
  "min_stock_alert": 15,
  "is_active": true
}
```

### 3.7 Delete Product

```
DELETE /api/v1/products/{id}
```

**Role:** Admin only

### 3.8 Add Pricing Tier

```
POST /api/v1/products/{id}/pricing-tiers
```

**Role:** Admin only

**Request Body:**

```json
{
  "name": "Promo Ramadhan",
  "min_quantity": 5,
  "max_quantity": 20,
  "price": 3400
}
```

### 3.9 Update Pricing Tier

```
PUT /api/v1/products/{id}/pricing-tiers/{tierId}
```

**Role:** Admin only

**Request Body:**

```json
{
  "name": "Promo Updated",
  "min_quantity": 5,
  "price": 3350
}
```

### 3.10 Delete Pricing Tier

```
DELETE /api/v1/products/{id}/pricing-tiers/{tierId}
```

**Role:** Admin only

---

## 4. Customers (6 endpoints)

### 4.1 List Customers

```
GET /api/v1/customers
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `search` | string | Cari by name/phone |
| `has_debt` | bool | true untuk filter yang punya hutang |
| `page` | int | Default: 1 |
| `per_page` | int | Default: 20 |

### 4.2 Create Customer

```
POST /api/v1/customers
```

**Request Body:**

```json
{
  "name": "Bu Tejo",
  "phone": "081234567890",
  "address": "Jl. Mawar No. 5, RT 01",
  "notes": "Pelanggan tetap",
  "credit_limit": 500000
}
```

### 4.3 Get Customers with Debt

```
GET /api/v1/customers/with-debt
```

### 4.4 Get Customer Detail

```
GET /api/v1/customers/{id}
```

### 4.5 Update Customer

```
PUT /api/v1/customers/{id}
```

**Request Body:**

```json
{
  "name": "Bu Tejo Updated",
  "phone": "081234567899",
  "address": "Jl. Melati No. 10",
  "notes": "Update catatan",
  "credit_limit": 750000,
  "is_active": true
}
```

### 4.6 Delete Customer

```
DELETE /api/v1/customers/{id}
```

**Role:** Admin only

---

## 5. Kasbon (4 endpoints)

### 5.1 Get Kasbon History

```
GET /api/v1/kasbon/customers/{id}
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `page` | int | Default: 1 |
| `per_page` | int | Default: 20 |
| `type` | string | debt, payment |

### 5.2 Get Kasbon Summary

```
GET /api/v1/kasbon/customers/{id}/summary
```

**Response:**

```json
{
  "data": {
    "customer_id": "uuid",
    "customer_name": "Bu Tejo",
    "total_debt": 200000,
    "total_payment": 50000,
    "current_balance": 150000,
    "credit_limit": 500000,
    "remaining_credit": 350000
  }
}
```

### 5.3 Download Billing PDF

```
GET /api/v1/kasbon/customers/{id}/billing/pdf
```

Returns PDF file.

### 5.4 Record Kasbon Payment

```
POST /api/v1/kasbon/customers/{id}/payments
```

**Request Body:**

```json
{
  "amount": 50000,
  "notes": "Bayar cicilan via transfer BCA",
  "created_by": "Kasir 1"
}
```

---

## 6. Transactions (5 endpoints)

### 6.1 List Transactions

```
GET /api/v1/transactions
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `customer_id` | uuid | Filter by customer |
| `status` | string | pending, completed, cancelled, refunded |
| `payment_method` | string | cash, kasbon, transfer, qris |
| `date_from` | date | Format: 2026-01-01 |
| `date_to` | date | Format: 2026-01-31 |
| `page` | int | Default: 1 |
| `per_page` | int | Default: 20 |

### 6.2 Create Transaction (Checkout)

```
POST /api/v1/transactions
```

**Request Body:**

```json
{
  "items": [
    {
      "product_id": "uuid-product-1",
      "quantity": 10,
      "discount_amount": 0,
      "notes": "Minta yang baru"
    },
    {
      "product_id": "uuid-product-2",
      "quantity": 2
    }
  ],
  "customer_id": "uuid-customer",
  "discount_amount": 5000,
  "tax_amount": 0,
  "payment_method": "cash",
  "amount_paid": 100000,
  "notes": "Pembeli langganan",
  "cashier_name": "Kasir 1"
}
```

**Payment Methods:** `cash`, `kasbon`, `transfer`, `qris`

### 6.3 Calculate Cart (Preview)

```
POST /api/v1/transactions/calculate
```

**Request Body:**

```json
{
  "items": [
    { "product_id": "uuid-1", "quantity": 10 },
    { "product_id": "uuid-2", "quantity": 5 }
  ]
}
```

**Response:**

```json
{
  "data": {
    "items": [
      {
        "product_id": "uuid-1",
        "product_name": "Indomie Goreng",
        "quantity": 10,
        "unit": "pcs",
        "unit_price": 3300,
        "tier_name": "Grosir 10+",
        "subtotal": 33000,
        "is_available": true,
        "available_qty": 100
      }
    ],
    "subtotal": 58000
  }
}
```

### 6.4 Get Transaction Detail

```
GET /api/v1/transactions/{id}
```

### 6.5 Cancel Transaction

```
POST /api/v1/transactions/{id}/cancel
```

No request body needed.

---

## 7. Inventory (6 endpoints)

### 7.1 Restock Product

```
POST /api/v1/inventory/restock
```

**Request Body:**

```json
{
  "product_id": "uuid-product",
  "quantity": 100,
  "cost_per_unit": 3000,
  "notes": "Restock dari Supplier ABC",
  "created_by": "Staff Gudang"
}
```

### 7.2 Manual Stock Adjustment

```
POST /api/v1/inventory/adjust
```

**Role:** Admin only

**Request Body:**

```json
{
  "product_id": "uuid-product",
  "quantity": -5,
  "reason": "Barang expired/rusak",
  "created_by": "Admin"
}
```

> Quantity bisa negatif (kurangi) atau positif (tambah)

### 7.3 Get Low Stock Products

```
GET /api/v1/inventory/low-stock
```

### 7.4 Get Inventory Report

```
GET /api/v1/inventory/report
```

**Response:**

```json
{
  "data": {
    "total_products": 150,
    "total_stock_value": 25000000,
    "low_stock_count": 12,
    "out_of_stock_count": 3,
    "low_stock_products": [...]
  }
}
```

### 7.5 Download Restock List PDF

```
GET /api/v1/inventory/restock-list/pdf
```

Returns PDF file.

### 7.6 Get Stock Movements

```
GET /api/v1/inventory/{productId}/movements
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `page` | int | Default: 1 |
| `per_page` | int | Default: 20 |
| `type` | string | initial, purchase, sale, adjustment, return, damage |

---

## 8. Reports (4 endpoints)

### 8.1 Get Daily Report

```
GET /api/v1/reports/daily
```

**Role:** Admin only

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `date` | string | Format: 2026-01-31 (default: today) |

**Response:**

```json
{
  "data": {
    "date": "2026-01-31",
    "total_sales": 1500000,
    "total_transactions": 45,
    "estimated_profit": 225000
  }
}
```

### 8.2 Get Kasbon Report

```
GET /api/v1/reports/kasbon
```

**Role:** Admin only

**Response:**

```json
{
  "data": {
    "total_outstanding": 750000,
    "total_customers": 50,
    "customers_with_debt": 15,
    "summaries": [...]
  }
}
```

### 8.3 Get Inventory Report

```
GET /api/v1/reports/inventory
```

**Role:** Admin only

### 8.4 Get Dashboard

```
GET /api/v1/reports/dashboard
```

**Role:** Admin only

**Response:**

```json
{
  "data": {
    "today": {
      "date": "2026-01-31",
      "total_sales": 1500000,
      "total_transactions": 45,
      "estimated_profit": 225000
    },
    "total_outstanding_kasbon": 750000,
    "low_stock_count": 12,
    "out_of_stock_count": 3
  }
}
```

---

## 9. Payments (4 endpoints)

### 9.1 Generate Snap Token (Midtrans)

```
POST /api/v1/payments/snap
```

**Request Body:**

```json
{
  "transaction_id": "uuid-transaction",
  "customer_name": "Budi",
  "customer_email": "budi@email.com",
  "customer_phone": "081234567890"
}
```

**Response:**

```json
{
  "data": {
    "token": "snap-token-xxx",
    "redirect_url": "https://app.midtrans.com/snap/v2/..."
  }
}
```

### 9.2 Payment Notification (Webhook)

```
POST /api/v1/payments/notification
```

**Auth:** Public (no auth required)

**Request Body:** Midtrans callback payload (handled automatically)

### 9.3 Manual Verify Payment

```
POST /api/v1/payments/{paymentId}/manual-verify
```

**Role:** Admin only

No request body needed.

### 9.4 Get Payment by Transaction

```
GET /api/v1/payments/transaction/{transactionId}
```

---

## 10. Stock Opname (9 endpoints)

### 10.1 Start Session

```
POST /api/v1/stock-opname/sessions
```

**Request Body:**

```json
{
  "notes": "Stock opname bulanan Januari 2026"
}
```

### 10.2 List Sessions

```
GET /api/v1/stock-opname/sessions
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `page` | int | Default: 1 |
| `per_page` | int | Default: 20 |
| `status` | string | draft, in_progress, completed, cancelled |

### 10.3 Get Session Detail

```
GET /api/v1/stock-opname/sessions/{id}
```

### 10.4 Record Count

```
POST /api/v1/stock-opname/sessions/{id}/items
```

**Request Body:**

```json
{
  "product_id": "uuid-product",
  "physical_count": 95,
  "notes": "Ada 5 unit rusak tidak dihitung"
}
```

### 10.5 Finalize Session

```
POST /api/v1/stock-opname/sessions/{id}/finalize
```

**Request Body:**

```json
{
  "apply_adjustments": true
}
```

> Jika `apply_adjustments: true`, sistem akan otomatis menyesuaikan stok berdasarkan hasil opname.

### 10.6 Get Variance Report

```
GET /api/v1/stock-opname/sessions/{id}/variance
```

**Response:**

```json
{
  "data": {
    "session_id": "uuid",
    "session_code": "OPNAME-202601-001",
    "total_products": 50,
    "total_variance": 15,
    "total_loss_value": 150000,
    "total_gain_value": 25000,
    "net_value": -125000,
    "items": [...]
  }
}
```

### 10.7 Cancel Session

```
POST /api/v1/stock-opname/sessions/{id}/cancel
```

No request body needed.

### 10.8 Get Shopping List

```
GET /api/v1/stock-opname/shopping-list
```

Auto-generated restock suggestions based on low stock.

### 10.9 Get Near Expiry Report

```
GET /api/v1/stock-opname/near-expiry
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `days` | int | Default: 30 (hari ke depan) |

---

## 11. Cash Flow (6 endpoints)

### 11.1 Open Drawer

```
POST /api/v1/cashflow/drawer/open
```

**Request Body:**

```json
{
  "opening_balance": 200000,
  "notes": "Modal awal shift pagi"
}
```

### 11.2 Close Drawer

```
POST /api/v1/cashflow/drawer/close
```

**Request Body:**

```json
{
  "session_id": "uuid-session",
  "closing_balance": 1500000,
  "notes": "Tutup shift, uang disetor"
}
```

### 11.3 Get Current Session

```
GET /api/v1/cashflow/drawer/current
```

### 11.4 Get Categories

```
GET /api/v1/cashflow/categories
```

### 11.5 Record Cash Flow

```
POST /api/v1/cashflow
```

**Request Body:**

```json
{
  "category_id": "uuid-category",
  "type": "expense",
  "amount": 25000,
  "description": "Beli gas untuk kompor"
}
```

**Types:** `income`, `expense`

### 11.6 List Cash Flows

```
GET /api/v1/cashflow
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `session_id` | uuid | Filter by session |
| `type` | string | income, expense |
| `date_from` | datetime | RFC3339 format |
| `page` | int | Default: 1 |
| `per_page` | int | Default: 20 |

---

## 12. POS Features (6 endpoints)

### 12.1 Hold Cart

```
POST /api/v1/pos/held-carts
```

**Request Body:**

```json
{
  "customer_id": "uuid-customer",
  "notes": "Customer ambil uang dulu",
  "items": [
    {
      "product_id": "uuid-product",
      "quantity": 5,
      "notes": "Item note"
    }
  ]
}
```

### 12.2 List Held Carts

```
GET /api/v1/pos/held-carts
```

### 12.3 Get Held Cart Detail

```
GET /api/v1/pos/held-carts/{id}
```

### 12.4 Resume Cart

```
POST /api/v1/pos/held-carts/{id}/resume
```

No request body needed.

### 12.5 Discard Cart

```
POST /api/v1/pos/held-carts/{id}/discard
```

No request body needed.

### 12.6 Create Refund

```
POST /api/v1/pos/refunds
```

**Request Body:**

```json
{
  "transaction_id": "uuid-transaction",
  "refund_method": "cash",
  "reason": "Salah beli produk",
  "notes": "Customer minta refund",
  "items": [
    {
      "transaction_item_id": "uuid-item",
      "quantity": 1,
      "reason": "Produk tidak sesuai",
      "restock": true
    }
  ]
}
```

**Refund Methods:** `cash`, `transfer`

---

## 13. Consignment (3 endpoints)

### 13.1 Create Consignor

```
POST /api/v1/consignors
```

**Role:** Admin only

**Request Body:**

```json
{
  "name": "Bu Tejo Kue Basah",
  "phone": "081234567890",
  "address": "Jl. Pasar No. 10",
  "bank_account": "1234567890",
  "bank_name": "BCA",
  "notes": "Penitip kue harian"
}
```

### 13.2 List Consignors

```
GET /api/v1/consignors
```

**Role:** Admin only

### 13.3 Update Consignor

```
PUT /api/v1/consignors/{id}
```

**Role:** Admin only

**Request Body:**

```json
{
  "name": "Bu Tejo Updated",
  "phone": "081234567899",
  "address": "Jl. Pasar No. 20",
  "bank_account": "0987654321",
  "bank_name": "Mandiri",
  "notes": "Update catatan",
  "is_active": true
}
```

---

## 14. Refillables (2 endpoints)

### 14.1 Get Containers

```
GET /api/v1/refillables
```

**Response:**

```json
{
  "data": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "container_type": "gas_3kg",
      "empty_count": 10,
      "full_count": 25,
      "product": {...}
    }
  ]
}
```

### 14.2 Adjust Container Stock

```
POST /api/v1/refillables/adjust
```

**Request Body:**

```json
{
  "container_id": "uuid-container",
  "type": "sale_exchange",
  "empty_change": 1,
  "full_change": -1,
  "notes": "Tukar tabung kosong dengan isi"
}
```

**Adjustment Types:**

- `sale_exchange`: Jual isi, terima kosong
- `restock_exchange`: Tukar kosong dengan isi dari supplier
- `purchase_empty`: Beli tabung kosong
- `purchase_full`: Beli tabung isi
- `return_empty`: Customer kembalikan tabung kosong
- `adjustment`: Koreksi manual

---

## 15. Image Handling & Development Setup

### ⚙️ Konfigurasi URL MinIO

**PENTING:** Secara default, endpoint MinIO di database mungkin tercatat sebagai `http://localhost:9000/...`.
Ini akan menyebabkan masalah pada aplikasi **Mobile** (Android/iOS) atau **Emulator** karena `localhost` di HP merujuk ke device itu sendiri, bukan server backend.

**Solusi untuk Developer Mobile:**

1.  **Ganti Environment Variable Backend (Recommended):**
    Saat menjalankan backend, set `MINIO_ENDPOINT` ke Local IP adress laptop/server Anda.
    Contoh (Linux/Mac):
    ```bash
    export MINIO_ENDPOINT="192.168.1.x:9000"
    ./main
    ```
2.  **Android Emulator:**
    Gunakan `10.0.2.2:9000` sebagai pengganti `localhost:9000` jika backend berjalan di mesin host yang sama.

### 🖼️ Alur Upload Gambar (Product)

Untuk fitur "Tambah Produk" atau "Edit Produk" di Mobile App:

1.  **Format Request:**
    Gunakan `Multipart/Form-Data` request.
2.  **Field Structure:**
    - Key `data`: JSON string dari object product. Pastikan **bukan** mengirim raw JSON di body, tapi sebagai form field string.
    - Key `image`: File binary gambar.
3.  **Client-Side Processing (Optional):**
    Backend sudah otomatis melakukan resize (max 1024px) dan kompresi (JPEG 85%). Mobile app **TIDAK PERLU** melakukan kompresi berat sebelum upload, cukup kirim file asli atau resize ringan jika file > 5MB.

4.  **Displaying Images:**
    Field `image_url` pada response JSON produk akan berisi full URL (contoh: `http://192.168.1.x:9000/warung-assets/products/uuid.jpg`).
    Mobile app cukup me-load URL tersebut ke component Image (React Native/Flutter/Android). Tidak perlu menambahkan Authorization header untuk me-load gambar (Public Read access).

---

## Role Access Summary

| Module           | Admin | Cashier | Inventory |
| ---------------- | ----- | ------- | --------- |
| Products (Read)  | ✅    | ✅      | ✅        |
| Products (Write) | ✅    | ❌      | ❌        |
| Customers        | ✅    | ✅      | ❌        |
| Transactions     | ✅    | ✅      | ❌        |
| Kasbon           | ✅    | ✅      | ❌        |
| Inventory        | ✅    | ❌      | ✅        |
| Stock Opname     | ✅    | ❌      | ✅        |
| Cash Flow        | ✅    | ✅      | ❌        |
| POS Features     | ✅    | ✅      | ❌        |
| Reports          | ✅    | ❌      | ❌        |
| Payments         | ✅    | ✅      | ❌        |
| Consignment      | ✅    | ❌      | ❌        |
| Refillables      | ✅    | ❌      | ✅        |

---

**Document Version:** 1.0
**Last Updated:** 2026-01-31
**Total Endpoints:** 71

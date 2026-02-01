# WarungOS Mobile Development Guide

This guide provides a comprehensive reference for developing the WarungOS mobile application. It details the **EXACT** valid request bodies and **COMPLETE** response structures for **ALL 71 ENDPOINTS**.

## 🔐 Authentication & Security

- **Base URL**: `https://api.warung-os.com` (or local `http://localhost:8080`)
- **Headers**:
  - `Authorization: Bearer <token>`
  - `Content-Type: application/json`

---

## 📡 API Reference

### 1. Authentication

### 2. Products

### 3. Categories

### 4. Customers

### 5. Transactions

#### User Login

**POST** `/auth/login`

```json
// Request Body
{
  "email": "admin@warung.com",
  "password": "password123"
}

// Response Body (200 OK)
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Super Admin",
      "email": "admin@warung.com",
      "role": "admin", // admin, cashier, inventory
      "is_active": true,
      "last_login_at": "2026-02-01T10:00:00Z",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-02-01T10:00:00Z"
    }
  }
}
```

#### Register User

**POST** `/auth/register`

```json
// Request Body
{
  "name": "Budi Cashier",
  "email": "budi@warung.com",
  "password": "password123",
  "role": "cashier" // admin, cashier, inventory
}

// Response Body (201 Created)
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174001",
    "name": "Budi Cashier",
    "email": "budi@warung.com",
    "role": "cashier",
    "is_active": true,
    "created_at": "2026-02-01T10:05:00Z",
    "updated_at": "2026-02-01T10:05:00Z"
  }
}
```

#### Refresh Token

**POST** `/auth/refresh`

```json
// Request Body
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

// Response Body (200 OK)
{
  "success": true,
  "data": {
    "access_token": "new_access_token...",
    "refresh_token": "new_refresh_token..."
  }
}
```

---

### 2. Products

#### List Products

**GET** `/api/v1/products`
Query Params: `page=1`, `per_page=20`, `search=name_or_barcode`, `category_id=uuid`, `low_stock_only=true`, `sort_by=name`, `sort_order=asc`

```json
// Response Body (200 OK)
{
  "success": true,
  "message": "Products retrieved",
  "data": [
    {
      "id": "product-uuid",
      "barcode": "89988123",
      "sku": "SKU-001",
      "name": "Indomie Goreng",
      "description": "Mie Instan Goreng",
      "category_id": "category-uuid",
      "unit": "bungkus",
      "base_price": 3500,
      "cost_price": 3100,
      "is_stock_active": true,
      "current_stock": 45,
      "min_stock_alert": 10,
      "max_stock": 100,
      "image_url": "https://storage...",
      "is_refillable": false,
      "is_active": true,
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "category": {
        "id": "category-uuid",
        "name": "Makanan Instan"
      },
      "pricing_tiers": [
        {
          "id": "tier-uuid",
          "product_id": "product-uuid",
          "name": "Grosir",
          "min_quantity": 40,
          "price": 3300,
          "is_active": true
        }
      ]
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### Create Product

**POST** `/api/v1/products`

```json
// Request Body
{
  "barcode": "899123456",
  "sku": "COFFEE-01",
  "name": "Kopi Kapal Api",
  "description": "Kopi Hitam",
  "category_id": "uuid",
  "unit": "sachet",
  "base_price": 1500,
  "cost_price": 1200,
  "is_stock_active": true,
  "current_stock": 100, // Optional initial stock
  "min_stock_alert": 20,
  "max_stock": 200,
  "image_url": "http://...",
  "pricing_tiers": [
    {
      "name": "Renceng",
      "min_quantity": 10,
      "price": 1400
    }
  ]
}
```

#### Search by Barcode

**GET** `/api/v1/products/search?barcode=899123`

```json
// Response Body (200 OK)
{
  "success": true,
  "data": {
    "id": "uuid",
    "barcode": "899123",
    "name": "Product Name"
    // ... complete product object
  }
}
```

#### Get Low Stock

**GET** `/api/v1/products/low-stock`

```json
// Response Body (200 OK)
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Product Name",
      "current_stock": 2,
      "min_stock_alert": 5
      // ... complete product object
    }
  ]
}
```

#### Get Product Detail

**GET** `/api/v1/products/{id}`

```json
// Response Body (200 OK)
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Product Name"
    // ... complete product object
  }
}
```

#### Update Product

**PUT** `/api/v1/products/{id}`

```json
// Request Body (All fields optional)
{
  "name": "Updated Name",
  "base_price": 2000,
  "is_active": true
}
```

#### Delete Product

**DELETE** `/api/v1/products/{id}`

```json
{ "success": true, "message": "Product deleted successfully" }
```

#### Add Pricing Tier

**POST** `/api/v1/products/{id}/pricing-tiers`

```json
// Request Body
{
  "name": "Special Promo",
  "min_quantity": 50,
  "price": 4000
}
```

#### Update Pricing Tier

**PUT** `/api/v1/products/{id}/pricing-tiers/{tierId}`

```json
// Request Body
{
  "price": 4200
}
```

#### Delete Pricing Tier

**DELETE** `/api/v1/products/{id}/pricing-tiers/{tierId}`

```json
{ "success": true, "message": "Pricing tier deleted" }
```

---

### 3. Categories (Product Categories)

#### List Categories

**GET** `/api/v1/categories`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Drinks",
      "description": "Minuman Segar",
      "is_active": true,
      "product_count": 5 // Number of active products in this category
    }
  ]
}
```

#### Create Category

**POST** `/api/v1/categories`

```json
// Request Body
{
  "name": "Snacks",
  "description": "Makanan Ringan"
}
```

#### Get Category Detail

**GET** `/api/v1/categories/{id}`

```json
// Response Body
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Drinks",
    "description": "Minuman Segar",
    "is_active": true
  }
}
```

#### Update Category

**PUT** `/api/v1/categories/{id}`

```json
// Request Body
{
  "name": "Beverages",
  "description": "Aneka Minuman"
}
```

#### Delete Category

**DELETE** `/api/v1/categories/{id}`

```json
// Response Body
{ "success": true, "message": "Category deleted successfully" }

// Note: Will fail with 400 Bad Request if category has active products.
```

---

### 4. Customers

#### List Customers

**GET** `/api/v1/customers`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Pak Budi",
      "phone": "08123456789",
      "address": "Jl. Mawar No 1",
      "notes": "Pelanggan Setia",
      "credit_limit": 500000,
      "current_debt": 0,
      "is_active": true,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ]
}
```

#### Create Customer

**POST** `/api/v1/customers`

```json
// Request Body
{
  "name": "Bu Siti",
  "phone": "0856...",
  "address": "Jl. Melati",
  "credit_limit": 1000000,
  "notes": "Warung Makan"
}
```

#### Customers with Debt

**GET** `/api/v1/customers/with-debt`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Pak Budi",
      "current_debt": 50000
      // ... complete customer object
    }
  ]
}
```

#### Get Customer Detail

**GET** `/api/v1/customers/{id}`

```json
// Response Body
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Pak Budi"
    // ... complete customer object
  }
}
```

#### Update Customer

**PUT** `/api/v1/customers/{id}`

```json
// Request Body
{
  "phone": "0812999999",
  "credit_limit": 600000
}
```

#### Delete Customer

**DELETE** `/api/v1/customers/{id}`

```json
{ "success": true, "message": "Customer deleted" }
```

#### Customer Kasbon History

**GET** `/api/v1/kasbon/customers/{id}`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "record-uuid",
      "customer_id": "customer-uuid",
      "transaction_id": "transaction-uuid", // Nullable if direct payment
      "type": "debt", // 'debt' or 'payment'
      "amount": 25000,
      "balance_before": 0,
      "balance_after": 25000,
      "notes": "Hutang Rokok",
      "created_by": "Cashier Name",
      "created_at": "timestamp"
    }
  ]
}
```

#### Customer Kasbon Summary

**GET** `/api/v1/kasbon/customers/{id}/summary`

```json
// Response Body
{
  "success": true,
  "data": {
    "customer_id": "uuid",
    "customer_name": "Pak Budi",
    "total_debt": 500000,
    "total_payment": 200000,
    "current_balance": 300000,
    "credit_limit": 1000000,
    "remaining_credit": 700000,
    "last_transaction_at": "timestamp"
  }
}
```

#### Download Billing PDF

**GET** `/api/v1/kasbon/customers/{id}/billing/pdf`

- **Response**: Binary stream (application/pdf)

#### Pay Debt

**POST** `/api/v1/kasbon/customers/{id}/payments`

```json
// Request Body
{
  "amount": 100000,
  "notes": "Transfer BCA",
  "created_by": "Admin" // Optional
}
```

---

### 4. Transactions

#### List Transactions

**GET** `/api/v1/transactions`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "invoice_number": "INV/2026/001",
      "customer_id": "uuid",
      "customer": { "name": "Pak Budi" },
      "subtotal": 50000,
      "discount_amount": 0,
      "tax_amount": 0,
      "total_amount": 50000,
      "payment_method": "cash",
      "amount_paid": 50000,
      "status": "completed",
      "created_at": "timestamp"
    }
  ],
  "meta": { "total": 100 }
}
```

#### Calculate Cart (Preview)

**POST** `/api/v1/transactions/calculate`

```json
// Request Body
{
  "items": [
    { "product_id": "uuid-1", "quantity": 10 },
    { "product_id": "uuid-2", "quantity": 1 }
  ]
}

// Response Body
{
  "success": true,
  "data": {
    "subtotal": 150000,
    "items": [
      {
        "product_id": "uuid-1",
        "product_name": "Item A",
        "quantity": 10,
        "unit": "pcs",
        "unit_price": 14000, // Price after wholesale tier
        "tier_name": "Grosir",
        "subtotal": 140000,
        "is_available": true,
        "available_qty": 50
      },
      {
        "product_id": "uuid-2",
        "product_name": "Item B",
        "quantity": 1,
        "unit": "pcs",
        "unit_price": 10000,
        "tier_name": "Harga Dasar",
        "subtotal": 10000,
        "is_available": true,
        "available_qty": 5
      }
    ]
  }
}
```

#### Create Transaction

**POST** `/api/v1/transactions`

```json
// Request Body
{
  "customer_id": "uuid-optional",
  "payment_method": "cash", // cash, kasbon, transfer, qris
  "amount_paid": 60000,
  "items": [
    {
      "product_id": "uuid",
      "quantity": 2,
      "discount_amount": 0,
      "notes": "Optional Item Note"
    }
  ],
  "discount_amount": 0,
  "tax_amount": 0,
  "notes": "Transaction Note"
}

// Response Body
{
  "success": true,
  "message": "Transaction created",
  "data": {
    "id": "uuid",
    "invoice_number": "INV/...",
    "total_amount": 50000,
    "change_amount": 10000,
    "status": "completed"
  }
}
```

#### Get Transaction Detail

**GET** `/api/v1/transactions/{id}`

```json
// Response Body
{
  "success": true,
  "data": {
    "id": "uuid",
    "invoice_number": "INV/...",
    "items": [
      {
        "id": "item-uuid",
        "product_name": "Product A",
        "quantity": 2,
        "unit_price": 25000,
        "total_amount": 50000
        // ... complete item object
      }
    ]
    // ... complete transaction object
  }
}
```

#### Cancel Transaction

**POST** `/api/v1/transactions/{id}/cancel`

```json
// Request Body
{
  "reason": "Salah input barang"
}

// Response Body
{ "success": true, "message": "Transaction cancelled" }
```

---

### 5. Inventory

#### Restock (Purchase)

**POST** `/api/v1/inventory/restock`

```json
// Request Body
{
  "product_id": "uuid",
  "quantity": 24,
  "cost_per_unit": 3200,
  "supplier_id": "uuid-opt", // not currently used but good for future
  "notes": "Restock from Agent A"
}
```

#### Manual Adjustment

**POST** `/api/v1/inventory/adjust`

```json
// Request Body
{
  "product_id": "uuid",
  "quantity": -2, // Negative to decrease
  "reason": "Expired / Damaged"
}
```

#### Low Stock Report

**GET** `/api/v1/inventory/low-stock`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "product": { "name": "Gula", "current_stock": 2 },
      "deficit_amount": 8
    }
  ]
}
```

#### Inventory Report

**GET** `/api/v1/inventory/report`

```json
// Response Body
{
  "success": true,
  "data": {
    "total_products": 150,
    "total_stock_value": 25000000, // Total HPP Asset
    "low_stock_count": 5,
    "out_of_stock_count": 2
  }
}
```

#### Restock PDF

**GET** `/api/v1/inventory/restock-list/pdf`

- **Response**: Binary PDF

#### Product Movements

**GET** `/api/v1/inventory/{productId}/movements`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "type": "sale", // sale, purchase, adjustment, return
      "quantity": -1,
      "stock_before": 10,
      "stock_after": 9,
      "reference_id": "transaction-uuid",
      "created_at": "timestamp"
    }
  ]
}
```

---

### 6. Reports

#### Dashboard

**GET** `/api/v1/reports/dashboard`

```json
// Response Body
{
  "success": true,
  "data": {
    "total_sales_today": 1500000,
    "transaction_count": 45,
    "gross_profit": 350000,
    "top_products": [{ "name": "Rokok", "qty": 50 }]
  }
}
```

#### Daily Report

**GET** `/api/v1/reports/daily?date=2026-02-01`

```json
// Response Body
{
  "success": true,
  "data": {
    "date": "2026-02-01",
    "summary": {
      "total_sales": 1500000,
      "total_transactions": 45,
      "cash_sales": 1000000,
      "qris_sales": 500000
    },
    "transactions": [ ...list of transactions... ]
  }
}
```

#### Kasbon Report

**GET** `/api/v1/reports/kasbon`

```json
// Response Body
{
  "success": true,
  "data": {
    "total_outstanding": 5000000,
    "total_customers": 10,
    "customers_with_debt": 5,
    "summaries": [ ...list of customer summaries... ]
  }
}
```

#### Inventory Value Report

**GET** `/api/v1/reports/inventory`

```json
// Response Body
{
  "success": true,
  "data": {
    "total_stock_value": 25000000,
    "total_items": 1500
  }
}
```

---

### 7. Payments (Midtrans)

#### Generate Snap Token

**POST** `/api/v1/payments/snap`

```json
// Request Body
{
  "transaction_id": "uuid",
  "gross_amount": 50000, // Should match transaction total
  "customer_name": "Budi", // Optional
  "customer_email": "budi@gmail.com", // Optional
  "item_details": [ // Optional
    { "id": "1", "price": 50000, "quantity": 1, "name": "Groceries" }
  ]
}

// Response Body
{
  "success": true,
  "data": {
    "token": "snap_token_xyz...",
    "redirect_url": "https://app.sandbox.midtrans.com/...",
    "order_id": "TRX-uuid-timestamp"
  }
}
```

#### Payment Notification (Webhook)

**POST** `/api/v1/payments/notification`

```json
// Request Body (From Midtrans)
{
  "transaction_status": "settlement",
  "order_id": "TRX-...",
  "gross_amount": "50000.00",
  "signature_key": "..."
  // ... other midtrans fields
}
```

#### Manual Verify

**POST** `/api/v1/payments/{id}/manual-verify`

```json
// Request Body
{
  "notes": "Verified via BCA Mobile"
}
```

#### Get Payment Info

**GET** `/api/v1/payments/transaction/{id}`

```json
// Response Body
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "settlement",
    "payment_type": "bank_transfer",
    "midtrans_response": { ... }
  }
}
```

---

### 8. Stock Opname

#### Start Session

**POST** `/api/v1/stock-opname/sessions`

```json
// Request Body
{
  "notes": "Opname Awal Bulan Februari",
  "created_by": "Inventory Staff"
}

// Response Body
{
  "success": true,
  "data": {
    "id": "uuid",
    "session_code": "OP/2026/02/001",
    "status": "in_progress",
    "created_at": "timestamp"
  }
}
```

#### List Sessions

**GET** `/api/v1/stock-opname/sessions`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "session_code": "OP/...",
      "status": "completed"
    }
  ]
}
```

#### Get Session Detail

**GET** `/api/v1/stock-opname/sessions/{id}`

```json
// Response Body
{
  "success": true,
  "data": {
    "id": "uuid",
    "items": [] // Empty if new
  }
}
```

#### Record Item Count

**POST** `/api/v1/stock-opname/sessions/{id}/items`

```json
// Request Body
{
  "product_id": "uuid",
  "physical_stock": 48,
  "notes": "Rusak 2",
  "counted_by": "Staff"
}
```

#### Finalize Session

**POST** `/api/v1/stock-opname/sessions/{id}/finalize`

```json
// Request Body
{
  "completed_by": "Supervisor",
  "apply_adjustments": true // If true, updates system stock automatically
}
```

#### Get Variance Report

**GET** `/api/v1/stock-opname/sessions/{id}/variance`

```json
// Response Body
{
  "success": true,
  "data": {
    "session_id": "uuid",
    "total_variance": -2,
    "total_loss_value": 5000,
    "net_value": -5000,
    "items": [
      {
        "product_id": "uuid",
        "product_name": "Gula",
        "system_stock": 50,
        "physical_stock": 48,
        "variance": -2,
        "variance_value": -5000
      }
    ]
  }
}
```

#### Cancel Session

**POST** `/api/v1/stock-opname/sessions/{id}/cancel`

```json
// Request Body
{ "notes": "Mistake" }
```

#### Shopping List

**GET** `/api/v1/stock-opname/shopping-list`

```json
// Response Body
{
  "success": true,
  "data": {
    "total_estimated_cost": 500000,
    "items": [
      {
        "product_id": "uuid",
        "product_name": "Kopi",
        "current_stock": 5,
        "min_stock": 20,
        "suggested_qty": 15
      }
    ]
  }
}
```

#### Near Expiry

**GET** `/api/v1/stock-opname/near-expiry`

```json
// Response Body
{
  "success": true,
  "data": {
    "days_ahead": 30,
    "items": [
      {
        "product_name": "Milk",
        "expiry_date": "2026-03-01",
        "days_until_expiry": 28
      }
    ]
  }
}
```

---

### 9. Cash Flow

#### Open Drawer

**POST** `/api/v1/cashflow/drawer/open`

```json
// Request Body
{
  "opening_balance": 150000,
  "opened_by": "Cashier Name",
  "notes": "Modal koin included"
}
```

#### Close Drawer

**POST** `/api/v1/cashflow/drawer/close`

```json
// Request Body
{
  "closing_balance": 550000,
  "closed_by": "Cashier Name",
  "notes": "Done"
}
// Response Body
{
  "success": true,
  "data": {
    "difference": 0,
    "status": "closed",
    "total_income": 400000
  }
}
```

#### Current Drawer Status

**GET** `/api/v1/cashflow/drawer/current`

```json
// Response Body
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "open",
    "opening_balance": 150000,
    "opened_at": "timestamp"
  }
}
```

#### Get Categories

**GET** `/api/v1/cashflow/categories`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Operasional",
      "type": "expense" // income, expense
    }
  ]
}
```

#### Record Cash Flow

**POST** `/api/v1/cashflow`

```json
// Request Body
{
  "type": "expense", // income, expense
  "amount": 15000,
  "description": "Beli Es Batu",
  "category_id": "uuid",
  "created_by": "Cashier Name"
}
```

#### List Cash Flows

**GET** `/api/v1/cashflow`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "type": "expense",
      "amount": 15000,
      "description": "Beli Es Batu"
    }
  ]
}
```

---

### 10. POS Features

#### Hold Cart

**POST** `/api/v1/pos/held-carts`

```json
// Request Body
{
  "customer_id": "uuid-optional",
  "held_by": "Cashier Name",
  "notes": "Pelanggan ke ATM",
  "items": [{ "product_id": "uuid", "quantity": 1, "notes": "Pedas" }]
}
```

#### List Held Carts

**GET** `/api/v1/pos/held-carts`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "hold_code": "XC92",
      "customer_name": "Pak Budi",
      "subtotal": 50000,
      "held_at": "timestamp"
    }
  ]
}
```

#### Get Held Cart Detail

**GET** `/api/v1/pos/held-carts/{id}`

```json
// Response Body
{
  "success": true,
  "data": {
    "id": "uuid",
    "items": [ ... ]
  }
}
```

#### Resume Cart

**POST** `/api/v1/pos/held-carts/{id}/resume`

```json
// Response Body
{
  "success": true,
  "message": "Cart resumed",
  "data": {
    // Returns full cart details to populate POS
    "items": [ ... ]
  }
}
```

#### Discard Held Cart

**POST** `/api/v1/pos/held-carts/{id}/discard`

```json
{ "success": true, "message": "Held cart discarded" }
```

#### Create Refund

**POST** `/api/v1/pos/refunds`

```json
// Request Body
{
  "transaction_id": "uuid",
  "refund_method": "cash",
  "reason": "Product Defect",
  "requested_by": "Cashier Name",
  "items": [
    {
      "transaction_item_id": "uuid",
      "quantity": 1,
      "restock": true, // Add back to inventory?
      "reason": "Broken seal"
    }
  ]
}
```

---

### 11. Consignment

#### Create Consignor

**POST** `/api/v1/consignors`

```json
// Request Body
{
  "name": "Ibu Yanti Donat",
  "phone": "0812345678",
  "address": "Pasar Lama",
  "bank_account": "BCA 123456",
  "bank_name": "BCA"
}
```

#### List Consignors

**GET** `/api/v1/consignors`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Ibu Yanti Donat",
      "products": [ ... ]
    }
  ]
}
```

#### Update Consignor

**PUT** `/api/v1/consignors/{id}`

```json
// Request Body
{
  "phone": "0899999",
  "is_active": true
}
```

---

### 12. Refillables

#### Get Containers

**GET** `/api/v1/refillables`

```json
// Response Body
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "container_type": "gas_3kg",
      "empty_count": 10,
      "full_count": 25,
      "product": { "name": "Gas 3kg Isi" }
    },
    {
      "id": "uuid",
      "container_type": "aqua_galon",
      "empty_count": 5,
      "full_count": 10,
      "product": { "name": "Aqua Galon" }
    }
  ]
}
```

#### Adjust Container Stock

**POST** `/api/v1/refillables/adjust`

```json
// Request Body
{
  "container_id": "uuid",
  "type": "adjustment", // adjustment, return_empty, purchase_full
  "empty_change": -1, // Change in empty count
  "full_change": 1 // Change in full count
}
```

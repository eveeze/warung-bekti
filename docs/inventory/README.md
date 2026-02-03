# Inventory Module

Base URL: `/api/v1`

## Business Context

Inventory tracks the **movement history** of products (not just current levels).

- **Audit Trail**: Who changed the stock? When? Why?
- **Restocking**: Receiving goods from suppliers (Cost Price tracking).

## Frontend Implementation Guide

### 1. Stock History

> [!TIP]
> **Optimistic UI**: Stock adjustments (Restock/Correction) return valid `ETag`.
> Use standard `success: true` wrapper for all responses.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

- Infinite scroll list of movements (`IN`, `OUT`, `ADJUST`).
- Color code: `IN` (Green), `OUT` (Red), `ADJUST` (Orange).

### 2. Low Stock Dashboard

- Highlight items where `current_stock <= min_stock_alert`.
- "Quick Restock" button to open restock modal.

## Endpoints

### 1. Restock Product

Add stock to a single product.

- **URL**: `/inventory/restock`
- **Method**: `POST`
- **Auth Required**: Yes (Inventory)

#### Request Body

```json
{
  "product_id": "uuid",
  "quantity": 100,
  "cost_per_unit": 12000,
  "notes": "Restock from Supplier A"
}
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Stock added successfully",
  "data": { "current_stock": 150 }
}
```

### 2. Adjust Stock (Opname/Correction)

Manually adjust stock (can be negative).

- **URL**: `/inventory/adjust`
- **Method**: `POST`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "product_id": "uuid",
  "quantity": -5, // Negative to reduce
  "reason": "Damaged goods / Expired"
}
```

### 3. Get Low Stock

List products running low on stock.

- **URL**: `/inventory/low-stock`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

### 4. Get Inventory Summary

Get overview of stock value and health.

- **URL**: `/inventory/report`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Inventory report",
  "data": {
    "total_products": 150,
    "total_stock_value": 15000000,
    "low_stock_count": 5,
    "out_of_stock_count": 2,
    "low_stock_products": [...]
  }
}
```

### 5. Download Restock PDF

Generate PDF list of items needing restock.

- **URL**: `/inventory/restock-list/pdf`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

### 6. Get Movements

View stock history for a product.

- **URL**: `/inventory/{productId}/movements`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

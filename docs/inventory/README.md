# Inventory Module

Base URL: `/api/v1`

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
  "total_products": 150,
  "total_stock_value": 15000000,
  "low_stock_count": 5,
  "out_of_stock_count": 2,
  "low_stock_products": [...]
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

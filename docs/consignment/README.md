# Consignment Module

Base URL: `/api/v1`

## Business Context

Manages "Titip Jual" (goods owned by others).

- **Consignor**: The person who owns the goods.
- **Settlement**: Paying the consignor for sold items (Revenue - Commission).

## Frontend Implementation Guide

### 1. Consignor Dashboard & Settlements

> [!TIP]
> **Optimistic UI Support**:
> Use `ETag` to cache consignor lists and settlement history.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

- **Owed Amount**: Calculation: `Total Sales of Consignment Items` - `Commission`.a product, show a dropdown **"Supplier / Consignor"**.
- **Context**: Linking a product to a consignor allows the system to track who owns the stock and who needs to be paid when it sells.

### 3. Inventory Filters

- **Stock List**: Add filter by "Consignor" to see stock belonging to specific suppliers.
- **Owed Amount**: Calculation: `Total Sales of Consignment Items` - `Commission`.
- **Settlement Button**: "Pay Consignor" -> Records an Expense in Cash Flow.

## Endpoints

### 1. Create Consignor

Register a new consignor (supplier).

- **URL**: `/consignors`
- **Method**: `POST`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "name": "Supplier B",
  "phone": "08xxxx",
  "bank_account": "123456",
  "bank_name": "BCA"
}
```

#### Response (201 Created)

```json
{
  "success": true,
  "message": "Consignor created",
  "data": { "id": "uuid", "name": "Supplier B", ... }
}
```

### 2. List Consignors

List all consignors.

- **URL**: `/consignors`
- **Method**: `GET`
- **Auth Required**: Yes (Admin only)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Consignors retrieved",
  "data": [ ... ]
}
```

### 3. Update Consignor

Update consignor details.

- **URL**: `/consignors/{id}`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "name": "Updated Name",
  "is_active": true
}
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Consignor updated",
  "data": { ... }
}
```

### 4. Delete Consignor

Soft delete a consignor.

- **URL**: `/consignors/{id}`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin only)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Consignor deleted successfully",
  "data": null
}
```

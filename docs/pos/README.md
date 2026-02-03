# POS Features Module

Base URL: `/api/v1`

## Business Context

The POS (Point of Sale) module is the core interface for cashiers. It handles:

- **Held Carts**: Allows pausing a transaction (e.g., customer forgot wallet or wants to add more items) to serve the next customer, preventing line blockage.
- **Refunds**: Managing returns and defective products with inventory restock options.

## Frontend Implementation Guide

### 1. Cart Management (State)

> [!TIP]
> **Optimistic UI Support**: Start the "Hold Cart" animation immediately.
> All mutation endpoints return an `ETag` header for cache sync.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

- Use a global store (Zustand/Redux) for the active cart.
- **Offline Support**: Persist the cart to `AsyncStorage` (Mobile) or `localStorage` (Web) to prevent data loss on crash/refresh.

### 2. Implementing "Hold Cart"

1.  **Trigger**: "Hold" button in the cart UI.
2.  **Logic**:
    - Validate cart is not empty.
    - Call `POST /api/v1/pos/held-carts`.
    - On success: Clear local cart state and show toast "Cart Held".
    - Refresh the "Held Carts" sidebar list.

### 3. Resuming a Cart

1.  **Trigger**: Clicking an item in the "Held Carts" sidebar.
2.  **Logic**:
    - Check if current active cart is empty. If not, prompt user to Hold or Clear it first.
    - Call `POST /api/v1/pos/held-carts/{id}/resume`.
    - On success: Populate the global cart state with the items from the response.
    - Remove the item from the sidebar list.

### 4. Refunds

- Normally accessed via **Transaction History**.
- Use a modal to select items to refund (partial or full).
- **Restock Checkbox**: Important! Ask user "Return items to inventory?". If true, send `restock: true`.

## Endpoints

### 1. Hold Cart

Save a shopping cart for later.

- **URL**: `/pos/held-carts`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "customer_id": "uuid", // Optional
  "held_by": "Staff Name",
  "items": [{ "product_id": "uuid", "quantity": 1 }]
}
```

#### Response (201 Created)

```json
{
  "success": true,
  "message": "Cart held successfully",
  "data": { "id": "uuid", ... }
}
```

### 2. List Held Carts

View all active held carts.

- **URL**: `/pos/held-carts`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

### 3. Resume Cart

Retrieve and reactivate a held cart.

- **URL**: `/pos/held-carts/{id}/resume`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

### 4. Discard Cart

Delete a held cart without processing.

- **URL**: `/pos/held-carts/{id}/discard`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

### 5. Create Refund

Refund a completed transaction.

- **URL**: `/pos/refunds`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "transaction_id": "uuid",
  "refund_method": "cash",
  "reason": "Defective Product",
  "requested_by": "Staff",
  "items": [
    {
      "transaction_item_id": "uuid",
      "quantity": 1,
      "restock": true // Return to inventory?
    }
  ]
}
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Refund processed successfully",
  "data": { ... }
}
```

# POS Features Module

Base URL: `/api/v1`

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

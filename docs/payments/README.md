# Payments Module

Base URL: `/api/v1`

## Endpoints

### 1. Generate Snap Token (Midtrans)

Initiate a payment and get a Snap token.

- **URL**: `/payments/snap`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "transaction_id": "uuid",
  "gross_amount": 50000,
  "customer_name": "Budi", // Optional
  "customer_email": "budi@mail.com", // Optional
  "customer_phone": "08123...", // Optional
  "item_details": [
    { "id": "uuid", "price": 50000, "quantity": 1, "name": "Product A" }
  ]
}
```

#### Response (200 OK)

```json
{
  "token": "snap-token-123",
  "redirect_url": "https://app.sandbox.midtrans.com/...",
  "order_id": "ORDER-123"
}
```

### 2. Handle Midtrans Notification

Webhook endpoint for payment status updates.

- **URL**: `/payments/notification`
- **Method**: `POST`
- **Auth Required**: No (Public, Verified via Signature)

#### Request Body

(Standard Midtrans Notification JSON)

### 3. Manual Verify Payment

Manually verify a payment if automated callback fails.

- **URL**: `/payments/{id}/manual-verify`
- **Method**: `POST`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "notes": "Verified via BCA Dashboard"
}
```

### 4. Get Payment By Transaction

Check payment status for a transaction.

- **URL**: `/payments/transaction/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

# Payments Module

Base URL: `/api/v1`

## Business Context

Handles all money collection methods.

- **Midtrans**: Automated payments (QRIS, VA).
- **Manual**: Cash (Physical), Transfer (Requires manual verification check).

## Frontend Implementation Guide

### 1. Midtrans Snap

> [!TIP]
> **Optimistic UI**: Polling status can be done in background.
> Use `ETag` on polling endpoint to reduce bandwidth (304 Not Modified).
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

1.  Frontend sends checkout data to `POST /payments/snap`.
2.  Backend returns `token` and `redirect_url`.
3.  **Web**: Redirect user to `redirect_url` or use Snap.js popup.
4.  **Mobile**: Use WebView to load `redirect_url`.

### 2. Payment Status

- **Polling**: Frontend polls `GET /payments/transaction/{id}` every 3s until status becomes `paid`.
- **UI**: Show "Waiting for Payment..." spinner.

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

````json
```json
{
  "success": true,
  "message": "Snap token generated",
  "data": {
    "token": "snap-token-123",
    "redirect_url": "https://app.sandbox.midtrans.com/...",
    "order_id": "ORDER-123"
  }
}
````

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

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Payment verified",
  "data": { "status": "paid" }
}
```

### 4. Get Payment By Transaction

Check payment status for a transaction.

- **URL**: `/payments/transaction/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Payment status",
  "data": {
    "transaction_id": "uuid",
    "status": "paid", // pending, paid, expired
    "payment_type": "qris"
  }
}
```

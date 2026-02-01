# Transactions Module

Base URL: `/api/v1`

## Endpoints

### 1. List Transactions

Retrieve sales history.

- **URL**: `/transactions`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

#### Query Parameters

| Parameter        | Type     | Description                   |
| :--------------- | :------- | :---------------------------- |
| `page`           | `int`    | Page number                   |
| `per_page`       | `int`    | Items per page                |
| `customer_id`    | `uuid`   | Filter by customer            |
| `status`         | `string` | pending, completed, cancelled |
| `payment_method` | `string` | cash, kasbon, transfer        |
| `date_from`      | `string` | ISO Date                      |
| `date_to`        | `string` | ISO Date                      |

### 2. Create Transaction (Checkout)

Process a sale.

- **URL**: `/transactions`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "customer_id": "uuid", // Optional
  "items": [
    {
      "product_id": "uuid",
      "quantity": 2,
      "discount_amount": 0, // Optional per item discount
      "notes": "..."
    }
  ],
  "discount_amount": 0, // Global discount
  "tax_amount": 0, // Optional
  "payment_method": "cash", // cash, kasbon, transfer, qris, mixed
  "amount_paid": 50000,
  "notes": "..."
}
```

#### Response (201 Created)

```json
{
  "id": "uuid",
  "invoice_number": "INV/2023/10/01/001",
  "total_amount": 45000,
  "change_amount": 5000,
  "status": "completed",
  ...
}
```

### 3. Calculate Cart

Preview totals before checkout (checks pricing tiers).

- **URL**: `/transactions/calculate`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "items": [{ "product_id": "uuid", "quantity": 10 }]
}
```

#### Response (200 OK)

```json
{
  "items": [
    {
      "product_id": "uuid",
      "unit_price": 10000,
      "tier_name": "Grosir",
      "subtotal": 100000
    }
  ],
  "subtotal": 100000
}
```

### 4. Get Transaction Details

Retrieve full details of a transaction.

- **URL**: `/transactions/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

### 5. Cancel Transaction

Void a transaction (restores stock if applicable).

- **URL**: `/transactions/{id}/cancel`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

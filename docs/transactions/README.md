# Transactions Module

Base URL: `/api/v1`

## Business Context

Transactions represent the sales activity in the Warung.

- **Workflow**: Cart -> Calculate (Apply Discounts/Wholesale) -> Payment -> Receipt.
- **Kasbon**: Supports "Pay Later" (Credit) which links to the Customer module.
- **Void/Cancel**: Reverses the sale and restores inventory.

## Frontend Implementation Guide

### 1. Checkout Flow

> [!TIP]
> **Optimistic UI**: Frontend can calculate totals locally for instant feedback.
> Use `POST /transactions/calculate` for the authoritative final price.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

1.  **Calculate**: Call `POST /transactions/calculate` whenever cart changes (debounce 300ms) to show accurate totals covering wholesale prices.
2.  **Payment Modal**: Select method (Cash, Transfer, Kasbon).
3.  **Submit**: Call `POST /transactions`.
4.  **Success**: Show "Change Due" (Kembalian) and "Print Receipt" button.

### 2. Receipt Printing

- Use a library like `react-native-thermal-receipt-printer` or standard Web Print API.
- Data comes from the `POST /transactions` response (including `invoice_number`, `items`, `total`, `change`).

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
| `search`         | `string` | Search Invoice or Customer    |
| `customer_id`    | `uuid`   | Filter by customer            |
| `status`         | `string` | pending, completed, cancelled |
| `payment_method` | `string` | cash, kasbon, transfer        |
| `date_from`      | `string` | ISO Date                      |
| `date_to`        | `string` | ISO Date                      |

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Transactions retrieved",
  "data": [
    {
      "id": "uuid",
      "invoice_number": "INV/...",
      "customer_id": "uuid",
      "customer": {
        "id": "uuid",
        "name": "Budi Santoso"
      },
      "total_amount": 50000,
      "status": "completed",
      "created_at": "2023-..."
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 100 }
}
```

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
  "success": true,
  "message": "Transaction created successfully",
  "data": {
    "id": "uuid",
    "invoice_number": "INV/2023/10/01/001",
    "total_amount": 45000,
    "change_amount": 5000,
    "status": "completed",
    ...
  }
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
  "success": true,
  "message": "Calculation result",
  "data": {
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
}
```

### 4. Get Transaction Details

Retrieve full details of a transaction.

- **URL**: `/transactions/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Transaction retrieved",
  "data": {
    "id": "uuid",
    "invoice_number": "INV/...",
    "customer": {
      "id": "uuid",
      "name": "Budi Santoso"
    },
    "items": [
      {
        "product_name": "Kopi",
        "quantity": 2,
        "unit_price": 5000,
        "total_amount": 10000
      }
    ],
    "subtotal": 10000,
    "discount_amount": 0,
    "tax_amount": 0,
    "total_amount": 10000,
    "payment_method": "cash",
    "amount_paid": 10000,
    "change_amount": 0
  }
}
```

### 5. Cancel Transaction

Void a transaction (restores stock if applicable).

- **URL**: `/transactions/{id}/cancel`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

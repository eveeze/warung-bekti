# Customers Module

Base URL: `/api/v1`

## Business Context

Manages customer relationships and, crucially, **Kasbon (Debt)**.

- **Kasbon**: "Buy now, pay later". System tracks credit limits and outstanding balances.
- **Loyalty**: Tracking purchase history for potential rewards (future).

## Frontend Implementation Guide

### 1. Customer List

> [!TIP]
> **Optimistic UI**: Update debt status immediately upon payment.
> Use `ETag` to validate customer lists.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

- **Debt Indicator**: Highlight customers with `current_debt > 0` (Red badge).
- **Search**: Optimize for searching by Name OR Phone Number.

### 2. Paying Debt (Kasbon Repayment)

- Accessible from **Customer Details** page.
- **Modal**: Input "Amount to Pay".
- **Logic**: Call `POST /kasbon/customers/{id}/payments`. Update UI to reflect lower debt immediately.

## Endpoints

### 1. List Customers

Retrieve a paginated list of customers.

- **URL**: `/customers`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

#### Query Parameters

| Parameter   | Type     | Description                |
| :---------- | :------- | :------------------------- |
| `page`      | `int`    | Page number                |
| `per_page`  | `int`    | Items per page             |
| `search`    | `string` | Search by name or phone    |
| `has_debt`  | `bool`   | Filter customers with debt |
| `is_active` | `bool`   | Filter by status           |

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Customers retrieved",
  "data": [
    {
      "id": "uuid",
      "name": "Customer Name",
      "phone": "08123456789",
      "address": "Alamat...",
      "credit_limit": 500000,
      "current_debt": 20000,
      "is_active": true
    }
  ]
}
```

### 2. Create Customer

Register a new customer.

- **URL**: `/customers`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "name": "New Customer",
  "phone": "081xxxx", // Optional
  "address": "Alamat", // Optional
  "notes": "Catatan", // Optional
  "credit_limit": 1000000 // Optional
}
```

#### Response (201 Created)

```json
{
  "success": true,
  "message": "Customer created",
  "data": { "id": "uuid", "name": "New Customer", ... }
}
```

### 3. Get Customers With Debt

Shortcut to get customers who owe money.

- **URL**: `/customers/with-debt`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

### 4. Get Customer Details

Retrieve specific customer details.

- **URL**: `/customers/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

### 5. Update Customer

Update customer information.

- **URL**: `/customers/{id}`
- **Method**: `PUT`
- **Auth Required**: Yes (Cashier)

### 6. Delete Customer

Soft delete a customer.

- **URL**: `/customers/{id}`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin only)

### 7. Kasbon History

View kasbon transaction history for a customer.

- **URL**: `/kasbon/customers/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

### 8. Kasbon Summary

Get summary of debt and payments.

- **URL**: `/kasbon/customers/{id}/summary`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

### 9. Download Billing PDF

Generate PDF statement for customer debt.

- **URL**: `/kasbon/customers/{id}/billing/pdf`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

### 10. Record Payment (Pay Debt)

Record a payment against a customer's debt.

- **URL**: `/kasbon/customers/{id}/payments`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

#### Request Body

```json
{
  "amount": 10000,
  "notes": "Partial payment"
}
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Payment recorded",
  "data": { "new_debt_balance": 10000 }
}
```

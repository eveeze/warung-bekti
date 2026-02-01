# Customers Module

Base URL: `/api/v1`

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

TODO: Check `PaymentHandler.RecordPayment` for body, likely `{ "amount": 10000, "notes": "..." }`

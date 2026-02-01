# Reports Module

Base URL: `/api/v1`

## Endpoints

### 1. Daily Report

Get sales summary for a specific date (default today).

- **URL**: `/reports/daily`
- **Method**: `GET`
- **Auth Required**: Yes (Admin only)

#### Query Parameters

| Parameter | Type     | Description |
| :-------- | :------- | :---------- |
| `date`    | `string` | YYYY-MM-DD  |

#### Response (200 OK)

```json
{
  "date": "2023-10-01",
  "total_sales": 500000,
  "total_transactions": 25,
  "estimated_profit": 150000
}
```

### 2. Kasbon Report

Overview of outstanding debts.

- **URL**: `/reports/kasbon`
- **Method**: `GET`
- **Auth Required**: Yes (Admin only)

### 3. Inventory Report

Value of current inventory.

- **URL**: `/reports/inventory`
- **Method**: `GET`
- **Auth Required**: Yes (Admin only)

### 4. Dashboard Summary

Combined stats for homepage dashboard.

- **URL**: `/reports/dashboard`
- **Method**: `GET`
- **Auth Required**: Yes (Admin only)

#### Response (200 OK)

```json
{
  "today": { "total_sales": ... },
  "total_outstanding_kasbon": 2000000,
  "low_stock_count": 5,
  "out_of_stock_count": 0
}
```

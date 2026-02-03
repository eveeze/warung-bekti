# Reports Module

Base URL: `/api/v1`

## Business Context

Provides insights into Warung health.

- **Key Metrics**: Total Sales, Profit Estimations, Kasbon Outstanding.
- **Decision Making**: Which products sell best? Who owes the most money?

## Frontend Implementation Guide

### 1. Dashboard Widgets

> [!TIP]
> **Performance**: Reports are heavy. Use `ETag` to cache results.
> Backend returns `304 Not Modified` if data hasn't changed.

- **Daily Sales**: Big number card.
- **Product Alerts**: "5 Items Low Stock" (Clickable -> goes to Inventory).
- **Kasbon**: Total outstanding debt card.

### 2. Charts

- Use `Recharts` (Web) or `Victory/react-native-chart-kit` (Mobile).
- **Sales Trend**: Line chart (Last 7 days/30 days).
- **Top Products**: Bar chart or Pie chart.

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
  "success": true,
  "message": "Daily report retrieved",
  "data": {
    "date": "2023-10-01",
    "total_sales": 500000,
    "total_transactions": 25,
    "estimated_profit": 150000
  }
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
  "success": true,
  "message": "Dashboard data retrieved",
  "data": {
    "today": { "total_sales": ... },
    "total_outstanding_kasbon": 2000000,
    "low_stock_count": 5,
    "out_of_stock_count": 0
  }
}
```

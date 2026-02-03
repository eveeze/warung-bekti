# Cash Flow Module

Base URL: `/api/v1`

## Business Context

Tracks money entering and leaving the physical drawer (Shift Management).

- **Drawer Sessions**: Ensuring cash in drawer matches system sales at the end of a shift.
- **Petty Cash**: Recording small expenses taken from the drawer (e.g., buying ice, gasoline).

## Frontend Implementation Guide

### 1. Drawer Workflow

> [!TIP]
> **Optimistic UI**: Record expenses locally primarily.
> Sync with backend using `ETag`.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

- **Open Session**: At start of day, prompt user to count cash. Input `opening_balance`.
- **Operating**: Throughout the day, record "Expense" for petty cash.
- **Close Session**: At end of day, user counts cash again (`closing_balance`). System calculates `difference` (Shortage/Surplus).

### 2. Expense Form

- Simple form: Category (Transport, Supplies), Amount, Description.

## Endpoints

### 1. Open Drawer Session

Start a cashier shift.

- **URL**: `/cashflow/drawer/open`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "opening_balance": 200000,
  "opened_by": "Cashier Name",
  "notes": "Start of Morning Shift"
}
```

#### Response (201 Created)

```json
{
  "success": true,
  "message": "Drawer opened",
  "data": { "id": "uuid", "opened_at": "..." }
}
```

### 2. Close Drawer Session

End a cashier shift.

- **URL**: `/cashflow/drawer/close`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "session_id": "uuid",
  "closing_balance": 1500000,
  "closed_by": "Cashier Name"
}
```

### 3. Get Current Session

Check if there is an active session.

- **URL**: `/cashflow/drawer/current`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

### 4. Get Categories

List income/expense categories.

- **URL**: `/cashflow/categories`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Categories retrieved",
  "data": [{ "id": "uuid", "name": "Transport", "type": "expense" }]
}
```

### 5. Record Cash Flow

Log a manual expense or income.

- **URL**: `/cashflow`
- **Method**: `POST`
- **Auth Required**: Yes (Cashier)

#### Request Body

```json
{
  "category_id": "uuid",
  "type": "expense", // income or expense
  "amount": 50000,
  "description": "Buy ice cubes",
  "created_by": "Staff"
}
```

#### Response (201 Created)

```json
{
  "success": true,
  "message": "Cash flow recorded",
  "data": { "id": "uuid", ... }
}
```

### 6. List Cash Flows

View history of cash movements.

- **URL**: `/cashflow`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

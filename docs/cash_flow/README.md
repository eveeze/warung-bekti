# Cash Flow Module

Base URL: `/api/v1`

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

### 6. List Cash Flows

View history of cash movements.

- **URL**: `/cashflow`
- **Method**: `GET`
- **Auth Required**: Yes (Cashier)

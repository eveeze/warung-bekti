# Stock Opname Module

Base URL: `/api/v1`

## Endpoints

### 1. List Sessions

List all opname sessions.

- **URL**: `/stock-opname/sessions`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

### 2. Start Session

Begin a new stock taking session.

- **URL**: `/stock-opname/sessions`
- **Method**: `POST`
- **Auth Required**: Yes (Inventory)

#### Request Body

```json
{
  "notes": "Monthly Audit",
  "created_by": "Staff Id"
}
```

### 3. Get Session Details

View details of a session.

- **URL**: `/stock-opname/sessions/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

### 4. Record Count

Submit physical count for a product.

- **URL**: `/stock-opname/sessions/{id}/items`
- **Method**: `POST`
- **Auth Required**: Yes (Inventory)

#### Request Body

```json
{
  "product_id": "uuid",
  "physical_stock": 45,
  "notes": "Found damaged items", // Optional
  "counted_by": "Staff Name"
}
```

### 5. Get Variance Report

View discrepancies in a session.

- **URL**: `/stock-opname/sessions/{id}/variance`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

#### Response (200 OK)

```json
{
  "total_variance": -5,
  "total_loss_value": 50000,
  "items": [
    {
      "product_name": "Item A",
      "system_stock": 50,
      "physical_stock": 45,
      "variance": -5,
      "variance_value": -50000
    }
  ]
}
```

### 6. Finalize Session

Close session and apply adjustments.

- **URL**: `/stock-opname/sessions/{id}/finalize`
- **Method**: `POST`
- **Auth Required**: Yes (Inventory)

#### Request Body

```json
{
  "apply_adjustments": true,
  "completed_by": "Supervisor Name"
}
```

### 7. Cancel Session

Abort an opname session.

- **URL**: `/stock-opname/sessions/{id}/cancel`
- **Method**: `POST`
- **Auth Required**: Yes (Inventory)

### 8. Get Shopping List

Auto-generate restock list based on min stock.

- **URL**: `/stock-opname/shopping-list`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

### 9. Get Near Expiry Report

List items nearing expiry (if batch tracking enabled).

- **URL**: `/stock-opname/near-expiry`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

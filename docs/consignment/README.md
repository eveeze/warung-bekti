# Consignment Module

Base URL: `/api/v1`

## Endpoints

### 1. Create Consignor

Register a new consignor (supplier).

- **URL**: `/consignors`
- **Method**: `POST`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "name": "Supplier B",
  "phone": "08xxxx",
  "bank_account": "123456",
  "bank_name": "BCA"
}
```

### 2. List Consignors

List all consignors.

- **URL**: `/consignors`
- **Method**: `GET`
- **Auth Required**: Yes (Admin only)

### 3. Update Consignor

Update consignor details.

- **URL**: `/consignors/{id}`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "name": "Updated Name",
  "is_active": true
}
```

# Refillables Module

Base URL: `/api/v1`

## Endpoints

### 1. Get Containers

View status of refillable containers (galon, gas).

- **URL**: `/refillables`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

#### Response (200 OK)

```json
{
  "data": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "container_type": "Gas 3kg",
      "empty_count": 10,
      "full_count": 40
    }
  ]
}
```

### 2. Adjust Stock

Manual adjustment for container counts.

- **URL**: `/refillables/adjust`
- **Method**: `POST`
- **Auth Required**: Yes (Inventory)

#### Request Body

```json
{
  "container_id": "uuid",
  "empty_change": 5, // Add 5 empty
  "full_change": -5, // Reduce 5 full
  "notes": "Found in warehouse"
}
```

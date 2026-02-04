# Refillables Module

Base URL: `/api/v1`

## Business Context

Special logic for "Swap" products (Gallon Water, LPG Gas).

- **Core Concept**: Customer buys the _contents_, but swapping the _container_ is essential.
- **Inventory**: Tracks "Full Containers" vs "Empty Containers".

## Frontend Implementation Guide

### 1. Swap Logic

> [!TIP]
> **Optimistic UI**: Track container counts locally.
> Use `ETag` to validate inventory state.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

- **Purchase**: Regular Transaction.
- **Stock Effect**:
  - `Full Container` stock -1.
  - `Empty Container` stock +1 (if swapped).
- **Deposit**: If customer has no empty container, sell "Container Deposit" (Non-refillable product).

## Endpoints

### 1. Get Containers

View status of refillable containers (galon, gas).

- **URL**: `/refillables`
- **Method**: `GET`
- **Auth Required**: Yes (Inventory)

#### Response (200 OK)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Containers retrieved",
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

````

### 2. Create Container

Initialize a new refillable type mapping.

- **URL**: `/refillables`
- **Method**: `POST`
- **Auth Required**: Yes (Inventory)

#### Request Body

```json
{
  "product_id": "uuid",
  "container_type": "Galon Aqua",
  "empty_count": 0,
  "full_count": 0,
  "notes": "Initial stock"
}
````

#### Response (201 Created)

```json
{
  "success": true,
  "message": "Container created",
  "data": { "id": "uuid", ... }
}
```

### 3. Adjust Stock

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

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Stock adjusted",
  "data": { ... }
}
```

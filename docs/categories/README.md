# Categories Module

Base URL: `/api/v1`

## Endpoints

### 1. List Categories

Retrieve all categories (flat list).

- **URL**: `/categories`
- **Method**: `GET`
- **Auth Required**: Yes (Protected)

#### Response (200 OK)

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Beverages",
      "description": "Drinks and liquids",
      "product_count": 15, // Count of active products
      "is_active": true
    }
  ]
}
```

### 2. Create Category

Create a new category (Admin only).

- **URL**: `/categories`
- **Method**: `POST`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "name": "Snacks",
  "description": "Chips and crackers", // Optional
  "parent_id": "uuid" // Optional (for subcategories)
}
```

#### Response (201 Created)

```json
{
  "id": "uuid",
  "name": "Snacks",
  ...
}
```

### 3. Get Category By ID

Retrieve a specific category.

- **URL**: `/categories/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Protected)

### 4. Update Category

Update an existing category.

- **URL**: `/categories/{id}`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "name": "Updated Name",
  "description": "Updated Description",
  "is_active": true
}
```

### 5. Delete Category

Soft delete a category (only if valid).

- **URL**: `/categories/{id}`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin only)

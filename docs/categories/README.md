# Categories Module

Base URL: `/api/v1`

## Business Context

Categories organize products for easier navigation in the POS and reporting.

- **Organization**: Grouping similar items (e.g., "Snacks", "Drinks").
- **Reporting**: Sales reports are often aggregated by category.

## Frontend Implementation Guide

### 1. POS Sidebar

> [!TIP]
> **Optimistic UI**: Pre-load subcategories or assume success when expanding trees.
> Use `ETag` headers to validate category lists.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md).

- Fetch categories to render the sidebar navigation.
- Filtering products by `category_id` when a user clicks a category.

### 2. Category Management

- **Deletion Logic**: Before deleting, check `product_count`. If > 0, warn the user "Category has X products. Please reassign them first." or allow force delete (setting product category to null).

## Endpoints

### 1. List Categories

Retrieve all categories (flat list).

- **URL**: `/categories`
- **Method**: `GET`
- **Auth Required**: Yes (Protected)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Categories retrieved",
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
  "success": true,
  "message": "Category created",
  "data": {
    "id": "uuid",
    "name": "Snacks",
    ...
  }
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

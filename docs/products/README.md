# Products Module

Base URL: `/api/v1`

## Business Context

Products are the items sold in the Warung.

- **Stock Management**: Tracks inventory levels (`current_stock`) and alerts (`min_stock_alert`).
- **Pricing Tiers**: Supports wholesale pricing (e.g., Buy 10 get lower price).
- **Barcodes**: Essential for fast checkout using scanners.

## Frontend Implementation Guide

> [!TIP]
> **Optimistic UI Support**: All mutation endpoints (POST/PUT/DELETE) return an `ETag` header.
> Use this `ETag` to update your local cache without refetching.
> See [Optimistic UI Guide](../OPTIMISTIC_UI.md) for implementation details.

### 1. Product List (Infinite Scroll)

- **API**: `GET /api/v1/products?page=1&per_page=20`
- **Logic**:
  - Use React Query / TanStack Query `useInfiniteQuery`.
  - Detect scroll to bottom -> fetch next page.
  - **Debounce Search**: Wait 500ms after user stops typing in search bar before calling API.

### 2. Product Form (Create/Update)

- **Validation**: Ensure `base_price` >= `cost_price` to prevent loss (warn user if not).
- **Image Upload**: See [Frontend Image Upload Guide](../../frontend_image_upload_guide.md).
- **Pricing Tiers**: Dynamic form array (add/remove tiers).

## Endpoints

### 1. List Products

Retrieve a paginated list of products with optional filtering.

- **URL**: `/products`
- **Method**: `GET`
- **Auth Required**: Yes (Protected)

#### Query Parameters

| Parameter         | Type     | Description                     | Default |
| :---------------- | :------- | :------------------------------ | :------ |
| `page`            | `int`    | Page number                     | 1       |
| `per_page`        | `int`    | Items per page                  | 20      |
| `search`          | `string` | Search by name, SKU, or barcode | -       |
| `category_id`     | `uuid`   | Filter by category ID           | -       |
| `is_active`       | `bool`   | Filter by active status         | true    |
| `is_stock_active` | `bool`   | Filter by stock active status   | -       |
| `low_stock_only`  | `bool`   | Filter for low stock items      | false   |
| `sort_by`         | `string` | Sort field (name, created_at)   | name    |
| `sort_order`      | `string` | Sort order (asc, desc)          | asc     |

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": "uuid",
      "name": "Product Name",
      "barcode": "123456",
      "sku": "SKU-123",
      "description": "...",
      "category_id": "uuid",
      "unit": "pcs",
      "base_price": 10000,
      "cost_price": 8000,
      "current_stock": 50,
      "min_stock_alert": 10,
      "is_active": true,
      "category": { "id": "uuid", "name": "Category Name" }
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 100, "total_pages": 5 }
}
```

### 2. Create Product

Create a new product (Admin only).

- **URL**: `/products`
- **Method**: `POST`
- **Auth Required**: Yes (Admin only)

#### Request Body

{
"name": "New Product",
"barcode": "123456", // Optional
"sku": "SKU-NEW", // Optional
"description": "Product Description", // Optional
"category_id": "uuid", // Optional
"unit": "pcs",
"base_price": 15000,
"cost_price": 12000,
"is_stock_active": true, // Optional (default false?)
"current_stock": 100, // Optional
"min_stock_alert": 10, // Optional
"image_url": "https://pub-....r2.dev/products/...", // Optional (if not uploading file)
"is_refillable": false,
"pricing_tiers": [{ "min_quantity": 10, "price": 14000, "name": "Grosir" }]
}

````

> [!NOTE]
> **Image Upload**: To upload an image, use `multipart/form-data`.
> - Field `data`: Contains the JSON payload (stringified).
> - Field `image`: Contains the image file.
> - **Optimization**: Backend will automatically resize images to max 800x800px and compress to JPEG (Quality 75).
> - **Max Size**: Please try to keep uploads under 5MB to avoid timeouts, though backend handles large files by resizing.


#### Response (201 Created)

Returns the created product data with **ETag** header.

```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "uuid",
    "name": "New Product",
    ...
  }
}
````

### 3. Get Product By ID

Retrieve details of a specific product.

- **URL**: `/products/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Protected)

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Product details",
  "data": {
    "id": "uuid",
    "name": "Product Name",
    "pricing_tiers": [...],
    ...
  }
}
```

### 4. Update Product

Update an existing product.

- **URL**: `/products/{id}`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin only)

#### Request Body

Same as Create Product, but all fields are optional (partial update).

> [!NOTE]
> **Image Update**: To update the image, use `multipart/form-data` similar to Create Product.
>
> - Field `data`: JSON payload.
> - Field `image`: New image file.
>
> **Image Removal**: To remove an existing image, send a standard JSON `PUT` request with `"image_url": ""` (empty string).
>
> **Best Practices**:
>
> - Frontend should display images using the `image_url` directly.
> - Images are cached heavily (`Cache-Control: public, max-age=1 year`).
> - Use the `ETag` returned in the response headers to update your local cache optimistically.

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Product updated successfully",
  "data": { ...updated_product_data... }
}
```

### 5. Delete Product

Soft delete a product.

- **URL**: `/products/{id}`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin only)

### 6. Search By Barcode

Quickly search for a single product by exact barcode.

- **URL**: `/products/search`
- **Method**: `GET`
- **Auth Required**: Yes (Protected)

#### Query Parameters

| Parameter | Type     | Description            |
| :-------- | :------- | :--------------------- |
| `barcode` | `string` | Exact barcode to match |

#### Response (200 OK)

Product object or 404.

### 7. Manage Pricing Tiers

Manage wholesale pricing for a product.

- **Add Tier**: `POST /products/{id}/pricing-tiers`
- **Update Tier**: `PUT /products/{id}/pricing-tiers/{tierId}`
- **Delete Tier**: `DELETE /products/{id}/pricing-tiers/{tierId}`

#### Tier Request Body

```json
{
  "name": "New Tier",
  "min_quantity": 50,
  "price": 13000
}
```

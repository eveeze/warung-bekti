# Store Locations API

This module provides the ability to manage physical and logical store locations (such as shelves, racks, refrigerators, and displays) where products are placed.

## Endpoints

### 1. List Locations

Retrieves a list of all locations.

**Request:**
`GET /api/v1/locations`

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Locations retrieved successfully",
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Rak Gondola 1",
      "category": "Rak Gondola",
      "x_coordinate": 10.5,
      "y_coordinate": 0.0,
      "z_coordinate": 5.0,
      "width": 120.0,
      "depth": 40.0,
      "height": 180.0,
      "created_at": "2026-03-19T10:00:00Z",
      "updated_at": "2026-03-19T10:00:00Z"
    }
  ],
  "meta": {
    "total": 1
  }
}
```

### 2. Create Location

Creates a new physical or logical store location.

**Request:**
`POST /api/v1/locations`

**Headers:**
`Authorization: Bearer <token>`
`Content-Type: application/json`

**Body:**

```json
{
  "name": "Chiller Minuman A",
  "category": "Chiller",
  "x_coordinate": 2.5,
  "y_coordinate": 0.0,
  "z_coordinate": 1.5,
  "width": 60.0,
  "depth": 60.0,
  "height": 200.0
}
```

**Response (201 Created):**

```json
{
  "success": true,
  "message": "Location created successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174001",
    "name": "Chiller Minuman A",
    "category": "Chiller",
    "x_coordinate": 2.5,
    "y_coordinate": 0.0,
    "z_coordinate": 1.5,
    "width": 60.0,
    "depth": 60.0,
    "height": 200.0,
    "created_at": "2026-03-19T10:15:00Z",
    "updated_at": "2026-03-19T10:15:00Z"
  }
}
```

**Error Response (400 Bad Request - Missing required fields):**

```json
{
  "error": "Name and Category are required"
}
```

### 3. Get Location Detail

Retrieves the details of a single location.

**Request:**
`GET /api/v1/locations/{id}`

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Location retrieved successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174001",
    "name": "Chiller Minuman A",
    ...
  }
}
```

**Error Response (404 Not Found):**

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Location not found"
  }
}
```

### 4. Update Location

Updates an existing location's properties. All fields are optional.

**Request:**
`PUT /api/v1/locations/{id}`

**Headers:**
`Authorization: Bearer <token>`
`Content-Type: application/json`

**Body:**

```json
{
  "name": "Chiller Minuman B",
  "x_coordinate": 3.0
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Location updated successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174001",
    "name": "Chiller Minuman B",
    "category": "Chiller",
    "x_coordinate": 3.0,
    ...
  }
}
```

**Error Response (404 Not Found):**

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Location not found"
  }
}
```

### 5. Delete Location

Deletes a location and implicitly unlinks any products associated with it.

**Request:**
`DELETE /api/v1/locations/{id}`

**Headers:**
`Authorization: Bearer <token>`

**Response (204 No Content):**
_(No body)_

**Error Response (404 Not Found):**

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Location not found"
  }
}
```

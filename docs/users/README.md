# Users Module

Base URL: `/api/v1`

## Endpoints

### 1. List Users

Retrieve a paginated list of users.

- **URL**: `/users`
- **Method**: `GET`
- **Auth Required**: Yes (Admin only)

#### Query Parameters

| Parameter  | Type     | Description             | Default |
| :--------- | :------- | :---------------------- | :------ |
| `page`     | `int`    | Page number             | 1       |
| `per_page` | `int`    | Items per page          | 10      |
| `search`   | `string` | Search by name or email | -       |

#### Response (200 OK)

```json
{
  "data": [
    {
      "id": "uuid-string",
      "name": "User Name",
      "email": "user@example.com",
      "role": "cashier",
      "is_active": true,
      "last_login_at": "...",
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 50,
    "total_pages": 5
  }
}
```

### 2. Create User

Create a new user account (Admin only).

- **URL**: `/users`
- **Method**: `POST`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "name": "New User",
  "email": "user@example.com",
  "password": "password123",
  "role": "admin|cashier|inventory"
}
```

#### Response (201 Created)

```json
{
  "id": "uuid-string",
  "name": "New User",
  "email": "user@example.com",
  "role": "cashier",
  "is_active": true,
  "...": "..."
}
```

### 3. Get User By ID

Retrieve details of a specific user.

- **URL**: `/users/{id}`
- **Method**: `GET`
- **Auth Required**: Yes (Admin only)

#### Response (200 OK)

```json
{
  "id": "uuid-string",
  "name": "User Name",
  "email": "user@example.com",
  "role": "cashier",
  "is_active": true,
  "...": "..."
}
```

### 4. Update User

Update an existing user's details.

- **URL**: `/users/{id}`
- **Method**: `PUT`
- **Auth Required**: Yes (Admin only)

#### Request Body

```json
{
  "name": "Updated Name",
  "email": "updated@example.com",
  "role": "admin",
  "is_active": true,
  "password": "newpassword" // Optional
}
```

#### Response (200 OK)

```json
{
  "id": "uuid-string",
  "name": "Updated Name",
  "email": "updated@example.com",
  "role": "admin",
  "is_active": true,
  "...": "..."
}
```

### 5. Delete User

Soft delete a user account.

- **URL**: `/users/{id}`
- **Method**: `DELETE`
- **Auth Required**: Yes (Admin only)

#### Response (200 OK)

```json
{
  "message": "User deleted successfully"
}
```

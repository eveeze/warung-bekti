# Authentication Module

Base URL: `/api/v1` (except health checks)

## Business Context

Secures the application.

- **JWT**: Stateless authentication.
- **PIN**: Simplified login for POS (Future).

## Frontend Implementation Guide

### 1. Token Storage

- **Web**: `localStorage` (if simple) or `HttpOnly Cookie`.
- **Mobile**: `SecureStore` (Expo) / `Keychain`.

### 2. Auto-Logout (Interceptor)

- Add an Axios/Fetch interceptor.
- If response is `401 Unauthorized`, clear token and redirect to Login screen immediately.

## Endpoints

### 1. Login

Authenticate a user and retrieve access tokens.

- **URL**: `/auth/login`
- **Method**: `POST`
- **Auth Required**: No

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "secretpassword"
}
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "uuid-string",
      "name": "User Name",
      "email": "user@example.com",
      "role": "admin|cashier|inventory",
      "is_active": true,
      "last_login_at": "2023-01-01T12:00:00Z",
      "created_at": "2023-01-01T10:00:00Z",
      "updated_at": "2023-01-01T10:00:00Z"
    }
  }
}
```

### 2. Register

Register a new user (Public endpoint, usually restricted in production but currently open).

- **URL**: `/auth/register`
- **Method**: `POST`
- **Auth Required**: No

#### Request Body

```json
{
  "name": "New User",
  "email": "newuser@example.com",
  "password": "password123",
  "role": "cashier" // Optional, defaults to cashier? Check handler.
}
```

#### Response (201 Created)

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": "uuid-string",
    "name": "New User",
    "email": "newuser@example.com",
    "role": "cashier",
    "is_active": true,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

### 3. Refresh Token

Refresh an expired access token using a refresh token.

- **URL**: `/auth/refresh`
- **Method**: `POST`
- **Auth Required**: No

#### Request Body

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Response (200 OK)

```json
{
  "success": true,
  "message": "Token refreshed",
  "data": {
    "access_token": "new-access-token",
    "refresh_token": "new-refresh-token"
  }
}
```

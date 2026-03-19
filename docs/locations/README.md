# Store Locations Module

This module isolates the documentation and integration guides for the physical/logical Store Locations API and its integrations with the React Native mobile application.

---

## 🤖 AI Agent Integration Checklist

To fully integrate the Location feature into the mobile frontend, the AI Agent must implement the following updates across other feature modules:

- [ ] **Product Input/Edit Forms**: Update the Product creation and editing screens to include a Select/Dropdown component for `location_id`. This should fetch from the `GET /api/v1/locations` endpoint.
- [ ] **Product Details Screen**: When viewing a product, parse the nested `location` object inside the product response and display the exact Rack/Chiller name where the product is located.
- [ ] **Data Models / Types**: Update the TypeScript definition for `Product` to include `location_id?: string;` and `location?: Location;`.
- [ ] **Inventory/Stock UI**: Display location badges in the stock opname lists to help staff physically navigate the store.

---

## 📦 React Native 3D Visualization Guide

The `Location` model provides X, Y, Z coordinates alongside physical dimensions (Width, Depth, Height). This enables the creation of an immersive 3D store map or an interactive 2D top-down isometric view.

### Recommended Stack

For true 3D on React Native, use **`@react-three/fiber`** and **`expo-gl`**. If a simpler isometric 2D view is preferred, **`react-native-skia`** or standard `react-native-reanimated` SVGs with orthographic projection mathematics can be used.

### Implementation Blueprint (`react-three-fiber`)

1. **Install Dependencies:**

```bash
npm install three @react-three/fiber @react-three/drei expo-gl
```

2. **Parsing Model Coordinates:**
   The backend coordinates (`x_coordinate`, `z_coordinate`) define the position on the floor plan. The `y_coordinate` represents the vertical height off the ground (e.g. shelf tier).

3. **React Native Component Boilerplate:**

```tsx
import React, { useMemo } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls, Box } from '@react-three/drei';

type API_Location = {
  id: string;
  name: string;
  x_coordinate: number;
  y_coordinate: number;
  z_coordinate: number;
  width: number;
  height: number;
  depth: number;
};

const StoreShelf = ({ location }: { location: API_Location }) => {
  // Convert API dimensions (e.g. cm) to Three.js units (e.g. meters)
  const scale = 0.01;
  const position = useMemo(
    () => [
      location.x_coordinate * scale,
      location.y_coordinate * scale,
      location.z_coordinate * scale,
    ],
    [location],
  );

  const geometry = useMemo(
    () => [
      location.width * scale,
      location.height * scale,
      location.depth * scale,
    ],
    [location],
  );

  return (
    <Box position={position} args={geometry}>
      <meshStandardMaterial color={0x808080} transparent opacity={0.8} />
    </Box>
  );
};

export const StoreMap3D = ({ locations }) => {
  return (
    <Canvas camera={{ position: [5, 5, 5], fov: 50 }}>
      <ambientLight intensity={0.5} />
      <directionalLight position={[10, 10, 5]} intensity={1} />

      {locations.map((loc) => (
        <StoreShelf key={loc.id} location={loc} />
      ))}

      <OrbitControls enablePan enableZoom />
    </Canvas>
  );
};
```

_(Note for AI Agent: Implement smooth camera panning (`react-spring/three`) towards a specific shelf when a user searches for a specific product!)_

---

## 📡 API Reference

### 1. List Locations

Retrieves a list of all locations.

**Request:** `GET /api/v1/locations`

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
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }
  ],
  "meta": {
    "total": 1
  }
}
```

### 2. Create Location

Creates a new physical store location.

**Request:** `POST /api/v1/locations`

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
  "data": { ... }
}
```

**Error Response (400 Bad Request):**

```json
{
  "error": "Name and Category are required"
}
```

### 3. Get Location Detail

**Request:** `GET /api/v1/locations/{id}`

**Response (200 OK):**

```json
{
  "success": true,
  "data": { ... }
}
```

**Error Response (404 Not Found):**

```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "Location not found" }
}
```

### 4. Update Location

**Request:** `PUT /api/v1/locations/{id}`

**Body:** (All optional)

```json
{
  "name": "Chiller B",
  "x_coordinate": 3.0
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Location updated successfully",
  "data": { ... }
}
```

**Error Response (404 Not Found):**

```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "Location not found" }
}
```

### 5. Delete Location

**Request:** `DELETE /api/v1/locations/{id}`

**Response (204 No Content):** _(No body)_

**Error Response (404 Not Found):**

```json
{
  "success": false,
  "error": { "code": "NOT_FOUND", "message": "Location not found" }
}
```

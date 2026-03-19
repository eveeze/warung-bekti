# Agentic AI Instruction: WarungOS 3D Store Map Prototype

**Context for the AI:**
You are an expert Frontend Developer specializing in **React Native**, **Expo**, **Three.js**, and **React Three Fiber (R3F)**.
Your task is to build a functional, interactive 3D Web/Expo prototype of a supermarket store layout. This prototype will be used to visualize product locations inside a physical store based on coordinates provided by a Go backend API.

## 1. Project Requirements

Build a React/Expo application that renders a 3D isometric store map.

1. **The Scene:** An isometric 3D canvas `(Camera FOV: 50, Position: [15, 15, 15])` showcasing a grid floor.
2. **The Objects (Locations):** Render 3D boxes/meshes representing store shelves, fridges, and warehouse areas based on the `locations` JSON array.
3. **The Interactivity:** The app must have a "Search Product" simulation. When a specific product ID is selected, the specific shelf (Location) that holds the product must **glow**, pulse, or be highlighted with an emissive material or floating 3D Pin.
4. **Tech Stack:**
   - React / Expo (Web/Native compatible code)
   - `@react-three/fiber`
   - `@react-three/drei` (for OrbitControls to allow zooming and panning)
   - TailwindCSS (for the 2D floating UI overlay)
5. **UI Overlay:** Place a floating Search Bar at the top and a floating "Product Info Card" at the bottom using glassmorphism.

## 2. Mock API Data (Dummy Data)

Please hardcode the following JSON responses into your prototype to simulate the backend.

### A. Locations Data (`GET /api/v1/locations`)

This defines the physical shelves in the 3D space. Note that coordinates are relative to the store center `(0,0,0)`.

```json
{
  "success": true,
  "data": [
    {
      "id": "loc-shelf-a1",
      "name": "Aisle 1 - Snacks",
      "category": "shelf",
      "x_coordinate": -5.0,
      "y_coordinate": 0.5,
      "z_coordinate": -2.0,
      "width": 4.0,
      "height": 2.5,
      "depth": 1.0,
      "description": "Main snack aisle near entrance"
    },
    {
      "id": "loc-fridge-1",
      "name": "Beverage Fridge",
      "category": "fridge",
      "x_coordinate": 6.0,
      "y_coordinate": 1.0,
      "z_coordinate": -5.0,
      "width": 3.0,
      "height": 3.0,
      "depth": 1.5,
      "description": "Cold drinks"
    },
    {
      "id": "loc-cashier",
      "name": "Main Cashier",
      "category": "cashier_area",
      "x_coordinate": 0.0,
      "y_coordinate": 0.5,
      "z_coordinate": 8.0,
      "width": 3.0,
      "height": 1.0,
      "depth": 2.0,
      "description": "Checkout counter"
    }
  ]
}
```

### B. Product Data (`GET /public/products/{id}`)

This represents the product the user is looking for. It contains a `location_id` pointing to the shelves above.

```json
{
  "success": true,
  "data": {
    "id": "prod-123",
    "name": "Better Caramel Biscuit",
    "description": "Sweet caramel biscuits",
    "base_price": 2500,
    "image_url": "https://example.com/biscuit.jpg",
    "location_id": "loc-shelf-a1",
    "category": {
      "name": "Snacks"
    }
  }
}
```

## 3. Implementation Steps for the AI

Please generate the complete codebase for this prototype in a single structured response or via multiple files:

1. **Scene Setup:** Initialize `<Canvas>` with lighting (`ambientLight`, `directionalLight`). Use `OrbitControls` restricted to a pleasant isometric viewing angle.
2. **Mesh Generation:** Map over the `locations` array. Return a `<mesh>` for each item. Apply `x,y,z` to `position` and `width,height,depth` to `scale` (or `args` of `boxGeometry`).
3. **Materials:** Apply distinct colors based on the `category` (e.g., standard shelves are wooden brown `#8B5A2B`, fridges are glass/blue `#add8e6`, cashier is gray `#a9a9a9`).
4. **Highlight State:** Accept a prop `activeLocationId`. If `mesh.id === activeLocationId`, override the material to pulse with a bright neon orange/red emissive map.
5. **App Component:** Tie the 3D Canvas and the HTML UI Overlay together. Add a simple dropdown or button to "Simulate Scanning Better Caramel Biscuit" which triggers the active state on `loc-shelf-a1`.

**Output Format:** Provide the exact `App.jsx` and `StoreModel.jsx` files ready to be pasted into a CodeSandbox or an Expo Snack. Make sure the code works gracefully on web browsers first before mobile porting.

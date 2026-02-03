# Frontend Implementation Guide: Consignment Integration

This guide details the changes required in the Frontend application to support the new **Consignor-Product Linking** features.

## 1. Product Form (Create & Edit)

We need to allow users to assign a "Supplier" or "Consignor" when creating or editing a product.

### UI Changes

- Add a **Dropdown/Select** input field labeled **"Consignor / Supplier"**.
- Place it ideally near "Category" or "Cost Price".
- The value should be the `id` of the consignor.

### Data Fetching

- **Fetch Consignors**: Call `GET /api/v1/consignors` to populate the dropdown options.
- **Cache**: This list changes infrequently. Use `staleTime: 5 minutes` or similar.

### Form Submission

- Include `consignor_id` in the payload if selected.
- If the user clears the selection, send `null` or omit the field (if backend allows). *Note: Currently sending `null` might require backend adjustment if not pointer, but our backend uses `*uuid.UUID`, so json `null` is valid.\*

**Payload Example:**

```json
{
  "name": "Kripik Singkong",
  "base_price": 5000,
  "category_id": "...",
  "consignor_id": "8939670d-f538-4034-8c76-589578278219" // <--- NEW FIELD
}
```

---

## 2. Inventory / Product List

We need to allow users to filter products to see only those belonging to a specific Consignor.

### UI Changes

- Add a **Filter Button** or **Dropdown** in the Product List header.
- Label: "Filter by Consignor".
- Options: List of all Consignors (fetch from `GET /api/v1/consignors`).

### API Integration

- Pass the selected ID as a query parameter to the list endpoint.
- Endpoint: `GET /api/v1/products`
- Parameter: `consignor_id={UUID}`

**Example URL:**
`/api/v1/products?page=1&per_page=20&consignor_id=8939670d-f538-4034-8c76-589578278219`

---

## 3. Consignor Detail Page (Optional / Future)

On the Consignor management page, it would be helpful to show a list of products they own.

- **Implementation**: Reuse the **Product List Component**.
- **Props**: Pass `filter={{ consignor_id: currentConsignor.id }}` to the list component.

---

## Summary of API Changes

| Feature            | Endpoint         | Method | New Parameter / Field          |
| :----------------- | :--------------- | :----- | :----------------------------- |
| **List Products**  | `/products`      | `GET`  | `?consignor_id=UUID`           |
| **Create Product** | `/products`      | `POST` | Body: `"consignor_id": "UUID"` |
| **Update Product** | `/products/{id}` | `PUT`  | Body: `"consignor_id": "UUID"` |

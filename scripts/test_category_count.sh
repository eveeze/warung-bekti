#!/bin/bash

BASE_URL="http://localhost:8080"
EMAIL="admin@warung.com"
PASSWORD="password"

# Login
LOGIN_RES=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")
TOKEN=$(echo $LOGIN_RES | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

# 1. Create Category
echo "Creating Category..."
CAT_RES=$(curl -s -X POST "$BASE_URL/api/v1/categories" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"name": "TestCount", "description": "Testing Count"}')
CAT_ID=$(echo $CAT_RES | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Category ID: $CAT_ID"

# 2. Create Product
# We need a dummy product JSON.
echo "Creating Product..."
PROD_RES=$(curl -s -X POST "$BASE_URL/api/v1/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"barcode\": \"TEST-$(date +%s)\",
    \"sku\": \"SKU-$(date +%s)\",
    \"name\": \"Test Product\",
    \"description\": \"For testing categories\",
    \"category_id\": \"$CAT_ID\",
    \"unit\": \"pcs\",
    \"base_price\": 1000,
    \"cost_price\": 800,
    \"is_stock_active\": false,
    \"min_stock_alert\": 0,
    \"max_stock\": 0,
    \"pricing_tiers\": []
  }")
PROD_ID=$(echo $PROD_RES | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Product ID: $PROD_ID"

# 3. Check Count
echo "Checking Product Count (Should be 1)..."
LIST_RES=$(curl -s -X GET "$BASE_URL/api/v1/categories" -H "Authorization: Bearer $TOKEN")
echo $LIST_RES | grep "\"product_count\":1" && echo "PASS: Count is 1" || echo "FAIL: Count is not 1"

# 4. Try Delete Category (Should Fail)
echo "Tentative Delete Category (Should Fail)..."
DEL_RES=$(curl -s -X DELETE "$BASE_URL/api/v1/categories/$CAT_ID" -H "Authorization: Bearer $TOKEN")
echo $DEL_RES
if echo $DEL_RES | grep -q "cannot delete"; then
    echo "PASS: Deletion blocked"
else
    echo "FAIL: Deletion NOT blocked"
fi

# 5. Delete Product
echo "Deleting Product..."
curl -s -X DELETE "$BASE_URL/api/v1/products/$PROD_ID" -H "Authorization: Bearer $TOKEN"

# 6. Delete Category (Should Succeed)
echo "Retry Delete Category..."
FINAL_DEL=$(curl -s -X DELETE "$BASE_URL/api/v1/categories/$CAT_ID" -H "Authorization: Bearer $TOKEN")
echo $FINAL_DEL

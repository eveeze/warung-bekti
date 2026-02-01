#!/bin/bash

# Configuration
BASE_URL="http://localhost:8080"
EMAIL="admin@warung.com"
PASSWORD="password"

echo "Logging in as Admin..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Login failed!"
  echo $LOGIN_RESPONSE
  exit 1
fi
echo "Login successful. Token: ${TOKEN:0:10}..."

echo -e "\n1. Create Category 'Drinks'..."
CREATE_RESP=$(curl -s -X POST "$BASE_URL/api/v1/categories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Drinks", "description": "Minuman Segar"}')
echo $CREATE_RESP

CATEGORY_ID=$(echo $CREATE_RESP | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created Category ID: $CATEGORY_ID"

echo -e "\n2. List Categories (expect 1)..."
curl -s -X GET "$BASE_URL/api/v1/categories" \
  -H "Authorization: Bearer $TOKEN" | grep "Drinks"

echo -e "\n3. Update Category..."
curl -s -X PUT "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Beverages", "description": "Aneka Minuman"}'

echo -e "\n4. Verify Update..."
curl -s -X GET "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n5. Delete Category..."
curl -s -X DELETE "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n6. Verify Deletion (expect 404 or empty list)..."
curl -s -X GET "$BASE_URL/api/v1/categories/$CATEGORY_ID" \
  -H "Authorization: Bearer $TOKEN"

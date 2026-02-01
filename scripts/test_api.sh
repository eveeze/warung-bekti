#!/bin/bash

BASE_URL="http://localhost:8080"
EMAIL="admin@warung.com"
PASSWORD="password"

echo "Testing Admin Login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

echo "Response: $LOGIN_RESPONSE"

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Failed to get token. Exiting."
  exit 1
fi

echo "Token: $TOKEN"

echo -e "\n--- Products ---"
curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/api/v1/products" -H "Authorization: Bearer $TOKEN"
echo " (Status)"

echo -e "\n--- Customers ---"
curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/api/v1/customers" -H "Authorization: Bearer $TOKEN"
echo " (Status)"

echo -e "\n--- Transactions ---"
curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/api/v1/transactions" -H "Authorization: Bearer $TOKEN"
echo " (Status)"

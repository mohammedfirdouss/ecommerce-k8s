#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost}"
TEST_EMAIL="test-$(date +%s)@example.com"
PASS=0
FAIL=0

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

check() {
    local desc="$1"
    local expected_code="$2"
    shift 2
    
    response=$(curl -s -w "\n%{http_code}" "$@")
    code=$(echo "$response" | tail -1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC}: ${desc} (HTTP ${code})"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: ${desc} (expected ${expected_code}, got ${code})"
        echo "  Response: ${body}"
        FAIL=$((FAIL + 1))
    fi
    
    # Return body for chaining
    echo "$body"
}

echo "  E-Commerce E2E Test Suite"
echo ""

# 1. Health checks
echo "--- Health Checks ---"
check "Auth health" "200" "${BASE_URL}/api/auth/healthz" > /dev/null
check "Product health" "200" "${BASE_URL}/api/products/healthz" > /dev/null
check "Cart health" "200" "${BASE_URL}/api/cart/healthz" > /dev/null
check "Order health" "200" "${BASE_URL}/api/orders/healthz" > /dev/null
check "Payment health" "200" "${BASE_URL}/api/payments/healthz" > /dev/null
echo ""

# 2. Register a user
echo "--- User Registration ---"
REGISTER_RESP=$(check "Register user" "201" \
    -X POST "${BASE_URL}/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"'"$TEST_EMAIL"'","password":"password123"}')
TOKEN=$(echo "$REGISTER_RESP" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)
USER_ID=$(echo "$REGISTER_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "  Token: ${TOKEN:0:20}..."
echo "  User ID: ${USER_ID}"
echo ""

# 3. Login
echo "--- User Login ---"
check "Login user" "200" \
    -X POST "${BASE_URL}/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"'"$TEST_EMAIL"'","password":"password123"}' > /dev/null
echo ""

# 4. Create products
echo "--- Product Catalog ---"
PRODUCT1_RESP=$(check "Create product 1" "201" \
    -X POST "${BASE_URL}/api/products/" \
    -H "Content-Type: application/json" \
    -d '{"name":"Laptop","description":"A powerful laptop","price":999.99,"stock":10}')
PRODUCT1_ID=$(echo "$PRODUCT1_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

PRODUCT2_RESP=$(check "Create product 2" "201" \
    -X POST "${BASE_URL}/api/products/" \
    -H "Content-Type: application/json" \
    -d '{"name":"Mouse","description":"Wireless mouse","price":29.99,"stock":50}')
PRODUCT2_ID=$(echo "$PRODUCT2_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

check "List products" "200" "${BASE_URL}/api/products/" > /dev/null
check "Get product" "200" "${BASE_URL}/api/products/${PRODUCT1_ID}" > /dev/null
echo ""

# 5. Shopping cart
echo "--- Shopping Cart ---"
check "Add item to cart" "201" \
    -X POST "${BASE_URL}/api/cart/items" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: ${USER_ID}" \
    -d "{\"product_id\":\"${PRODUCT1_ID}\",\"quantity\":1,\"price\":999.99}" > /dev/null

check "Add second item" "201" \
    -X POST "${BASE_URL}/api/cart/items" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: ${USER_ID}" \
    -d "{\"product_id\":\"${PRODUCT2_ID}\",\"quantity\":2,\"price\":29.99}" > /dev/null

check "Get cart" "200" \
    "${BASE_URL}/api/cart/" \
    -H "X-User-ID: ${USER_ID}" > /dev/null
echo ""

# 6. Create order
echo "--- Orders ---"
ORDER_RESP=$(check "Create order" "201" \
    -X POST "${BASE_URL}/api/orders/" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: ${USER_ID}" \
    -d "{\"items\":[{\"product_id\":\"${PRODUCT1_ID}\",\"quantity\":1,\"price\":999.99},{\"product_id\":\"${PRODUCT2_ID}\",\"quantity\":2,\"price\":29.99}]}")
ORDER_ID=$(echo "$ORDER_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "  Order ID: ${ORDER_ID}"

check "List orders" "200" \
    "${BASE_URL}/api/orders/" \
    -H "X-User-ID: ${USER_ID}" > /dev/null

check "Get order" "200" \
    "${BASE_URL}/api/orders/${ORDER_ID}" > /dev/null
echo ""

# 7. Wait for payment processing
echo "--- Payment Processing ---"
echo "Waiting 5 seconds for async payment processing..."
sleep 5

check "Get payment status" "200" \
    "${BASE_URL}/api/payments/${ORDER_ID}" > /dev/null
echo ""

# Summary
echo "  Results: ${PASS} passed, ${FAIL} failed"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi

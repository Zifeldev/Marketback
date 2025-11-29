#!/usr/bin/env bash
set -euo pipefail

AUTH_URL=${AUTH_URL:-"http://localhost:8081"}
MARKET_URL=${MARKET_URL:-"http://localhost:8080"}

EMAIL=${BUYER_EMAIL:-"smoke_$(date +%s)@example.com"}
PASSWORD=${BUYER_PASSWORD:-"Sm0keTest123!"}
ROLE=${BUYER_ROLE:-"user"}

echo "[smoke] register $EMAIL"
RESP=$(curl -s -X POST "$AUTH_URL/auth/register" -H "Content-Type: application/json" -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"role\":\"$ROLE\"}")
echo "[smoke] auth response: $RESP"

TOKEN=""
# Try jq/python/sed to extract access_token
if command -v jq >/dev/null 2>&1; then
	TOKEN=$(echo "$RESP" | jq -r .access_token)
fi
if [ -z "$TOKEN" ] && command -v python3 >/dev/null 2>&1; then
	TOKEN=$(echo "$RESP" | python3 -c 'import sys, json
try:
 d=json.load(sys.stdin)
 print(d.get("access_token",""))
except:
 print("")')
fi
if [ -z "$TOKEN" ]; then
	# fallback to sed extraction
	TOKEN=$(echo "$RESP" | sed -n 's/.*"access_token":"\([^"\\]*\)".*/\1/p')
fi

if [ -z "$TOKEN" ]; then
	echo "[smoke] register did not return token, trying login"
	LOGIN_RESP=$(curl -s -X POST "$AUTH_URL/auth/login" -H "Content-Type: application/json" -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
	echo "[smoke] login response: $LOGIN_RESP"
	if command -v jq >/dev/null 2>&1; then
		TOKEN=$(echo "$LOGIN_RESP" | jq -r .access_token)
	fi
	if [ -z "$TOKEN" ] && command -v python3 >/dev/null 2>&1; then
		TOKEN=$(echo "$LOGIN_RESP" | python3 -c 'import sys, json
try:
 d=json.load(sys.stdin)
 print(d.get("access_token",""))
except:
 print("")')
	fi
	if [ -z "$TOKEN" ]; then
		TOKEN=$(echo "$LOGIN_RESP" | sed -n 's/.*"access_token":"\([^"\\]*\)".*/\1/p')
	fi
	if [ -z "$TOKEN" ]; then
		echo "[smoke] failed to obtain access_token"
		exit 1
	fi
fi

echo "[smoke] token obtained (len=${#TOKEN})"

echo "[smoke] add to cart (product_id=${PRODUCT_ID:-1})"
RESP2=$(curl -s -w "\n%{http_code}" -X POST "$MARKET_URL/api/cart/items" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "{\"product_id\":${PRODUCT_ID:-1},\"quantity\":1,\"size\":\"M\",\"color\":\"black\"}")
BODY=$(echo "$RESP2" | sed '$d')
STATUS=$(echo "$RESP2" | tail -n1)
echo "[smoke] add-to-cart status=$STATUS body=$BODY"

if [ "$STATUS" != "201" ]; then
	echo "[smoke] add-to-cart failed"
	exit 1
fi

echo "[smoke] get cart"
CART=$(curl -s -X GET "$MARKET_URL/api/cart" -H "Authorization: Bearer $TOKEN")
echo "[smoke] cart: $CART"

echo "[smoke] create order"
ORDER_RESP=$(curl -s -w "\n%{http_code}" -X POST "$MARKET_URL/api/user/orders" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"payment_method":"card","delivery_address":"Smoke Street 1"}')
ORDER_BODY=$(echo "$ORDER_RESP" | sed '$d')
ORDER_STATUS=$(echo "$ORDER_RESP" | tail -n1)
echo "[smoke] create-order status=$ORDER_STATUS body=$ORDER_BODY"

echo "[smoke] list my orders"
MY_ORDERS=$(curl -s -X GET "$MARKET_URL/api/user/orders" -H "Authorization: Bearer $TOKEN")
echo "[smoke] my orders: $MY_ORDERS"

echo "[smoke] smoke test finished successfully"
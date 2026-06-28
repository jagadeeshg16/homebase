#!/bin/bash
set -euo pipefail

CURRENT_IP=$(curl -s https://api.ipify.org)
IP_FILE="/tmp/.last_home_ip"
STORED_IP=$(cat "$IP_FILE" 2>/dev/null || echo "")

if [ "$CURRENT_IP" = "$STORED_IP" ]; then
    exit 0
fi

echo "$(date): IP changed $STORED_IP -> $CURRENT_IP"

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/dns/update \
    -H "X-Internal-Token: ${INTERNAL_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"ip\": \"$CURRENT_IP\"}")

HTTP_CODE=$(echo "$RESPONSE" | tail -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo "$CURRENT_IP" > "$IP_FILE"
    echo "$(date): DNS updated successfully to $CURRENT_IP"
else
    echo "$(date): DNS update failed (HTTP $HTTP_CODE)" >&2
    exit 1
fi

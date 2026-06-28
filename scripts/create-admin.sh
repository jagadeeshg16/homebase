#!/bin/bash
# Run once to create the admin user.
# Usage: ./create-admin.sh <username> <password>

set -euo pipefail

USERNAME="${1:-jagadeesh}"
PASSWORD="${2:?Usage: $0 <username> <password>}"
DB_PATH="${DB_PATH:-$HOME/server/data/server.db}"

HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw(b'$PASSWORD', bcrypt.gensalt()).decode())" 2>/dev/null || \
      htpasswd -bnBC 10 "" "$PASSWORD" | tr -d ':\n' | sed 's/\$2y/\$2a/')

sqlite3 "$DB_PATH" \
    "INSERT OR REPLACE INTO users (username, password_hash) VALUES ('$USERNAME', '$HASH');"

echo "Admin user '$USERNAME' created."

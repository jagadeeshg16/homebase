#!/bin/bash
# Run once to create the admin user.
# Usage: ./create-admin.sh <username> <password>

set -euo pipefail

USERNAME="${1:-jagadeesh}"
PASSWORD="${2:?Usage: $0 <username> <password>}"
DB_PATH="${DB_PATH:-$HOME/server/data/server.db}"

mkdir -p "$(dirname "$DB_PATH")"

python3 - <<EOF
import bcrypt, sqlite3

db = sqlite3.connect("$DB_PATH")
db.execute("""
    CREATE TABLE IF NOT EXISTS users (
        id            INTEGER PRIMARY KEY,
        username      TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
    )
""")
hash = bcrypt.hashpw(b"$PASSWORD", bcrypt.gensalt()).decode()
db.execute("INSERT OR REPLACE INTO users (username, password_hash) VALUES (?, ?)", ("$USERNAME", hash))
db.commit()
db.close()
print("Admin user '$USERNAME' created.")
EOF

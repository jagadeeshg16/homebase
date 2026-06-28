#!/bin/bash
# homebase setup — run once on a fresh machine
# Works from any directory: bash ~/server/scripts/setup.sh

set -euo pipefail

# Resolve project root regardless of where this script is called from
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "==> homebase setup"
echo "    Project root: $PROJECT_ROOT"
echo ""

echo "==> Installing dependencies..."

# Caddy
if ! command -v caddy &>/dev/null; then
    echo "  Installing Caddy..."
    sudo apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
    sudo apt-get update && sudo apt-get install -y caddy
else
    echo "  Caddy: $(caddy version)"
fi

# Go
if ! command -v go &>/dev/null; then
    echo "  Installing Go..."
    GO_VERSION="1.23.4"
    curl -Lo /tmp/go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
else
    echo "  Go: $(go version)"
fi

# Python bcrypt
echo "  Installing Python bcrypt..."
pip3 install bcrypt --break-system-packages 2>/dev/null || pip3 install bcrypt

# Node.js
if ! command -v node &>/dev/null; then
    echo "  Installing Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
    sudo apt-get install -y nodejs
else
    echo "  Node: $(node --version)"
fi

echo ""
echo "==> Building backend..."
cd "$PROJECT_ROOT/backend"
go build -o homeserver .
echo "  Done."

echo ""
echo "==> Building frontend..."
cd "$PROJECT_ROOT/frontend"
npm install && npm run build
echo "  Done."

echo ""
echo "==> Setting up .env..."
ENV_FILE="$PROJECT_ROOT/.env"
if [ ! -f "$ENV_FILE" ]; then
    cp "$PROJECT_ROOT/.env.example" "$ENV_FILE"
    INTERNAL_TOKEN=$(python3 -c "import secrets; print(secrets.token_hex(32))")
    SESSION_SECRET=$(python3 -c "import secrets; print(secrets.token_hex(32))")
    sed -i "s/INTERNAL_TOKEN=.*/INTERNAL_TOKEN=$INTERNAL_TOKEN/" "$ENV_FILE"
    sed -i "s/SESSION_SECRET=.*/SESSION_SECRET=$SESSION_SECRET/" "$ENV_FILE"
    echo "  .env created with random secrets."
else
    echo "  .env already exists, skipping."
fi

echo ""
echo "==> Creating data directory..."
mkdir -p "$PROJECT_ROOT/data"

echo ""
echo "==> Done! Next steps:"
echo ""
echo "  1. Fill in your domain and DNS keys:"
echo "     nano $PROJECT_ROOT/.env"
echo ""
echo "  2. Create admin user:"
echo "     $PROJECT_ROOT/scripts/create-admin.sh jagadeesh yourpassword"
echo ""
echo "  3. Start the server:"
echo "     $PROJECT_ROOT/backend/homeserver"
echo ""
echo "  4. Start Caddy:"
echo "     caddy run --config $PROJECT_ROOT/caddy/Caddyfile"
echo ""
echo "  5. Add DDNS to cron (crontab -e):"
echo "     */5 * * * * INTERNAL_TOKEN=\$(grep INTERNAL_TOKEN $PROJECT_ROOT/.env | cut -d= -f2) $PROJECT_ROOT/scripts/ddns.sh"

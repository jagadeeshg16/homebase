#!/bin/bash
# homebase setup — run once on a fresh machine
# Usage: ./scripts/setup.sh

set -euo pipefail

echo "==> Installing dependencies..."

# Caddy
if ! command -v caddy &>/dev/null; then
    echo "  Installing Caddy..."
    sudo apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
    sudo apt-get update && sudo apt-get install -y caddy
else
    echo "  Caddy already installed: $(caddy version)"
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
    echo "  Go already installed: $(go version)"
fi

# Python bcrypt
echo "  Installing Python bcrypt..."
pip3 install bcrypt --break-system-packages 2>/dev/null || pip3 install bcrypt

# Node.js (for frontend build)
if ! command -v node &>/dev/null; then
    echo "  Installing Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
    sudo apt-get install -y nodejs
else
    echo "  Node already installed: $(node --version)"
fi

echo ""
echo "==> Building backend..."
cd "$(dirname "$0")/../backend"
go build -o homeserver .
echo "  Backend built."

echo ""
echo "==> Building frontend..."
cd "$(dirname "$0")/../frontend"
npm install && npm run build
echo "  Frontend built."

echo ""
echo "==> Setting up .env..."
ENV_FILE="$(dirname "$0")/../.env"
if [ ! -f "$ENV_FILE" ]; then
    cp "$(dirname "$0")/../.env.example" "$ENV_FILE"

    # generate random secrets
    INTERNAL_TOKEN=$(python3 -c "import secrets; print(secrets.token_hex(32))")
    SESSION_SECRET=$(python3 -c "import secrets; print(secrets.token_hex(32))")

    sed -i "s/INTERNAL_TOKEN=.*/INTERNAL_TOKEN=$INTERNAL_TOKEN/" "$ENV_FILE"
    sed -i "s/SESSION_SECRET=.*/SESSION_SECRET=$SESSION_SECRET/" "$ENV_FILE"

    echo "  .env created with random secrets."
    echo "  Edit ~/server/.env to add your domain and DNS API keys."
else
    echo "  .env already exists, skipping."
fi

echo ""
echo "==> Creating data directory..."
mkdir -p "$(dirname "$0")/../data"

echo ""
echo "==> Done! Next steps:"
echo ""
echo "  1. Fill in your domain and DNS keys:"
echo "     nano ~/server/.env"
echo ""
echo "  2. Create admin user:"
echo "     ~/server/scripts/create-admin.sh jagadeesh yourpassword"
echo ""
echo "  3. Start the server:"
echo "     ~/server/backend/homeserver"
echo ""
echo "  4. Start Caddy:"
echo "     caddy run --config ~/server/caddy/Caddyfile"
echo ""
echo "  5. Add DDNS to cron (crontab -e):"
echo "     */5 * * * * INTERNAL_TOKEN=\$(grep INTERNAL_TOKEN ~/server/.env | cut -d= -f2) ~/server/scripts/ddns.sh"

#!/bin/bash
# homebase runner
# Usage:
#   ./scripts/run.sh -b    build backend
#   ./scripts/run.sh -f    build frontend
#   ./scripts/run.sh -s    start server
#   ./scripts/run.sh -a    build all + start
#   ./scripts/run.sh -b -f build both (no start)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DO_BACKEND=false
DO_FRONTEND=false
DO_START=false

if [ $# -eq 0 ]; then
    echo "Usage: $0 [-b] [-f] [-s] [-a]"
    echo "  -b  build backend"
    echo "  -f  build frontend"
    echo "  -s  start server"
    echo "  -a  build backend + frontend + start"
    exit 1
fi

while getopts "bfsa" opt; do
    case $opt in
        b) DO_BACKEND=true ;;
        f) DO_FRONTEND=true ;;
        s) DO_START=true ;;
        a) DO_BACKEND=true; DO_FRONTEND=true; DO_START=true ;;
        *) echo "Unknown option: -$opt"; exit 1 ;;
    esac
done

if $DO_BACKEND; then
    echo "==> Building backend..."
    cd "$ROOT/backend"
    go build -o homeserver .
    echo "    Done — $ROOT/backend/homeserver"
fi

if $DO_FRONTEND; then
    echo "==> Building frontend..."
    cd "$ROOT/frontend"
    npm install --silent
    npm run build
    echo "    Done — $ROOT/sites/admin/"
fi

if $DO_START; then
    echo "==> Starting homeserver..."
    cd "$ROOT/backend"
    exec ./homeserver
fi

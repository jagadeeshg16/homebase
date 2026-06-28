# homebase

A self-hosted home server platform. Serve your portfolio, spin up subdomains, and manage everything from a single admin dashboard — all running from your own machine.

## What it does

- **Portfolio** at your root domain (`jagadeeshg.in`)
- **Dynamic subdomains** — drop a folder into `sites/` and it's auto-detected
- **Admin dashboard** — toggle subdomains public/private, set rate limits, monitor health
- **DDNS** — automatically keeps your domain pointed at your home IP even when it changes
- **Privacy gate** — private subdomains are password protected
- **Pluggable DNS** — works with GoDaddy or Cloudflare

## Stack

| Layer | Tech |
|---|---|
| Reverse proxy | Caddy (auto HTTPS) |
| Backend | Go |
| Admin UI | React + Vite |
| Database | SQLite |
| DNS | GoDaddy / Cloudflare API |

## Project structure

```
homebase/
├── backend/          # Go backend — API + subdomain serving
│   ├── config/       # Config from .env
│   ├── db/           # SQLite setup + schema
│   ├── dns/          # GoDaddy + Cloudflare providers
│   ├── handlers/     # API route handlers
│   ├── middleware/   # Auth + rate limiting
│   └── server/       # Static file serving + folder watcher
├── frontend/         # React admin dashboard (source)
├── sites/
│   ├── root/         # Portfolio — served at root domain
│   └── {name}/       # Each subdomain's files
├── caddy/
│   └── Caddyfile
├── scripts/
│   ├── ddns.sh       # DDNS cron script
│   └── create-admin.sh
└── .env.example
```

## Setup

### 1. Clone and configure

```bash
git clone https://github.com/jagadeeshg16/homebase.git ~/server
cp ~/server/.env.example ~/server/.env
# fill in your DNS API keys and domain in .env
```

### 2. Build the backend

```bash
cd ~/server/backend
go build -o homeserver .
```

### 3. Create admin user

```bash
~/server/scripts/create-admin.sh jagadeesh yourpassword
```

### 4. Start the server

```bash
cd ~/server/backend && ./homeserver
```

### 5. Start Caddy

```bash
caddy run --config ~/server/caddy/Caddyfile
```

### 6. Set up DDNS cron

```bash
# Add to crontab (crontab -e)
*/5 * * * * INTERNAL_TOKEN=your-token ~/server/scripts/ddns.sh >> ~/server/data/ddns.log 2>&1
```

## Admin dashboard

Visit `admin.yourdomain.com` → login with your admin credentials.

**Pages:**
- **Dashboard** — overview of active subdomains and health
- **Subdomains** — add, delete, toggle public/private, set rate limits
- **Health** — live status of all subdomains
- **DNS** — current IP, manual DDNS trigger
- **Settings** — change admin password

### Build the frontend

```bash
cd ~/server/frontend
npm install && npm run build
# output goes to sites/admin/ automatically
```

### Dev mode (hot reload)

```bash
cd ~/server/frontend && npm run dev
# proxies /api to localhost:8080
```

## Adding a subdomain

**Option 1 — from admin dashboard:**
Add via the Subdomains page. Folder is created automatically, DNS record registered, live immediately (private by default).

**Option 2 — drop a folder:**
```bash
mkdir ~/server/sites/myblog
# copy your files in
# → auto-detected, shows up in admin as inactive+private
# → flip active from dashboard when ready
```

## Environment variables

```env
PORT=8080
INTERNAL_TOKEN=          # secret for DDNS script → backend calls
SESSION_SECRET=          # session cookie signing key

DNS_PROVIDER=godaddy     # or cloudflare
ROOT_DOMAIN=jagadeeshg.in

GODADDY_API_KEY=
GODADDY_API_SECRET=

CLOUDFLARE_API_TOKEN=
CLOUDFLARE_ZONE_ID=
```

## License

MIT

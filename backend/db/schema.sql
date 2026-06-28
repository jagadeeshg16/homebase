CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subdomains (
    id            INTEGER PRIMARY KEY,
    name          TEXT UNIQUE NOT NULL,
    full_domain   TEXT NOT NULL,
    is_public     BOOLEAN DEFAULT 1,
    password_hash TEXT,
    rate_limit    INTEGER DEFAULT 100,
    is_active     BOOLEAN DEFAULT 1,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS health_logs (
    id           INTEGER PRIMARY KEY,
    subdomain_id INTEGER REFERENCES subdomains(id),
    status       INTEGER,
    response_ms  INTEGER,
    checked_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS dns_log (
    id        INTEGER PRIMARY KEY,
    old_ip    TEXT,
    new_ip    TEXT,
    provider  TEXT,
    success   BOOLEAN,
    logged_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

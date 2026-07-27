-- +goose Up
-- Foam scaffold baseline schema (SQLite).
-- No physical foreign keys; use explicit indexes (Wavelet convention).

CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_admins_username CHECK (length(trim(username)) BETWEEN 1 AND 100)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_admins_username ON admins (username);

CREATE TABLE IF NOT EXISTS admin_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id INTEGER NOT NULL,
    refresh_token_hash VARCHAR(64) NOT NULL,
    expires_at DATETIME NOT NULL,
    last_used_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_admin_sessions_token_hash CHECK (length(refresh_token_hash) = 64)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_sessions_refresh_token_hash ON admin_sessions (refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_admin_created ON admin_sessions (admin_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires ON admin_sessions (expires_at);

CREATE TABLE IF NOT EXISTS examples (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    description VARCHAR(1024) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_examples_name CHECK (length(trim(name)) BETWEEN 1 AND 160)
);
CREATE INDEX IF NOT EXISTS idx_examples_created_id ON examples (created_at DESC, id);
CREATE INDEX IF NOT EXISTS idx_examples_name ON examples (name);

CREATE TABLE IF NOT EXISTS runtime_settings (
    key VARCHAR(64) PRIMARY KEY,
    value_json TEXT NOT NULL,
    revision INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_runtime_settings_key CHECK (length(trim(key)) BETWEEN 1 AND 64),
    CONSTRAINT chk_runtime_settings_json_length CHECK (length(value_json) <= 1048576)
);

-- +goose Down
DROP TABLE IF EXISTS runtime_settings;
DROP TABLE IF EXISTS examples;
DROP TABLE IF EXISTS admin_sessions;
DROP TABLE IF EXISTS admins;

-- Mortar SQLite schema
--
-- This file is the single source of truth for the Mortar database schema.
-- Migrations are applied by internal/db/db.go at server startup by executing
-- this file when the database is first created. Future schema changes should
-- be applied via sequential migration files (not by editing this file in place).

-- ---------------------------------------------------------------------------
-- Users
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS users (
    id           TEXT    NOT NULL PRIMARY KEY,  -- UUID
    username     TEXT    NOT NULL UNIQUE,
    role         TEXT    NOT NULL CHECK (role IN ('admin', 'user')),
    password_hash TEXT   NOT NULL,
    created_at   TEXT    NOT NULL               -- ISO 8601
);

-- ---------------------------------------------------------------------------
-- Sessions
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT NOT NULL PRIMARY KEY,  -- UUID
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL               -- ISO 8601
);

CREATE INDEX IF NOT EXISTS idx_sessions_token   ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);

-- ---------------------------------------------------------------------------
-- External account links
-- Stores per-user, per-plugin upstream identity associations.
-- These are required for features like Browse & Play that need personalised
-- upstream context (see ADR 0005).
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS external_account_links (
    id               TEXT NOT NULL PRIMARY KEY,  -- UUID
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plugin_id         TEXT NOT NULL,
    upstream_user_id  TEXT NOT NULL,
    upstream_username TEXT,                      -- optional display name from upstream
    created_at        TEXT NOT NULL,             -- ISO 8601

    UNIQUE (user_id, plugin_id)
);

CREATE INDEX IF NOT EXISTS idx_ext_links_user_plugin ON external_account_links(user_id, plugin_id);

-- ---------------------------------------------------------------------------
-- Health snapshots
-- Stores the last-known health state per plugin. Only the most recent
-- snapshot per plugin is retained in v1; history is out of scope.
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS health_snapshots (
    plugin_id  TEXT NOT NULL PRIMARY KEY,
    status     TEXT NOT NULL CHECK (status IN ('unknown', 'healthy', 'degraded', 'unreachable')),
    checked_at TEXT NOT NULL,   -- ISO 8601
    detail     TEXT             -- error message, if any
);

-- ---------------------------------------------------------------------------
-- Request snapshots
-- A durable convenience index for request history and duplicate prevention.
-- The upstream request system remains authoritative; this is a local cache.
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS request_snapshots (
    id                  TEXT NOT NULL PRIMARY KEY,  -- UUID
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plugin_id           TEXT NOT NULL,
    upstream_request_id TEXT NOT NULL,
    media_item_id       TEXT NOT NULL,              -- Mortar internal ID (plugin:externalId)
    status              TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'available', 'declined', 'failed')),
    created_at          TEXT NOT NULL,              -- ISO 8601
    updated_at          TEXT NOT NULL,              -- ISO 8601
    fulfilled_at        TEXT                        -- ISO 8601; NULL until fulfilled
);

CREATE INDEX IF NOT EXISTS idx_request_snapshots_user_id  ON request_snapshots(user_id);
CREATE INDEX IF NOT EXISTS idx_request_snapshots_plugin   ON request_snapshots(plugin_id);
CREATE INDEX IF NOT EXISTS idx_request_snapshots_upstream ON request_snapshots(plugin_id, upstream_request_id);

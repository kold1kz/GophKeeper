-- =========================
-- EXTENSIONS
-- =========================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =========================
-- USERS
-- =========================
CREATE TABLE IF NOT EXISTS users (
                                     id              BIGSERIAL PRIMARY KEY,
                                     login           TEXT NOT NULL UNIQUE,
                                     password_hash   TEXT NOT NULL,
                                     created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                     updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_login ON users(login);

-- =========================
-- USER SESSIONS
-- =========================
CREATE TABLE IF NOT EXISTS user_sessions (
                                             id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                             user_id              BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                             refresh_token_hash   TEXT NOT NULL UNIQUE,
                                             device_id            TEXT,
                                             device_name          TEXT,
                                             user_agent           TEXT,
                                             ip                   INET,
                                             expires_at           TIMESTAMPTZ NOT NULL,
                                             created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                             last_used_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                             revoked_at           TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id
    ON user_sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at
    ON user_sessions(expires_at);

-- =========================
-- VAULT ITEMS
-- =========================
CREATE TABLE IF NOT EXISTS vault_items (
                           id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           user_id              BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                           type                 TEXT NOT NULL CHECK (
                               type IN ('login_password', 'text', 'binary', 'bank_card')
                               ),

                           title                TEXT NOT NULL,
                           meta                 TEXT,

                           payload_encrypted    BYTEA NOT NULL,
                           payload_nonce        BYTEA,
                           payload_hash         TEXT,

                           version              BIGINT NOT NULL DEFAULT 1,

                           client_updated_at    TIMESTAMPTZ,
                           created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           deleted_at           TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_vault_items_user_id
    ON vault_items(user_id);

CREATE INDEX IF NOT EXISTS idx_vault_items_user_id_updated_at
    ON vault_items(user_id, updated_at);

CREATE INDEX IF NOT EXISTS idx_vault_items_user_id_deleted_at
    ON vault_items(user_id, deleted_at);
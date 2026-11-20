-- FlexGram / NovaMessenger - Initial schema
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(32)  NOT NULL UNIQUE,
    email           VARCHAR(255) UNIQUE,
    password_hash   TEXT         NOT NULL,
    display_name    VARCHAR(64)  NOT NULL DEFAULT '',
    bio             VARCHAR(300) NOT NULL DEFAULT '',
    avatar_url      TEXT         NOT NULL DEFAULT '',
    status_emoji    VARCHAR(16)  NOT NULL DEFAULT '',
    profile_color   VARCHAR(32)  NOT NULL DEFAULT '#2AABEE',
    profile_pattern VARCHAR(32)  NOT NULL DEFAULT 'none',
    is_verified     BOOLEAN      NOT NULL DEFAULT FALSE,
    is_premium      BOOLEAN      NOT NULL DEFAULT FALSE,
    is_online       BOOLEAN      NOT NULL DEFAULT FALSE,
    last_seen       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_email    ON users (email);

CREATE TABLE IF NOT EXISTS wallets (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chats (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type        VARCHAR(20) NOT NULL DEFAULT 'direct'
                    CHECK (type IN ('direct', 'group', 'channel')),
    title       VARCHAR(128),
    description VARCHAR(500),
    avatar_url  TEXT,
    created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chat_members (
    chat_id      UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role         VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_read_at TIMESTAMPTZ,
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_members_user ON chat_members(user_id);

CREATE TABLE IF NOT EXISTS messages (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id      UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    content      TEXT NOT NULL DEFAULT '',
    message_type VARCHAR(20) NOT NULL DEFAULT 'text',
    reply_to     UUID REFERENCES messages(id) ON DELETE SET NULL,
    edited_at    TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON messages(chat_id, created_at DESC);

CREATE TABLE IF NOT EXISTS message_receipts (
    message_id   UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    delivered_at TIMESTAMPTZ,
    read_at      TIMESTAMPTZ,
    PRIMARY KEY (message_id, user_id)
);

CREATE TABLE IF NOT EXISTS gifts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(80)  NOT NULL,
    description VARCHAR(300) NOT NULL DEFAULT '',
    price       BIGINT       NOT NULL DEFAULT 0 CHECK (price >= 0),
    rarity      VARCHAR(24)  NOT NULL DEFAULT 'common',
    image_url   TEXT         NOT NULL DEFAULT '',
    background  VARCHAR(64)  NOT NULL DEFAULT '',
    pattern     VARCHAR(64)  NOT NULL DEFAULT '',
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_gifts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gift_id       UUID NOT NULL REFERENCES gifts(id),
    owner_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    serial_number BIGINT,
    is_equipped   BOOLEAN NOT NULL DEFAULT FALSE,
    acquired_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_gifts_owner ON user_gifts(owner_id);

-- Seed some gifts
INSERT INTO gifts (name, description, price, rarity, background, pattern)
VALUES
    ('Golden Star', 'Classic collectible star', 50, 'rare', 'gold', 'stars'),
    ('Crystal Heart', 'Shiny crystal heart', 120, 'epic', 'crystal', 'hearts'),
    ('Neon Rocket', 'Limited neon rocket', 300, 'legendary', 'neon', 'rockets'),
    ('Pixel Cat', 'Cute pixel cat', 80, 'rare', 'pixel', 'cats'),
    ('Diamond Crown', 'Ultimate crown', 1000, 'mythic', 'diamond', 'crowns')
ON CONFLICT DO NOTHING;

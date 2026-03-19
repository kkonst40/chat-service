-- schema.sql

-- ===== Роли пользователей =====
CREATE TYPE user_role AS ENUM ('common', 'admin', 'owner');

-- ===== Чаты =====
CREATE TABLE chats (
    id              UUID PRIMARY KEY,
    name            TEXT NOT NULL,
    is_group        BOOLEAN NOT NULL,
    last_message_at TIMESTAMP NOT NULL DEFAULT now()
);

-- ===== Пользователи в чатах =====
CREATE TABLE users (
    id      UUID NOT NULL,
    chat_id UUID NOT NULL,
    role    user_role NOT NULL DEFAULT 'common',

    PRIMARY KEY (id, chat_id),

    CONSTRAINT fk_users_chat
        FOREIGN KEY (chat_id)
        REFERENCES chats(id)
        ON DELETE CASCADE
);

-- ===== Сообщения =====
CREATE TABLE messages (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    chat_id    UUID NOT NULL,
    text       TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT fk_messages_user
        FOREIGN KEY (user_id, chat_id)
        REFERENCES users(id, chat_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_messages_chat
        FOREIGN KEY (chat_id)
        REFERENCES chats(id)
        ON DELETE CASCADE
);

-- ===== Индексы =====
CREATE INDEX idx_users_chat_id ON users(chat_id);

CREATE INDEX idx_messages_chat_id ON messages(chat_id);
CREATE INDEX idx_messages_user_chat ON messages(user_id, chat_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

CREATE INDEX idx_chats_last_action ON chats(last_message_at DESC);
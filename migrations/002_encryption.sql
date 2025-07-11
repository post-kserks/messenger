-- Миграция для добавления шифрования сообщений
-- Выполняется после 001_init.sql

-- Таблица для хранения публичных ключей пользователей
CREATE TABLE user_keys (
    user_id INTEGER PRIMARY KEY,
    public_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Таблица для хранения зашифрованных сообщений
CREATE TABLE encrypted_messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER,
    sender_id INTEGER,
    encrypted_data TEXT NOT NULL,
    nonce TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Индексы для оптимизации
CREATE INDEX idx_encrypted_messages_chat_id ON encrypted_messages(chat_id);
CREATE INDEX idx_encrypted_messages_sent_at ON encrypted_messages(sent_at);
CREATE INDEX idx_encrypted_messages_sender_id ON encrypted_messages(sender_id);

-- Добавляем поля для шифрования в существующую таблицу messages
ALTER TABLE messages ADD COLUMN is_encrypted BOOLEAN DEFAULT FALSE;
ALTER TABLE messages ADD COLUMN encrypted_data TEXT;
ALTER TABLE messages ADD COLUMN nonce TEXT;

-- Индексы для новых полей
CREATE INDEX idx_messages_is_encrypted ON messages(is_encrypted);
CREATE INDEX idx_messages_encrypted_data ON messages(encrypted_data) WHERE encrypted_data IS NOT NULL;

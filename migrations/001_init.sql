CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);

CREATE TABLE contacts (
    user_id INTEGER,
    contact_id INTEGER,
    PRIMARY KEY (user_id, contact_id)
);

CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    name TEXT,
    is_group BOOLEAN DEFAULT FALSE
);

CREATE TABLE chat_members (
    chat_id INTEGER,
    user_id INTEGER,
    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER,
    sender_id INTEGER,
    text TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chat_reads (
    user_id INTEGER,
    chat_id INTEGER,
    last_read TIMESTAMP,
    PRIMARY KEY (user_id, chat_id)
);

CREATE TABLE message_reactions (
    message_id INTEGER,
    user_id INTEGER,
    emoji TEXT NOT NULL,
    reacted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (message_id, user_id)
);

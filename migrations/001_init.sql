-- users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);

-- contacts
CREATE TABLE contacts (
    user_id INT,
    contact_id INT,
    PRIMARY KEY (user_id, contact_id)
);

-- chats
CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    name TEXT,
    is_group BOOLEAN DEFAULT FALSE
);

-- chat_members
CREATE TABLE chat_members (
    chat_id INT,
    user_id INT,
    PRIMARY KEY (chat_id, user_id)
);

-- messages
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    chat_id INT,
    sender_id INT,
    text TEXT,
    sent_at TIMESTAMP DEFAULT now()
);

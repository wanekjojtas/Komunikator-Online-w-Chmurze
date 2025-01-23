--Create the `users` table
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
    username VARCHAR(255) NOT NULL UNIQUE,         
    email VARCHAR(255) NOT NULL UNIQUE,            
    password TEXT NOT NULL,                        
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

-- Create the `chats` table
CREATE TABLE IF NOT EXISTS chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Chat ID
    name VARCHAR(255),                             -- Chat name
    creator_id UUID REFERENCES users(id),          -- ID of the user who created the chat
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Chat creation timestamp
);

-- Create the `chat_members` table
CREATE TABLE IF NOT EXISTS chat_members (
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE, -- Chat ID
    user_id UUID REFERENCES users(id),                   -- User ID
    PRIMARY KEY (chat_id, user_id)                       -- Composite primary key
);

-- Create the `messages` table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),       -- Message ID
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE, -- Chat ID
    sender_id UUID REFERENCES users(id),                 -- User ID of the sender
    content TEXT NOT NULL,                               -- Message content
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP       -- Timestamp of the message
);
--dummy
-- Create the `chats` table
CREATE TABLE IF NOT EXISTS chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Chat ID
    name VARCHAR(255),                             -- Chat name
    creator_id UUID REFERENCES users(id),          -- ID of the user who created the chat
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Chat creation timestamp
);

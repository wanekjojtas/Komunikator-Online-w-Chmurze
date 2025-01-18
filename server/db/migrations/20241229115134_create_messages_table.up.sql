-- Create the `messages` table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),       -- Message ID
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE, -- Chat ID
    sender_id UUID REFERENCES users(id),                 -- User ID of the sender
    content TEXT NOT NULL,                               -- Message content
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP       -- Timestamp of the message
);

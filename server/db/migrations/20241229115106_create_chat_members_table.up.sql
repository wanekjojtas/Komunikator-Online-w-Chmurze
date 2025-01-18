-- Create the `chat_members` table
CREATE TABLE IF NOT EXISTS chat_members (
    chat_id UUID REFERENCES chats(id) ON DELETE CASCADE, -- Chat ID
    user_id UUID REFERENCES users(id),                   -- User ID
    PRIMARY KEY (chat_id, user_id)                       -- Composite primary key
);

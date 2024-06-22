CREATE TABLE IF NOT EXISTS messages (
    id UUID NOT NULL,
    sender_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    is_sent BOOLEAN DEFAULT true,
    message TEXT NOT NULL,
    message_type TEXT NOT NULL,
    drop_code TEXT NOT NULL,
    drop_message TEXT NOT NULL,
    reply_parent_message_id TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW()
);


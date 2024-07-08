CREATE TABLE IF NOT EXISTS activities (
    id UUID NOT NULL,
    sender_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    type TEXT,
    message TEXT NOT NULL DEFAULT false,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);


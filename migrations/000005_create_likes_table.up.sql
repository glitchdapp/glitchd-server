CREATE TABLE IF NOT EXISTS likes (
    id UUID NOT NULL,
    user_id TEXT NOT NULL,
    post_id TEXT NOT NULL,
    updated_at timestamp NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS followers (
    id UUID NOT NULL,
    user_id TEXT NOT NULL,
    follower_id TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, follower_id)
);


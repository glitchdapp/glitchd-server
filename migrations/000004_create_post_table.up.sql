CREATE TABLE IF NOT EXISTS posts (
    id UUID NOT NULL,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    caption TEXT NOT NULL,
    is_premium BOOLEAN NOT NULL DEFAULT FALSE,
    is_visible BOOLEAN NOT NULL DEFAULT FALSE,
    thumbnail TEXT NOT NULL,
    type TEXT NOT NULL,
    media TEXT NOT NULL,
    updated_at timestamp NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
);


CREATE TABLE IF NOT EXISTS channels (
    id UUID NOT NULL,
    user_id TEXT,
    title TEXT,
    category TEXT,
    streamkey TEXT,
    playback_id TEXT,
    tags TEXT[],
    is_branded BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);



CREATE TABLE IF NOT EXISTS channels (
    id UUID NOT NULL,
    user_id TEXT UNIQUE,
    title TEXT,
    notification TEXT,
    category TEXT,
    livestream_id TEXT,
    streamkey TEXT,
    playback_id TEXT,
    tags TEXT,
    is_branded BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);



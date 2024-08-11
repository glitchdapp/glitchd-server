CREATE TABLE IF NOT EXISTS videos (
    id UUID NOT NULL,
    channel_id TEXT NOT NULL,
    title TEXT NOT NULL,
    caption TEXT,
    category TEXT,
    thumbnail TEXT,
    poster TEXT,
    media TEXT,
    job_id TEXT,
    tier INTEGER NOT NULL DEFAULT 1,
    is_premium BOOLEAN NOT NULL DEFAULT TRUE,
    is_visible BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at timestamp NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS video_views (
    id UUID NOT NULL,
    channel_id TEXT NOT NULL,
    video_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS channel_viewers (
    id UUID NOT NULL,
    channel_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS video_jobs (
    id UUID NOT NULL,
    job_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS movie_night (
    id UUID NOT NULL,
    name TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    is_private BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at timestamp NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS movie_night_members (
    id UUID NOT NULL,
    movie_night_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS movie_night_messages (
    id UUID NOT NULL,
    movie_night_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    message TEXT NOT NULL,
    media TEXT,
    created_at timestamp NOT NULL DEFAULT NOW()
);

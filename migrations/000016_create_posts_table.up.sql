CREATE TABLE IF NOT EXISTS posts (
    id UUID NOT NULL,
    author TEXT NOT NULL,
    message TEXT NOT NULL,
    media TEXT,
    reply_to TEXT,
    created_at timestamp NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS reposts (
    id UUID NOT NULL,
    user_id TEXT NOT NULL,
    post_id TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS likes (
    id UUID NOT NULL,
    user_id TEXT NOT NULL,
    post_id TEXT NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

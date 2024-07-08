CREATE TABLE IF NOT EXISTS chat_identities (
    id UUID NOT NULL,
    user_id TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#be123c',
    badge TEXT NOT NULL,
    UNIQUE (user_id)
);


CREATE TABLE IF NOT EXISTS apps (
    id UUID NOT NULL,
    owner_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    vanity VARCHAR(255) NOT NULL,
    favicon TEXT NOT NULL,
    logo TEXT NOT NULL
);

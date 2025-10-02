CREATE TABLE IF NOT EXISTS support_requests (
    id UUID NOT NULL,
    email TEXT NOT NULL,
    message TEXT NOT NULL,
    image TEXT,
    resolved BOOLEAN NOT NULL DEFAULT false,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL DEFAULT NOW()
);

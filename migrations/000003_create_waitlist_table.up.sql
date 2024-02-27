CREATE TABLE IF NOT EXISTS waitlists (
    id UUID NOT NULL,
    email TEXT NOT NULL,
    can_enter BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp NOT NULL DEFAULT NOW() 
);


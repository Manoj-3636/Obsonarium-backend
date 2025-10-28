CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email citext UNIQUE NOT NULL,
    name TEXT NOT NULL,                
    profile_picture_url TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW()
);
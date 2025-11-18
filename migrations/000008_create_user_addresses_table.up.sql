CREATE TABLE user_addresses (
    id SERIAL PRIMARY KEY,
    
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    label TEXT, -- e.g., "Home", "Work", "Office"
    street_address TEXT NOT NULL,
    city TEXT NOT NULL,
    state TEXT,
    postal_code TEXT NOT NULL,
    country TEXT NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);



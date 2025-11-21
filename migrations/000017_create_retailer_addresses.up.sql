CREATE TABLE retailer_addresses (
    id SERIAL PRIMARY KEY,
    retailer_id INT NOT NULL REFERENCES retailers(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    street_address TEXT NOT NULL,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    country TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_retailer_addresses_retailer_id ON retailer_addresses(retailer_id);

CREATE TABLE retailer_products (
    id SERIAL PRIMARY KEY,
    retailer_id INT NOT NULL REFERENCES retailers(id),

    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    stock_qty INT DEFAULT 0,
    image_url TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

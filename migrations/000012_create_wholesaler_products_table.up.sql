CREATE TABLE wholesaler_products (
    id SERIAL PRIMARY KEY,
    wholesaler_id INT NOT NULL REFERENCES wholesalers(id),

    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    stock_qty INT DEFAULT 0,
    image_url TEXT,
    description TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);


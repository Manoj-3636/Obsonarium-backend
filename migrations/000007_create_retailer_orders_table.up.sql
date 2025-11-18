CREATE TABLE retailer_orders (
    id SERIAL PRIMARY KEY,
    
    retailer_id INT NOT NULL REFERENCES retailers(id),
    product_id INT NOT NULL REFERENCES retailer_products(id),
    user_id INT NOT NULL REFERENCES users(id),
    
    quantity INT NOT NULL DEFAULT 1,
    total_price NUMERIC(10,2) NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);


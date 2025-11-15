CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    
    user_id INT NOT NULL REFERENCES users(id),
    product_id INT NOT NULL REFERENCES retailer_products(id),

    quantity INT NOT NULL DEFAULT 1,

    UNIQUE (user_id, product_id)
);

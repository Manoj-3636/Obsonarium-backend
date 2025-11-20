CREATE TABLE retailer_cart_items (
    id SERIAL PRIMARY KEY,
    
    retailer_id INT NOT NULL REFERENCES retailers(id),
    product_id INT NOT NULL REFERENCES wholesaler_products(id),

    quantity INT NOT NULL DEFAULT 1,

    UNIQUE (retailer_id, product_id)
);


CREATE TABLE product_queries (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES retailer_products(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    query_text TEXT NOT NULL,
    response_text TEXT,
    is_resolved BOOLEAN DEFAULT FALSE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_product_queries_product_id ON product_queries(product_id);
CREATE INDEX idx_product_queries_user_id ON product_queries(user_id);
CREATE INDEX idx_product_queries_is_resolved ON product_queries(is_resolved);


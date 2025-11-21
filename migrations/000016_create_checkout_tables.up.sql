-- Refactor retailer_orders to be the header
ALTER TABLE retailer_orders DROP COLUMN product_id;
ALTER TABLE retailer_orders DROP COLUMN quantity;
ALTER TABLE retailer_orders ADD COLUMN stripe_session_id TEXT;
ALTER TABLE retailer_orders ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();

-- Create retailer_order_items
CREATE TABLE retailer_order_items (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES retailer_orders(id),
    product_id INT NOT NULL REFERENCES retailer_products(id),
    quantity INT NOT NULL,
    price NUMERIC(10,2) NOT NULL
);

-- Create wholesaler_orders
CREATE TABLE wholesaler_orders (
    id SERIAL PRIMARY KEY,
    wholesaler_id INT NOT NULL REFERENCES wholesalers(id),
    retailer_id INT NOT NULL REFERENCES retailers(id),
    total_price NUMERIC(10,2) NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    stripe_session_id TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create wholesaler_order_items
CREATE TABLE wholesaler_order_items (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES wholesaler_orders(id),
    product_id INT NOT NULL REFERENCES wholesaler_products(id),
    quantity INT NOT NULL,
    price NUMERIC(10,2) NOT NULL
);

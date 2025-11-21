DROP TABLE wholesaler_order_items;
DROP TABLE wholesaler_orders;
DROP TABLE retailer_order_items;

ALTER TABLE retailer_orders DROP COLUMN updated_at;
ALTER TABLE retailer_orders DROP COLUMN stripe_session_id;
ALTER TABLE retailer_orders ADD COLUMN quantity INT NOT NULL DEFAULT 1;
ALTER TABLE retailer_orders ADD COLUMN product_id INT NOT NULL REFERENCES retailer_products(id);

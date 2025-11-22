ALTER TABLE consumer_orders ADD COLUMN retailer_id INT NOT NULL REFERENCES retailers(id) ON DELETE CASCADE;

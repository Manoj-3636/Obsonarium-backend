ALTER TABLE retailer_orders
ADD COLUMN address_id INT NOT NULL REFERENCES user_addresses(id);



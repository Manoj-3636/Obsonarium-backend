ALTER TABLE retailer_orders
DROP CONSTRAINT retailer_orders_address_id_fkey,
ADD CONSTRAINT retailer_orders_address_id_fkey 
    FOREIGN KEY (address_id) REFERENCES user_addresses(id);

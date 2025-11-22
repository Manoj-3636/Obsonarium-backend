CREATE TABLE retailer_wholesale_order_items (
  id SERIAL PRIMARY KEY,
  order_id INT REFERENCES retailer_wholesale_orders(id) ON DELETE CASCADE,
  product_id INT NOT NULL REFERENCES wholesaler_products(id),
  qty INT NOT NULL,
  unit_price NUMERIC NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending'
);


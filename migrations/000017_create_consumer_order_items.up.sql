CREATE TABLE consumer_order_items (
  id SERIAL PRIMARY KEY,
  order_id INT REFERENCES consumer_orders(id) ON DELETE CASCADE,
  product_id INT NOT NULL REFERENCES retailer_products(id),
  qty INT NOT NULL,
  unit_price NUMERIC NOT NULL
);

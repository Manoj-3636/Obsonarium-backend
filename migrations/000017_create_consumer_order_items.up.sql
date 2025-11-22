CREATE TABLE consumer_order_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID REFERENCES consumer_orders(id) ON DELETE CASCADE,
  product_id UUID NOT NULL,
  qty INT NOT NULL,
  unit_price NUMERIC NOT NULL
);

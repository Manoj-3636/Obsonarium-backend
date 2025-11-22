CREATE TABLE consumer_orders (
  id SERIAL PRIMARY KEY,
  consumer_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  retailer_id INT NOT NULL REFERENCES retailers(id) ON DELETE CASCADE,
  payment_method TEXT NOT NULL, -- 'offline' (or 'online' later)
  payment_status TEXT NOT NULL DEFAULT 'pending', -- pending/paid/failed/refunded
  order_status TEXT NOT NULL DEFAULT 'placed', -- placed/accepted/rejected/packed/shipped/delivered/cancelled
  total_amount NUMERIC NOT NULL,
  scheduled_at TIMESTAMP NULL, -- for offline scheduling
  address_id INT REFERENCES user_addresses(id) ON DELETE SET NULL,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);
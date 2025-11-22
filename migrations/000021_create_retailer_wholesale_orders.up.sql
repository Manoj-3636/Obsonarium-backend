CREATE TABLE retailer_wholesale_orders (
  id SERIAL PRIMARY KEY,
  retailer_id INT NOT NULL REFERENCES retailers(id) ON DELETE CASCADE,
  payment_method TEXT NOT NULL, -- 'offline' (or 'online' later)
  payment_status TEXT NOT NULL DEFAULT 'pending', -- pending/success/failed/refunded
  order_status TEXT NOT NULL DEFAULT 'placed', -- placed/accepted/rejected/packed/shipped/delivered
  total_amount NUMERIC NOT NULL,
  scheduled_at TIMESTAMP NULL, -- for offline scheduling
  stripe_session_id TEXT,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);


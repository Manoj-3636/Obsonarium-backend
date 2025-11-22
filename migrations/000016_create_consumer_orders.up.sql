CREATE TABLE consumer_orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  consumer_id UUID NOT NULL,
  retailer_id UUID NOT NULL,
  payment_method TEXT NOT NULL, -- 'offline' (or 'online' later)
  payment_status TEXT NOT NULL DEFAULT 'pending', -- pending/paid/failed/refunded
  order_status TEXT NOT NULL DEFAULT 'placed', -- placed/accepted/rejected/packed/shipped/delivered/cancelled
  total_amount NUMERIC NOT NULL,
  scheduled_at TIMESTAMP NULL, -- for offline scheduling
  address_id UUID,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);
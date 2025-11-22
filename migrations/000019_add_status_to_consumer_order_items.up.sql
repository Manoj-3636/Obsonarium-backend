ALTER TABLE consumer_order_items 
ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';

COMMENT ON COLUMN consumer_order_items.status IS 'Status of the order item: pending/accepted/rejected/shipped/delivered/cancelled';


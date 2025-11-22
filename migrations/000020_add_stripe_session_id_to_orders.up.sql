ALTER TABLE consumer_orders 
ADD COLUMN stripe_session_id TEXT;

COMMENT ON COLUMN consumer_orders.stripe_session_id IS 'Stripe checkout session ID for online payments';


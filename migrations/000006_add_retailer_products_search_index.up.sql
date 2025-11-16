-- Create a GIN index on the expression used in your full-text search
-- This matches your repo SQL:
-- to_tsvector('simple', name || ' ' || coalesce(description, ''))
CREATE INDEX IF NOT EXISTS retailer_products_search_idx
ON retailer_products
USING GIN (
    to_tsvector(
        'simple',
        name || ' ' || coalesce(description, '')
    )
);

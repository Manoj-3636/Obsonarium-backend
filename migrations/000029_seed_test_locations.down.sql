-- Revert seeded locations (optional, but good practice to have a down migration)
-- We can't easily know which ones were empty before, so we might just leave this empty 
-- or try to revert specific known test values if we wanted to be strict.
-- For now, we will just leave it as a no-op or maybe set them back to empty if they match exactly the seeded value.

UPDATE retailers
SET 
    address = '',
    street_address = '',
    city = '',
    state = '',
    postal_code = '',
    country = '',
    latitude = NULL,
    longitude = NULL
WHERE address = 'Hitech City, Hyderabad, Telangana, 500081, India' 
  AND latitude = 17.4435 
  AND longitude = 78.3772;

UPDATE wholesalers
SET 
    address = '',
    street_address = '',
    city = '',
    state = '',
    postal_code = '',
    country = '',
    latitude = NULL,
    longitude = NULL
WHERE address = 'Gachibowli, Hyderabad, Telangana, 500032, India'
  AND latitude = 17.4401
  AND longitude = 78.3489;

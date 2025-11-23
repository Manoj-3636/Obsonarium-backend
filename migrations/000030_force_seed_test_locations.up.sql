-- Force update all retailers to test location
UPDATE retailers
SET 
    address = 'Hitech City, Hyderabad, Telangana, 500081, India',
    street_address = 'Hitech City Main Rd',
    city = 'Hyderabad',
    state = 'Telangana',
    postal_code = '500081',
    country = 'India',
    latitude = 17.4435,
    longitude = 78.3772;

-- Force update all wholesalers to test location
UPDATE wholesalers
SET 
    address = 'Gachibowli, Hyderabad, Telangana, 500032, India',
    street_address = 'Gachibowli - Miyapur Rd',
    city = 'Hyderabad',
    state = 'Telangana',
    postal_code = '500032',
    country = 'India',
    latitude = 17.4401,
    longitude = 78.3489;

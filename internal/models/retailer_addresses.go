package models

type RetailerAddress struct {
	Id             int
	Retailer_id    int
	Label          string
	Street_address string
	City           string
	State          string
	Postal_code    string
	Country        string
}

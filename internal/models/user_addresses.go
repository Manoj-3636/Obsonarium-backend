package models

type UserAddress struct {
	Id             int
	User_id        int
	Label          string
	Street_address string
	City           string
	State          string
	Postal_code    string
	Country        string
}

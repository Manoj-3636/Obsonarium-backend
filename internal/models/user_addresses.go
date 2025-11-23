package models

type UserAddress struct {
	Id             int     `json:"id"`
	User_id        int     `json:"user_id"`
	Label          string  `json:"label"`
	Street_address string  `json:"street_address"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Postal_code    string  `json:"postal_code"`
	Country        string  `json:"country"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
}

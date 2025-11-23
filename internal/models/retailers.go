package models

type Retailer struct {
	Id            int      `json:"id"`
	Name          string   `json:"name"`
	BusinessName  string   `json:"business_name"`
	Email         string   `json:"email"`
	Phone         string   `json:"phone"`
	Address       string   `json:"address"`
	StreetAddress string   `json:"street_address"`
	City          string   `json:"city"`
	State         string   `json:"state"`
	PostalCode    string   `json:"postal_code"`
	Country       string   `json:"country"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
}

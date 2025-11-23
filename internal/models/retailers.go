package models

type Retailer struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	BusinessName string  `json:"business_name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	Address      string  `json:"address"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

package models

type CartItem struct {
	Id         int             `json:"id"`
	User_id    int             `json:"user_id"`
	Product_id int             `json:"product_id"`
	Quantity   int             `json:"quantity"`
	Product    RetailerProduct `json:"product"`
}

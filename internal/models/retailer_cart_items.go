package models

type RetailerCartItem struct {
	Id          int               `json:"id"`
	Retailer_id int               `json:"retailer_id"`
	Product_id  int               `json:"product_id"`
	Quantity    int               `json:"quantity"`
	Product     WholesalerProduct `json:"product"`
}

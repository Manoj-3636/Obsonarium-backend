package models

type RetailerProduct struct {
	Id          int     `json:"id"`
	Retailer_id int     `json:"retailer_id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Stock_qty   int     `json:"stock_qty"`
	Image_url   string  `json:"image_url"`
	Description string  `json:"description"`
}

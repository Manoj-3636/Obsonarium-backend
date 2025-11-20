package models

type WholesalerProduct struct {
	Id            int     `json:"id"`
	Wholesaler_id int     `json:"wholesaler_id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Stock_qty     int     `json:"stock_qty"`
	Image_url     string  `json:"image_url"`
	Description   string  `json:"description"`
}

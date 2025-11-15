package models

type RetailerProduct struct {
	Id          int
	Retailer_id int
	Name        string
	Price       float64
	Stock_qty   int
	Image_url   string
	Description string
}

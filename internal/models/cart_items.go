package models

type CartItem struct {
	Id         int
	User_id    int
	Product_id int
	Quantity   int
	Product    RetailerProduct
}

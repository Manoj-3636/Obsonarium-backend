package models

type ProductReview struct {
	Id         int    `json:"id"`
	Product_id int    `json:"product_id"`
	User_id    int    `json:"user_id"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}


package models

type ProductReview struct {
	Id            int    `json:"id"`
	Product_id    int    `json:"product_id"`
	User_id       int    `json:"user_id"`
	Reviewer_name string `json:"reviewer_name"`
	Rating        int    `json:"rating"`
	Comment       string `json:"comment"`
	Created_at    string `json:"created_at"`
	Updated_at    string `json:"updated_at"`
}


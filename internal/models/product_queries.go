package models

type ProductQuery struct {
	Id          int    `json:"id"`
	Product_id  int    `json:"product_id"`
	User_id     int    `json:"user_id"`
	Query_text  string `json:"query_text"`
	Response_text *string `json:"response_text,omitempty"`
	Is_resolved bool   `json:"is_resolved"`
	Created_at  string `json:"created_at"`
	Updated_at  string `json:"updated_at"`
	Resolved_at *string `json:"resolved_at,omitempty"`
}


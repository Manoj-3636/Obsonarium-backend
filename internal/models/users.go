package models

type User struct {
	Id      int    `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Pfp_url string `json:"pfp_url"`
}

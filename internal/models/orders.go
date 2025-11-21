package models

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
)

type ConsumerOrder struct {
	Id              int                 `json:"id"`
	RetailerId      int                 `json:"retailer_id"`
	UserId          int                 `json:"user_id"`
	AddressId       int                 `json:"address_id"`
	TotalPrice      float64             `json:"total_price"`
	Status          OrderStatus         `json:"status"`
	StripeSessionId string              `json:"stripe_session_id"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
	Items           []ConsumerOrderItem `json:"items,omitempty"`
}

type ConsumerOrderItem struct {
	Id        int              `json:"id"`
	OrderId   int              `json:"order_id"`
	ProductId int              `json:"product_id"`
	Quantity  int              `json:"quantity"`
	Price     float64          `json:"price"`
	Product   *RetailerProduct `json:"product,omitempty"`
}

type RetailerOrder struct {
	Id              int                 `json:"id"`
	WholesalerId    int                 `json:"wholesaler_id"`
	RetailerId      int                 `json:"retailer_id"`
	AddressId       int                 `json:"address_id"`
	TotalPrice      float64             `json:"total_price"`
	Status          OrderStatus         `json:"status"`
	StripeSessionId string              `json:"stripe_session_id"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
	Items           []RetailerOrderItem `json:"items,omitempty"`
}

type RetailerOrderItem struct {
	Id        int                `json:"id"`
	OrderId   int                `json:"order_id"`
	ProductId int                `json:"product_id"`
	Quantity  int                `json:"quantity"`
	Price     float64            `json:"price"`
	Product   *WholesalerProduct `json:"product,omitempty"`
}

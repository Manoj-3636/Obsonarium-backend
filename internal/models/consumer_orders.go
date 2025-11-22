package models

import (
	"time"
)

type ConsumerOrder struct {
	ID            int                 `json:"id" db:"id"`
	ConsumerID    int                 `json:"consumer_id" db:"consumer_id"`
	PaymentMethod string              `json:"payment_method" db:"payment_method"`
	PaymentStatus string              `json:"payment_status" db:"payment_status"`
	OrderStatus   string              `json:"order_status" db:"order_status"`
	TotalAmount   float64             `json:"total_amount" db:"total_amount"`
	ScheduledAt   *time.Time          `json:"scheduled_at" db:"scheduled_at"`
	AddressID     *int                `json:"address_id" db:"address_id"`
	CreatedAt     time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at" db:"updated_at"`
	Items         []ConsumerOrderItem `json:"items,omitempty"`
}

type ConsumerOrderItem struct {
	ID        int     `json:"id" db:"id"`
	OrderID   int     `json:"order_id" db:"order_id"`
	ProductID int     `json:"product_id" db:"product_id"`
	Qty       int     `json:"qty" db:"qty"`
	UnitPrice float64 `json:"unit_price" db:"unit_price"`
}

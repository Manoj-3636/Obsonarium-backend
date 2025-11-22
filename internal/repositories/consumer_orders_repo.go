package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"Obsonarium-backend/internal/models"
)

type ConsumerOrdersRepository struct {
	DB *sql.DB
}

func NewConsumerOrdersRepository(db *sql.DB) *ConsumerOrdersRepository {
	return &ConsumerOrdersRepository{DB: db}
}

func (r *ConsumerOrdersRepository) CreateOrder(ctx context.Context, order *models.ConsumerOrder, items []models.ConsumerOrderItem) (*models.ConsumerOrder, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert Order
	query := `
		INSERT INTO consumer_orders (
			consumer_id, payment_method, payment_status, order_status, 
			total_amount, scheduled_at, address_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRowContext(ctx, query,
		order.ConsumerID,
		order.PaymentMethod,
		order.PaymentStatus,
		order.OrderStatus,
		order.TotalAmount,
		order.ScheduledAt,
		order.AddressID,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert Items
	itemQuery := `
		INSERT INTO consumer_order_items (
			order_id, product_id, qty, unit_price
		) VALUES ($1, $2, $3, $4)`

	for _, item := range items {
		_, err := tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Qty, item.UnitPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	order.Items = items
	return order, nil
}

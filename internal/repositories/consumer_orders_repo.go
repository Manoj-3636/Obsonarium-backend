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
			total_amount, scheduled_at, address_id, stripe_session_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	var stripeSessionID interface{}
	if order.StripeSessionID != nil {
		stripeSessionID = *order.StripeSessionID
	}

	err = tx.QueryRowContext(ctx, query,
		order.ConsumerID,
		order.PaymentMethod,
		order.PaymentStatus,
		order.OrderStatus,
		order.TotalAmount,
		order.ScheduledAt,
		order.AddressID,
		stripeSessionID,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert Items
	itemQuery := `
		INSERT INTO consumer_order_items (
			order_id, product_id, qty, unit_price, status
		) VALUES ($1, $2, $3, $4, $5)`

	for _, item := range items {
		status := item.Status
		if status == "" {
			status = "pending" // Default status
		}
		_, err := tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Qty, item.UnitPrice, status)
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

// GetActiveOrdersByRetailerID gets orders that contain products from the retailer (excludes delivered items)
func (r *ConsumerOrdersRepository) GetActiveOrdersByRetailerID(retailerID int) ([]models.ConsumerOrder, error) {
	query := `
		SELECT DISTINCT
			co.id, co.consumer_id, co.payment_method, co.payment_status, 
			co.order_status, co.total_amount, co.scheduled_at, co.address_id,
			co.stripe_session_id, co.created_at, co.updated_at
		FROM consumer_orders co
		INNER JOIN consumer_order_items coi ON coi.order_id = co.id
		INNER JOIN retailer_products rp ON rp.id = coi.product_id
		WHERE rp.retailer_id = $1
		ORDER BY co.created_at DESC
	`

	rows, err := r.DB.Query(query, retailerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.ConsumerOrder
	orderMap := make(map[int]*models.ConsumerOrder)

	for rows.Next() {
		var order models.ConsumerOrder
		var scheduledAt sql.NullTime
		var addressID sql.NullInt64
		var stripeSessionID sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.ConsumerID,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.OrderStatus,
			&order.TotalAmount,
			&scheduledAt,
			&addressID,
			&stripeSessionID,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if scheduledAt.Valid {
			order.ScheduledAt = &scheduledAt.Time
		}
		if addressID.Valid {
			order.AddressID = new(int)
			*order.AddressID = int(addressID.Int64)
		}

		if _, exists := orderMap[order.ID]; !exists {
			orderMap[order.ID] = &order
			orders = append(orders, order)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	// Get order items for each order (excluding rejected and delivered items)
	// Filter out orders that have no items after excluding rejected/delivered ones
	var filteredOrders []models.ConsumerOrder
	for i := range orders {
		items, err := r.GetActiveOrderItemsByOrderID(orders[i].ID, retailerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get items for order %d: %w", orders[i].ID, err)
		}
		// Only include orders that have at least one active item
		if len(items) > 0 {
			orders[i].Items = items
			filteredOrders = append(filteredOrders, orders[i])
		}
	}

	return filteredOrders, nil
}

// GetActiveOrderItemsByOrderID gets active order items for a specific order, filtered by retailer
// Excludes rejected and delivered items (for active orders page)
func (r *ConsumerOrdersRepository) GetActiveOrderItemsByOrderID(orderID int, retailerID int) ([]models.ConsumerOrderItem, error) {
	query := `
		SELECT coi.id, coi.order_id, coi.product_id, coi.qty, coi.unit_price, coi.status
		FROM consumer_order_items coi
		INNER JOIN retailer_products rp ON rp.id = coi.product_id
		WHERE coi.order_id = $1 AND rp.retailer_id = $2 AND coi.status NOT IN ('rejected', 'delivered')
		ORDER BY coi.id
	`

	rows, err := r.DB.Query(query, orderID, retailerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []models.ConsumerOrderItem
	for rows.Next() {
		var item models.ConsumerOrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Qty,
			&item.UnitPrice,
			&item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return items, nil
}

// UpdateOrderItemStatus updates the status of an order item
func (r *ConsumerOrdersRepository) UpdateOrderItemStatus(itemID int, retailerID int, status string) error {
	// Verify the item belongs to a product from this retailer
	query := `
		UPDATE consumer_order_items coi
		SET status = $1
		FROM retailer_products rp
		WHERE coi.id = $2 
			AND coi.product_id = rp.id 
			AND rp.retailer_id = $3
		RETURNING coi.id
	`

	var updatedID int
	err := r.DB.QueryRow(query, status, itemID, retailerID).Scan(&updatedID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("order item not found or does not belong to this retailer")
	}
	if err != nil {
		return fmt.Errorf("failed to update order item status: %w", err)
	}

	return nil
}

// GetOrderItemsByOrderID gets all order items for a specific order, filtered by retailer
// Includes all statuses (for history page)
func (r *ConsumerOrdersRepository) GetOrderItemsByOrderID(orderID int, retailerID int) ([]models.ConsumerOrderItem, error) {
	query := `
		SELECT coi.id, coi.order_id, coi.product_id, coi.qty, coi.unit_price, coi.status
		FROM consumer_order_items coi
		INNER JOIN retailer_products rp ON rp.id = coi.product_id
		WHERE coi.order_id = $1 AND rp.retailer_id = $2
		ORDER BY coi.id
	`

	rows, err := r.DB.Query(query, orderID, retailerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []models.ConsumerOrderItem
	for rows.Next() {
		var item models.ConsumerOrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Qty,
			&item.UnitPrice,
			&item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return items, nil
}

// GetHistoryOrdersByRetailerID gets completed orders (delivered/rejected items) for history page
func (r *ConsumerOrdersRepository) GetHistoryOrdersByRetailerID(retailerID int) ([]models.ConsumerOrder, error) {
	query := `
		SELECT DISTINCT
			co.id, co.consumer_id, co.payment_method, co.payment_status,
			co.order_status, co.total_amount, co.scheduled_at, co.address_id,
			co.stripe_session_id, co.created_at, co.updated_at
		FROM consumer_orders co
		INNER JOIN consumer_order_items coi ON coi.order_id = co.id
		INNER JOIN retailer_products rp ON rp.id = coi.product_id
		WHERE rp.retailer_id = $1 AND coi.status IN ('rejected', 'delivered')
		ORDER BY co.created_at DESC
	`

	rows, err := r.DB.Query(query, retailerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query history orders: %w", err)
	}
	defer rows.Close()

	var orders []models.ConsumerOrder
	orderMap := make(map[int]*models.ConsumerOrder)

	for rows.Next() {
		var order models.ConsumerOrder
		var scheduledAt sql.NullTime
		var addressID sql.NullInt64
		var stripeSessionID sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.ConsumerID,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.OrderStatus,
			&order.TotalAmount,
			&scheduledAt,
			&addressID,
			&stripeSessionID,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if scheduledAt.Valid {
			order.ScheduledAt = &scheduledAt.Time
		}
		if addressID.Valid {
			order.AddressID = new(int)
			*order.AddressID = int(addressID.Int64)
		}
		if stripeSessionID.Valid {
			order.StripeSessionID = &stripeSessionID.String
		}

		if _, exists := orderMap[order.ID]; !exists {
			orderMap[order.ID] = &order
			orders = append(orders, order)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history orders: %w", err)
	}

	// Get all order items for each history order
	for i := range orders {
		items, err := r.GetOrderItemsByOrderID(orders[i].ID, retailerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get items for history order %d: %w", orders[i].ID, err)
		}
		orders[i].Items = items
	}

	return orders, nil
}

// UpdateStripeSessionID updates the Stripe session ID for an order
func (r *ConsumerOrdersRepository) UpdateStripeSessionID(orderID int, sessionID string) error {
	query := `
		UPDATE consumer_orders
		SET stripe_session_id = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.DB.Exec(query, sessionID, orderID)
	if err != nil {
		return fmt.Errorf("failed to update Stripe session ID: %w", err)
	}

	return nil
}

// GetOrdersByConsumerID gets all orders for a consumer (both ongoing and past)
func (r *ConsumerOrdersRepository) GetOrdersByConsumerID(consumerID int) ([]models.ConsumerOrder, error) {
	query := `
		SELECT 
			co.id, co.consumer_id, co.payment_method, co.payment_status, 
			co.order_status, co.total_amount, co.scheduled_at, co.address_id,
			co.stripe_session_id, co.created_at, co.updated_at
		FROM consumer_orders co
		WHERE co.consumer_id = $1
		ORDER BY 
			CASE 
				WHEN co.order_status != 'cancelled' AND EXISTS (
					SELECT 1 FROM consumer_order_items coi 
					WHERE coi.order_id = co.id 
					AND coi.status NOT IN ('delivered', 'rejected')
				) THEN 0
				ELSE 1
			END,
			co.created_at DESC
	`

	rows, err := r.DB.Query(query, consumerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.ConsumerOrder
	orderMap := make(map[int]*models.ConsumerOrder)

	for rows.Next() {
		var order models.ConsumerOrder
		var scheduledAt sql.NullTime
		var addressID sql.NullInt64
		var stripeSessionID sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.ConsumerID,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.OrderStatus,
			&order.TotalAmount,
			&scheduledAt,
			&addressID,
			&stripeSessionID,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if scheduledAt.Valid {
			order.ScheduledAt = &scheduledAt.Time
		}
		if addressID.Valid {
			order.AddressID = new(int)
			*order.AddressID = int(addressID.Int64)
		}
		if stripeSessionID.Valid {
			order.StripeSessionID = &stripeSessionID.String
		}

		if _, exists := orderMap[order.ID]; !exists {
			orderMap[order.ID] = &order
			orders = append(orders, order)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	// Get all order items for each order
	for i := range orders {
		itemsQuery := `
			SELECT coi.id, coi.order_id, coi.product_id, coi.qty, coi.unit_price, coi.status
			FROM consumer_order_items coi
			WHERE coi.order_id = $1
			ORDER BY coi.id
		`
		itemRows, err := r.DB.Query(itemsQuery, orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query order items: %w", err)
		}

		var items []models.ConsumerOrderItem
		for itemRows.Next() {
			var item models.ConsumerOrderItem
			err := itemRows.Scan(
				&item.ID,
				&item.OrderID,
				&item.ProductID,
				&item.Qty,
				&item.UnitPrice,
				&item.Status,
			)
			if err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("failed to scan order item: %w", err)
			}
			items = append(items, item)
		}
		itemRows.Close()

		if err := itemRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating order items: %w", err)
		}

		orders[i].Items = items
	}

	return orders, nil
}

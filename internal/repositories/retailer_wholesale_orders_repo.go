package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"Obsonarium-backend/internal/models"
)

type RetailerWholesaleOrdersRepository struct {
	DB *sql.DB
}

func NewRetailerWholesaleOrdersRepository(db *sql.DB) *RetailerWholesaleOrdersRepository {
	return &RetailerWholesaleOrdersRepository{DB: db}
}

func (r *RetailerWholesaleOrdersRepository) CreateOrder(ctx context.Context, order *models.RetailerWholesaleOrder, items []models.RetailerWholesaleOrderItem) (*models.RetailerWholesaleOrder, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert Order
	query := `
		INSERT INTO retailer_wholesale_orders (
			retailer_id, payment_method, payment_status, order_status, 
			total_amount, scheduled_at, stripe_session_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	var stripeSessionID interface{}
	if order.StripeSessionID != nil {
		stripeSessionID = *order.StripeSessionID
	}

	err = tx.QueryRowContext(ctx, query,
		order.RetailerID,
		order.PaymentMethod,
		order.PaymentStatus,
		order.OrderStatus,
		order.TotalAmount,
		order.ScheduledAt,
		stripeSessionID,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert Items
	itemQuery := `
		INSERT INTO retailer_wholesale_order_items (
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

// GetActiveOrdersByWholesalerID gets orders that contain products from the wholesaler (excludes delivered items)
func (r *RetailerWholesaleOrdersRepository) GetActiveOrdersByWholesalerID(wholesalerID int) ([]models.RetailerWholesaleOrder, error) {
	query := `
		SELECT DISTINCT
			rwo.id, rwo.retailer_id, rwo.payment_method, rwo.payment_status, 
			rwo.order_status, rwo.total_amount, rwo.scheduled_at,
			rwo.stripe_session_id, rwo.created_at, rwo.updated_at
		FROM retailer_wholesale_orders rwo
		INNER JOIN retailer_wholesale_order_items rwoi ON rwoi.order_id = rwo.id
		INNER JOIN wholesaler_products wp ON wp.id = rwoi.product_id
		WHERE wp.wholesaler_id = $1
		ORDER BY rwo.created_at DESC
	`

	rows, err := r.DB.Query(query, wholesalerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.RetailerWholesaleOrder
	orderMap := make(map[int]*models.RetailerWholesaleOrder)

	for rows.Next() {
		var order models.RetailerWholesaleOrder
		var scheduledAt sql.NullTime
		var stripeSessionID sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.RetailerID,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.OrderStatus,
			&order.TotalAmount,
			&scheduledAt,
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

	// Get order items for each order (excluding rejected and delivered items)
	// Filter out orders that have no items after excluding rejected/delivered ones
	var filteredOrders []models.RetailerWholesaleOrder
	for i := range orders {
		items, err := r.GetActiveOrderItemsByOrderID(orders[i].ID, wholesalerID)
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

// GetActiveOrderItemsByOrderID gets active order items for a specific order, filtered by wholesaler
// Excludes rejected and delivered items (for active orders page)
func (r *RetailerWholesaleOrdersRepository) GetActiveOrderItemsByOrderID(orderID int, wholesalerID int) ([]models.RetailerWholesaleOrderItem, error) {
	query := `
		SELECT rwoi.id, rwoi.order_id, rwoi.product_id, rwoi.qty, rwoi.unit_price, rwoi.status
		FROM retailer_wholesale_order_items rwoi
		INNER JOIN wholesaler_products wp ON wp.id = rwoi.product_id
		WHERE rwoi.order_id = $1 AND wp.wholesaler_id = $2 AND rwoi.status NOT IN ('rejected', 'delivered')
		ORDER BY rwoi.id
	`

	rows, err := r.DB.Query(query, orderID, wholesalerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []models.RetailerWholesaleOrderItem
	for rows.Next() {
		var item models.RetailerWholesaleOrderItem
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
func (r *RetailerWholesaleOrdersRepository) UpdateOrderItemStatus(itemID int, wholesalerID int, status string) error {
	// Verify the item belongs to a product from this wholesaler
	query := `
		UPDATE retailer_wholesale_order_items rwoi
		SET status = $1
		FROM wholesaler_products wp
		WHERE rwoi.id = $2 
			AND rwoi.product_id = wp.id 
			AND wp.wholesaler_id = $3
		RETURNING rwoi.id
	`

	var updatedID int
	err := r.DB.QueryRow(query, status, itemID, wholesalerID).Scan(&updatedID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("order item not found or does not belong to this wholesaler")
	}
	if err != nil {
		return fmt.Errorf("failed to update order item status: %w", err)
	}

	// If marking as delivered for offline payment, update payment status
	if status == "delivered" {
		// Check if all items in the order are delivered
		checkQuery := `
			SELECT COUNT(*) 
			FROM retailer_wholesale_order_items rwoi
			INNER JOIN retailer_wholesale_orders rwo ON rwo.id = rwoi.order_id
			INNER JOIN wholesaler_products wp ON wp.id = rwoi.product_id
			WHERE rwo.id = (
				SELECT rwoi2.order_id 
				FROM retailer_wholesale_order_items rwoi2 
				WHERE rwoi2.id = $1
			)
			AND wp.wholesaler_id = $2
			AND rwoi.status != 'delivered'
		`
		var remainingCount int
		err := r.DB.QueryRow(checkQuery, itemID, wholesalerID).Scan(&remainingCount)
		if err == nil && remainingCount == 0 {
			// All items delivered, update payment status if offline
			updatePaymentQuery := `
				UPDATE retailer_wholesale_orders
				SET payment_status = 'success', updated_at = NOW()
				WHERE id = (
					SELECT rwoi2.order_id 
					FROM retailer_wholesale_order_items rwoi2 
					WHERE rwoi2.id = $1
				)
				AND payment_method = 'offline'
			`
			_, _ = r.DB.Exec(updatePaymentQuery, itemID)
		}
	}

	return nil
}

// GetOrderItemsByOrderID gets all order items for a specific order, filtered by wholesaler
// Includes all statuses (for history page)
func (r *RetailerWholesaleOrdersRepository) GetOrderItemsByOrderID(orderID int, wholesalerID int) ([]models.RetailerWholesaleOrderItem, error) {
	query := `
		SELECT rwoi.id, rwoi.order_id, rwoi.product_id, rwoi.qty, rwoi.unit_price, rwoi.status
		FROM retailer_wholesale_order_items rwoi
		INNER JOIN wholesaler_products wp ON wp.id = rwoi.product_id
		WHERE rwoi.order_id = $1 AND wp.wholesaler_id = $2
		ORDER BY rwoi.id
	`

	rows, err := r.DB.Query(query, orderID, wholesalerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []models.RetailerWholesaleOrderItem
	for rows.Next() {
		var item models.RetailerWholesaleOrderItem
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

// GetHistoryOrdersByWholesalerID gets completed orders (delivered/rejected items) for history page
func (r *RetailerWholesaleOrdersRepository) GetHistoryOrdersByWholesalerID(wholesalerID int) ([]models.RetailerWholesaleOrder, error) {
	query := `
		SELECT DISTINCT
			rwo.id, rwo.retailer_id, rwo.payment_method, rwo.payment_status,
			rwo.order_status, rwo.total_amount, rwo.scheduled_at,
			rwo.stripe_session_id, rwo.created_at, rwo.updated_at
		FROM retailer_wholesale_orders rwo
		INNER JOIN retailer_wholesale_order_items rwoi ON rwoi.order_id = rwo.id
		INNER JOIN wholesaler_products wp ON wp.id = rwoi.product_id
		WHERE wp.wholesaler_id = $1 AND rwoi.status IN ('rejected', 'delivered')
		ORDER BY rwo.created_at DESC
	`

	rows, err := r.DB.Query(query, wholesalerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query history orders: %w", err)
	}
	defer rows.Close()

	var orders []models.RetailerWholesaleOrder
	orderMap := make(map[int]*models.RetailerWholesaleOrder)

	for rows.Next() {
		var order models.RetailerWholesaleOrder
		var scheduledAt sql.NullTime
		var stripeSessionID sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.RetailerID,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.OrderStatus,
			&order.TotalAmount,
			&scheduledAt,
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
		items, err := r.GetOrderItemsByOrderID(orders[i].ID, wholesalerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get items for history order %d: %w", orders[i].ID, err)
		}
		orders[i].Items = items
	}

	return orders, nil
}

// GetOrdersByRetailerID gets all orders for a retailer (for retailer's order history)
func (r *RetailerWholesaleOrdersRepository) GetOrdersByRetailerID(retailerID int) ([]models.RetailerWholesaleOrder, error) {
	query := `
		SELECT 
			id, retailer_id, payment_method, payment_status, 
			order_status, total_amount, scheduled_at,
			stripe_session_id, created_at, updated_at
		FROM retailer_wholesale_orders
		WHERE retailer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.DB.Query(query, retailerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.RetailerWholesaleOrder

	for rows.Next() {
		var order models.RetailerWholesaleOrder
		var scheduledAt sql.NullTime
		var stripeSessionID sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.RetailerID,
			&order.PaymentMethod,
			&order.PaymentStatus,
			&order.OrderStatus,
			&order.TotalAmount,
			&scheduledAt,
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
		if stripeSessionID.Valid {
			order.StripeSessionID = &stripeSessionID.String
		}

		// Get all items for this order
		itemsQuery := `
			SELECT id, order_id, product_id, qty, unit_price, status
			FROM retailer_wholesale_order_items
			WHERE order_id = $1
			ORDER BY id
		`
		itemRows, err := r.DB.Query(itemsQuery, order.ID)
		if err == nil {
			var items []models.RetailerWholesaleOrderItem
			for itemRows.Next() {
				var item models.RetailerWholesaleOrderItem
				err := itemRows.Scan(
					&item.ID,
					&item.OrderID,
					&item.ProductID,
					&item.Qty,
					&item.UnitPrice,
					&item.Status,
				)
				if err == nil {
					items = append(items, item)
				}
			}
			itemRows.Close()
			order.Items = items
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// UpdateStripeSessionID updates the Stripe session ID for an order
func (r *RetailerWholesaleOrdersRepository) UpdateStripeSessionID(orderID int, sessionID string) error {
	query := `
		UPDATE retailer_wholesale_orders
		SET stripe_session_id = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.DB.Exec(query, sessionID, orderID)
	if err != nil {
		return fmt.Errorf("failed to update Stripe session ID: %w", err)
	}

	return nil
}


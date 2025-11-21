package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"fmt"
)

type IOrdersRepo interface {
	CreateConsumerOrder(order *models.ConsumerOrder) error
	CreateRetailerOrder(order *models.RetailerOrder) error
	GetConsumerOrderBySessionID(sessionID string) (*models.ConsumerOrder, error)
	GetRetailerOrderBySessionID(sessionID string) (*models.RetailerOrder, error)
	UpdateConsumerOrderStatus(sessionID string, status models.OrderStatus) error
	UpdateRetailerOrderStatus(sessionID string, status models.OrderStatus) error
	UpdateConsumerOrderStripeSession(orderID int, sessionID string) error
	GetConsumerOrdersByUserID(userID int) ([]models.ConsumerOrder, error)
	GetConsumerOrdersByRetailerID(retailerID int) ([]models.ConsumerOrder, error)
	GetRetailerOrdersByRetailerID(retailerID int) ([]models.RetailerOrder, error)
	GetRetailerOrdersByWholesalerID(wholesalerID int) ([]models.RetailerOrder, error)
}

type OrdersRepo struct {
	db *sql.DB
}

func NewOrdersRepo(db *sql.DB) *OrdersRepo {
	return &OrdersRepo{db: db}
}

func (r *OrdersRepo) CreateConsumerOrder(order *models.ConsumerOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO retailer_orders (retailer_id, user_id, total_price, status, stripe_session_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(query, order.RetailerId, order.UserId, order.TotalPrice, order.Status, order.StripeSessionId).Scan(&order.Id, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `
		INSERT INTO retailer_order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
	`
	stmt, err := tx.Prepare(itemQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range order.Items {
		_, err = stmt.Exec(order.Id, item.ProductId, item.Quantity, item.Price)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *OrdersRepo) CreateRetailerOrder(order *models.RetailerOrder) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO wholesaler_orders (wholesaler_id, retailer_id, total_price, status, stripe_session_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(query, order.WholesalerId, order.RetailerId, order.TotalPrice, order.Status, order.StripeSessionId).Scan(&order.Id, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `
		INSERT INTO wholesaler_order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
	`
	stmt, err := tx.Prepare(itemQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range order.Items {
		_, err = stmt.Exec(order.Id, item.ProductId, item.Quantity, item.Price)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *OrdersRepo) GetConsumerOrderBySessionID(sessionID string) (*models.ConsumerOrder, error) {
	query := `
		SELECT id, retailer_id, user_id, total_price, status, stripe_session_id, created_at, updated_at
		FROM retailer_orders
		WHERE stripe_session_id = $1
	`
	var order models.ConsumerOrder
	err := r.db.QueryRow(query, sessionID).Scan(
		&order.Id, &order.RetailerId, &order.UserId, &order.TotalPrice, &order.Status, &order.StripeSessionId, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrdersRepo) GetRetailerOrderBySessionID(sessionID string) (*models.RetailerOrder, error) {
	query := `
		SELECT id, wholesaler_id, retailer_id, total_price, status, stripe_session_id, created_at, updated_at
		FROM wholesaler_orders
		WHERE stripe_session_id = $1
	`
	var order models.RetailerOrder
	err := r.db.QueryRow(query, sessionID).Scan(
		&order.Id, &order.WholesalerId, &order.RetailerId, &order.TotalPrice, &order.Status, &order.StripeSessionId, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrdersRepo) UpdateConsumerOrderStatus(sessionID string, status models.OrderStatus) error {
	query := `UPDATE retailer_orders SET status = $1, updated_at = NOW() WHERE stripe_session_id = $2`
	_, err := r.db.Exec(query, status, sessionID)
	return err
}

func (r *OrdersRepo) UpdateRetailerOrderStatus(sessionID string, status models.OrderStatus) error {
	query := `UPDATE wholesaler_orders SET status = $1, updated_at = NOW() WHERE stripe_session_id = $2`
	_, err := r.db.Exec(query, status, sessionID)
	return err
}

func (r *OrdersRepo) UpdateConsumerOrderStripeSession(orderID int, sessionID string) error {
	query := `UPDATE retailer_orders SET stripe_session_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, sessionID, orderID)
	return err
}

func (r *OrdersRepo) GetConsumerOrdersByUserID(userID int) ([]models.ConsumerOrder, error) {
	query := `
		SELECT id, retailer_id, user_id, total_price, status, stripe_session_id, created_at, updated_at
		FROM retailer_orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.ConsumerOrder
	for rows.Next() {
		var o models.ConsumerOrder
		if err := rows.Scan(&o.Id, &o.RetailerId, &o.UserId, &o.TotalPrice, &o.Status, &o.StripeSessionId, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrdersRepo) GetConsumerOrdersByRetailerID(retailerID int) ([]models.ConsumerOrder, error) {
	query := `
		SELECT id, retailer_id, user_id, total_price, status, stripe_session_id, created_at, updated_at
		FROM retailer_orders
		WHERE retailer_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, retailerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.ConsumerOrder
	for rows.Next() {
		var o models.ConsumerOrder
		if err := rows.Scan(&o.Id, &o.RetailerId, &o.UserId, &o.TotalPrice, &o.Status, &o.StripeSessionId, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrdersRepo) GetRetailerOrdersByRetailerID(retailerID int) ([]models.RetailerOrder, error) {
	query := `
		SELECT id, wholesaler_id, retailer_id, total_price, status, stripe_session_id, created_at, updated_at
		FROM wholesaler_orders
		WHERE retailer_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, retailerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.RetailerOrder
	for rows.Next() {
		var o models.RetailerOrder
		if err := rows.Scan(&o.Id, &o.WholesalerId, &o.RetailerId, &o.TotalPrice, &o.Status, &o.StripeSessionId, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrdersRepo) GetRetailerOrdersByWholesalerID(wholesalerID int) ([]models.RetailerOrder, error) {
	query := `
		SELECT id, wholesaler_id, retailer_id, total_price, status, stripe_session_id, created_at, updated_at
		FROM wholesaler_orders
		WHERE wholesaler_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, wholesalerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.RetailerOrder
	for rows.Next() {
		var o models.RetailerOrder
		if err := rows.Scan(&o.Id, &o.WholesalerId, &o.RetailerId, &o.TotalPrice, &o.Status, &o.StripeSessionId, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

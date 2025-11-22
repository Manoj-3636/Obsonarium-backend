package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrRetailerCartItemNotFound = errors.New("retailer cart item not found")

type IRetailerCartRepo interface {
	GetCartItemsByRetailerID(retailerID int) ([]models.RetailerCartItem, error)
	AddCartItem(retailerID int, productID int, quantity int) (int, error)
	RemoveCartItem(retailerID int, productID int) error
	DecreaseCartItem(retailerID int, productID int) (int, error)
	GetCartNumber(retailerID int) (int, error)
	ClearCart(retailerID int) error
}

type RetailerCartRepo struct {
	DB *sql.DB
}

func NewRetailerCartRepo(db *sql.DB) *RetailerCartRepo {
	return &RetailerCartRepo{DB: db}
}

func (repo *RetailerCartRepo) GetCartItemsByRetailerID(retailerID int) ([]models.RetailerCartItem, error) {
	query := `
		SELECT c.id, c.retailer_id, c.product_id, c.quantity,
			   p.id, p.wholesaler_id, p.name, p.price, p.stock_qty, p.image_url, p.description
		FROM retailer_cart_items c
		JOIN wholesaler_products p ON p.id = c.product_id
		WHERE c.retailer_id = $1
		ORDER BY c.id`

	rows, err := repo.DB.Query(query, retailerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartItems []models.RetailerCartItem

	for rows.Next() {
		var item models.RetailerCartItem
		// scan cart item fields then product fields
		err := rows.Scan(
			&item.Id,
			&item.Retailer_id,
			&item.Product_id,
			&item.Quantity,
			&item.Product.Id,
			&item.Product.Wholesaler_id,
			&item.Product.Name,
			&item.Product.Price,
			&item.Product.Stock_qty,
			&item.Product.Image_url,
			&item.Product.Description,
		)
		if err != nil {
			return nil, err
		}
		cartItems = append(cartItems, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cartItems, nil
}

func (repo *RetailerCartRepo) AddCartItem(retailerID int, productID int, quantity int) (int, error) {
	query := `
		INSERT INTO retailer_cart_items (retailer_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (retailer_id, product_id) DO UPDATE
		SET quantity = retailer_cart_items.quantity + EXCLUDED.quantity
		RETURNING quantity
		`

	var newQuantity int
	err := repo.DB.QueryRow(query, retailerID, productID, quantity).Scan(&newQuantity)

	return newQuantity, err
}

func (repo *RetailerCartRepo) DecreaseCartItem(retailerID int, productID int) (int, error) {
	// First decrease quantity
	query := `
		UPDATE retailer_cart_items
		SET quantity = quantity - 1
		WHERE retailer_id = $1 AND product_id = $2
		RETURNING quantity
	`

	var newQty int
	err := repo.DB.QueryRow(query, retailerID, productID).Scan(&newQty)

	if err == sql.ErrNoRows {
		// no cart item found
		return 0, ErrRetailerCartItemNotFound
	}

	if err != nil {
		return 0, err
	}

	// If quantity is now 0 → delete the row
	if newQty <= 0 {
		_, err := repo.DB.Exec(
			`DELETE FROM retailer_cart_items WHERE retailer_id = $1 AND product_id = $2`,
			retailerID, productID,
		)
		return newQty, err
	}

	return newQty, nil
}

func (repo *RetailerCartRepo) RemoveCartItem(retailerID int, productID int) error {
	query := `
		DELETE FROM retailer_cart_items
		WHERE retailer_id = $1 AND product_id = $2`

	result, err := repo.DB.Exec(query, retailerID, productID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRetailerCartItemNotFound
	}

	return nil
}

func (repo *RetailerCartRepo) GetCartNumber(retailerID int) (int, error) {
	query := `
		SELECT COALESCE(SUM(quantity),0)
		FROM retailer_cart_items
		WHERE retailer_id = $1
	`

	var count int
	err := repo.DB.QueryRow(query, retailerID).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *RetailerCartRepo) ClearCart(retailerID int) error {
	query := `
		DELETE FROM retailer_cart_items
		WHERE retailer_id = $1
	`

	_, err := repo.DB.Exec(query, retailerID)
	if err != nil {
		return err
	}

	return nil
}

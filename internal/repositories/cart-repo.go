package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrCartItemNotFound = errors.New("cart item not found")

type ICartRepo interface {
	GetCartItemsByUserID(userID int) ([]models.CartItem, error)
	AddCartItem(userID int, productID int, quantity int) (int, error)
	RemoveCartItem(userID int, productID int) error
	DecreaseCartItem(userID int, productID int) (int, error)
	GetCartNumber(userID int) (int, error)
	ClearCart(userID int) error
}

type CartRepo struct {
	DB *sql.DB
}

func NewCartRepo(db *sql.DB) *CartRepo {
	return &CartRepo{DB: db}
}

func (repo *CartRepo) GetCartItemsByUserID(userID int) ([]models.CartItem, error) {
	query := `
		SELECT c.id, c.user_id, c.product_id, c.quantity,
			   p.id, p.retailer_id, p.name, p.price, p.stock_qty, p.image_url, p.description
		FROM cart_items c
		JOIN retailer_products p ON p.id = c.product_id
		WHERE c.user_id = $1
		ORDER BY c.id`

	rows, err := repo.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartItems []models.CartItem

	for rows.Next() {
		var item models.CartItem
		// scan cart item fields then product fields
		err := rows.Scan(
			&item.Id,
			&item.User_id,
			&item.Product_id,
			&item.Quantity,
			&item.Product.Id,
			&item.Product.Retailer_id,
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

func (repo *CartRepo) AddCartItem(userID int, productID int, quantity int) (int, error) {
	query := `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + EXCLUDED.quantity
		RETURNING quantity
		`

	var newQuantity int
	err := repo.DB.QueryRow(query, userID, productID, quantity).Scan(&newQuantity)

	return newQuantity, err
}

func (repo *CartRepo) DecreaseCartItem(userID int, productID int) (int, error) {
	// First decrease quantity
	query := `
		UPDATE cart_items
		SET quantity = quantity - 1
		WHERE user_id = $1 AND product_id = $2
		RETURNING quantity
	`

	var newQty int
	err := repo.DB.QueryRow(query, userID, productID).Scan(&newQty)

	if err == sql.ErrNoRows {
		// no cart item found
		return 0, ErrCartItemNotFound
	}

	if err != nil {
		return 0, err
	}

	// If quantity is now 0 → delete the row
	if newQty <= 0 {
		_, err := repo.DB.Exec(
			`DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`,
			userID, productID,
		)
		return newQty, err
	}

	return newQty, nil
}

func (repo *CartRepo) RemoveCartItem(userID int, productID int) error {
	query := `
		DELETE FROM cart_items
		WHERE user_id = $1 AND product_id = $2`

	result, err := repo.DB.Exec(query, userID, productID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCartItemNotFound
	}

	return nil
}

func (repo *CartRepo) GetCartNumber(userID int) (int, error) {
	query := `
		SELECT COALESCE(SUM(quantity),0)
		FROM cart_items
		WHERE user_id = $1
	`

	var count int
	err := repo.DB.QueryRow(query, userID).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *CartRepo) ClearCart(userID int) error {
	query := `
		DELETE FROM cart_items
		WHERE user_id = $1
	`

	_, err := repo.DB.Exec(query, userID)
	if err != nil {
		return err
	}

	return nil
}

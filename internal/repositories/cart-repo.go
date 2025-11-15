package repositories

import (
	"Obsonarium-backend/internal/models"
	"database/sql"
	"errors"
)

var ErrCartItemNotFound = errors.New("cart item not found")

type ICartRepo interface {
	GetCartItemsByUserID(userID int) ([]models.CartItem, error)
	AddCartItem(userID int, productID int, quantity int) error
	RemoveCartItem(userID int, productID int) error
}

type CartRepo struct {
	DB *sql.DB
}

func NewCartRepo(db *sql.DB) *CartRepo {
	return &CartRepo{DB: db}
}

func (repo *CartRepo) GetCartItemsByUserID(userID int) ([]models.CartItem, error) {
	query := `
		SELECT id, user_id, product_id, quantity
		FROM cart_items
		WHERE user_id = $1
		ORDER BY id`

	rows, err := repo.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartItems []models.CartItem

	for rows.Next() {
		var item models.CartItem
		err := rows.Scan(
			&item.Id,
			&item.User_id,
			&item.Product_id,
			&item.Quantity,
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

func (repo *CartRepo) AddCartItem(userID int, productID int, quantity int) error {
	query := `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + EXCLUDED.quantity`

	_, err := repo.DB.Exec(query, userID, productID, quantity)
	return err
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

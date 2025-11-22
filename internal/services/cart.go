package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type CartService struct {
	cartRepo     repositories.ICartRepo
	usersRepo    repositories.IUsersRepo
	productsRepo repositories.IRetailerProductsRepo
}

func NewCartService(cartRepo repositories.ICartRepo, usersRepo repositories.IUsersRepo, productsRepo repositories.IRetailerProductsRepo) *CartService {
	return &CartService{
		cartRepo:     cartRepo,
		usersRepo:    usersRepo,
		productsRepo: productsRepo,
	}
}

func (s *CartService) GetCartItemsByEmail(email string) ([]models.CartItem, error) {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("service error fetching user: %w", err)
	}

	cartItems, err := s.cartRepo.GetCartItemsByUserID(user.Id)
	if err != nil {
		return nil, fmt.Errorf("service error fetching cart items: %w", err)
	}

	return cartItems, nil
}

func (s *CartService) AddCartItem(email string, productID int, quantity int) (int, error) {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return 0, fmt.Errorf("service error fetching user: %w", err)
	}

	var newQuantity int
	if quantity == 1 {
		newQuantity, err = s.cartRepo.AddCartItem(user.Id, productID, quantity)
	}

	if quantity == -1 {
		newQuantity, err = s.cartRepo.DecreaseCartItem(user.Id, productID)
	}
	if err != nil {
		return 0, fmt.Errorf("service error adding cart item: %w", err)
	}

	return newQuantity, nil
}

func (s *CartService) RemoveCartItem(email string, productID int) error {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("service error fetching user: %w", err)
	}

	err = s.cartRepo.RemoveCartItem(user.Id, productID)
	if err != nil {
		if err == repositories.ErrCartItemNotFound {
			return err
		}
		return fmt.Errorf("service error removing cart item: %w", err)
	}

	return nil
}

func (s *CartService) GetCartNumberByEmail(email string) (int, error) {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return 0, fmt.Errorf("service error fetching user: %w", err)
	}

	count, err := s.cartRepo.GetCartNumber(user.Id)
	if err != nil {
		return 0, fmt.Errorf("service error fetching cart number: %w", err)
	}

	return count, nil
}

func (s *CartService) ClearCartByUserID(userID int) error {
	err := s.cartRepo.ClearCart(userID)
	if err != nil {
		return fmt.Errorf("service error clearing cart: %w", err)
	}
	return nil
}

type StockValidationError struct {
	ProductID   int
	ProductName string
	Requested   int
	Available   int
}

func (e *StockValidationError) Error() string {
	return fmt.Sprintf("insufficient stock for %s (requested: %d, available: %d)", e.ProductName, e.Requested, e.Available)
}

// ValidateCartStock validates that all cart items have sufficient stock
func (s *CartService) ValidateCartStock(email string) ([]StockValidationError, error) {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("service error fetching user: %w", err)
	}

	cartItems, err := s.cartRepo.GetCartItemsByUserID(user.Id)
	if err != nil {
		return nil, fmt.Errorf("service error fetching cart items: %w", err)
	}

	var validationErrors []StockValidationError

	for _, item := range cartItems {
		product, err := s.productsRepo.GetProduct(item.Product_id)
		if err != nil {
			// If product not found, skip it (will be caught during checkout)
			continue
		}

		if item.Quantity > product.Stock_qty {
			validationErrors = append(validationErrors, StockValidationError{
				ProductID:   item.Product_id,
				ProductName: product.Name,
				Requested:   item.Quantity,
				Available:   product.Stock_qty,
			})
		}
	}

	return validationErrors, nil
}

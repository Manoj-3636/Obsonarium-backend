package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type RetailerCartService struct {
	cartRepo      repositories.IRetailerCartRepo
	retailersRepo repositories.IRetailersRepo
	productsRepo  repositories.IWholesalerProductRepository
}

func NewRetailerCartService(cartRepo repositories.IRetailerCartRepo, retailersRepo repositories.IRetailersRepo, productsRepo repositories.IWholesalerProductRepository) *RetailerCartService {
	return &RetailerCartService{
		cartRepo:      cartRepo,
		retailersRepo: retailersRepo,
		productsRepo:  productsRepo,
	}
}

func (s *RetailerCartService) GetCartItemsByEmail(email string) ([]models.RetailerCartItem, error) {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("service error fetching retailer: %w", err)
	}

	cartItems, err := s.cartRepo.GetCartItemsByRetailerID(retailer.Id)
	if err != nil {
		return nil, fmt.Errorf("service error fetching cart items: %w", err)
	}

	return cartItems, nil
}

func (s *RetailerCartService) AddCartItem(email string, productID int, quantity int) (int, error) {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return 0, fmt.Errorf("service error fetching retailer: %w", err)
	}

	var newQuantity int
	if quantity == 1 {
		newQuantity, err = s.cartRepo.AddCartItem(retailer.Id, productID, quantity)
	}

	if quantity == -1 {
		newQuantity, err = s.cartRepo.DecreaseCartItem(retailer.Id, productID)
	}
	if err != nil {
		return 0, fmt.Errorf("service error adding cart item: %w", err)
	}

	return newQuantity, nil
}

func (s *RetailerCartService) RemoveCartItem(email string, productID int) error {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return fmt.Errorf("service error fetching retailer: %w", err)
	}

	err = s.cartRepo.RemoveCartItem(retailer.Id, productID)
	if err != nil {
		if err == repositories.ErrRetailerCartItemNotFound {
			return err
		}
		return fmt.Errorf("service error removing cart item: %w", err)
	}

	return nil
}

func (s *RetailerCartService) GetCartNumberByEmail(email string) (int, error) {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return 0, fmt.Errorf("service error fetching retailer: %w", err)
	}

	count, err := s.cartRepo.GetCartNumber(retailer.Id)
	if err != nil {
		return 0, fmt.Errorf("service error fetching cart number: %w", err)
	}

	return count, nil
}

func (s *RetailerCartService) ClearCartByRetailerID(retailerID int) error {
	err := s.cartRepo.ClearCart(retailerID)
	if err != nil {
		return fmt.Errorf("service error clearing cart: %w", err)
	}
	return nil
}

type RetailerStockValidationError struct {
	ProductID   int
	ProductName string
	Requested   int
	Available   int
}

func (e *RetailerStockValidationError) Error() string {
	return fmt.Sprintf("insufficient stock for %s (requested: %d, available: %d)", e.ProductName, e.Requested, e.Available)
}

// ValidateCartStock validates that all cart items have sufficient stock
func (s *RetailerCartService) ValidateCartStock(email string) ([]RetailerStockValidationError, error) {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("service error fetching retailer: %w", err)
	}

	cartItems, err := s.cartRepo.GetCartItemsByRetailerID(retailer.Id)
	if err != nil {
		return nil, fmt.Errorf("service error fetching cart items: %w", err)
	}

	var validationErrors []RetailerStockValidationError

	for _, item := range cartItems {
		product, err := s.productsRepo.GetProduct(item.Product_id)
		if err != nil {
			// If product not found, skip it (will be caught during checkout)
			continue
		}

		if item.Quantity > product.Stock_qty {
			validationErrors = append(validationErrors, RetailerStockValidationError{
				ProductID:   item.Product_id,
				ProductName: product.Name,
				Requested:   item.Quantity,
				Available:   product.Stock_qty,
			})
		}
	}

	return validationErrors, nil
}

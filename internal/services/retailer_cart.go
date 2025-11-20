package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type RetailerCartService struct {
	cartRepo      repositories.IRetailerCartRepo
	retailersRepo repositories.IRetailersRepo
}

func NewRetailerCartService(cartRepo repositories.IRetailerCartRepo, retailersRepo repositories.IRetailersRepo) *RetailerCartService {
	return &RetailerCartService{
		cartRepo:      cartRepo,
		retailersRepo: retailersRepo,
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

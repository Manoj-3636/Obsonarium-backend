package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type CartService struct {
	cartRepo  repositories.ICartRepo
	usersRepo repositories.IUsersRepo
}

func NewCartService(cartRepo repositories.ICartRepo, usersRepo repositories.IUsersRepo) *CartService {
	return &CartService{
		cartRepo:  cartRepo,
		usersRepo: usersRepo,
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

func (s *CartService) AddCartItem(email string, productID int, quantity int) error {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("service error fetching user: %w", err)
	}

	err = s.cartRepo.AddCartItem(user.Id, productID, quantity)
	if err != nil {
		return fmt.Errorf("service error adding cart item: %w", err)
	}

	return nil
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

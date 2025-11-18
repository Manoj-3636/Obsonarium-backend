package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type UserAddressesService struct {
	addressesRepo repositories.IUserAddressesRepo
	usersRepo     repositories.IUsersRepo
}

func NewUserAddressesService(addressesRepo repositories.IUserAddressesRepo, usersRepo repositories.IUsersRepo) *UserAddressesService {
	return &UserAddressesService{
		addressesRepo: addressesRepo,
		usersRepo:     usersRepo,
	}
}

func (s *UserAddressesService) GetAddressesByEmail(email string) ([]models.UserAddress, error) {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("service error fetching user: %w", err)
	}

	addresses, err := s.addressesRepo.GetAddressesByUserID(user.Id)
	if err != nil {
		return nil, fmt.Errorf("service error fetching addresses: %w", err)
	}

	return addresses, nil
}

func (s *UserAddressesService) AddAddress(email string, address *models.UserAddress) error {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("service error fetching user: %w", err)
	}

	address.User_id = user.Id
	err = s.addressesRepo.AddAddress(address)
	if err != nil {
		return fmt.Errorf("service error adding address: %w", err)
	}

	return nil
}

func (s *UserAddressesService) RemoveAddress(email string, addressID int) error {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("service error fetching user: %w", err)
	}

	err = s.addressesRepo.RemoveAddress(user.Id, addressID)
	if err != nil {
		if err == repositories.ErrAddressNotFound {
			return err
		}
		return fmt.Errorf("service error removing address: %w", err)
	}

	return nil
}

package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type RetailersService struct {
	retailersRepo repositories.IRetailersRepo
}

func NewRetailersService(retailersRepo repositories.IRetailersRepo) *RetailersService {
	return &RetailersService{
		retailersRepo: retailersRepo,
	}
}

func (s *RetailersService) GetRetailer(id int) (*models.Retailer, error) {
	retailer, err := s.retailersRepo.GetRetailerByID(id)
	if err != nil {
		if err == repositories.ErrRetailerNotFound {
			return &models.Retailer{}, err
		}
		return &models.Retailer{}, fmt.Errorf("service error fetching retailer: %w", err)
	}

	return retailer, nil
}

func (s *RetailersService) GetRetailerByEmail(email string) (*models.Retailer, error) {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		if err == repositories.ErrRetailerNotFound {
			return &models.Retailer{}, err
		}
		return &models.Retailer{}, fmt.Errorf("service error fetching retailer: %w", err)
	}

	return retailer, nil
}

func (s *RetailersService) UpdateRetailer(email string, businessName, phone, address string) (*models.Retailer, error) {
	// First, get the current retailer to preserve the name (which comes from Google OAuth)
	currentRetailer, err := s.GetRetailerByEmail(email)
	if err != nil {
		return &models.Retailer{}, fmt.Errorf("service error fetching retailer: %w", err)
	}

	retailer := &models.Retailer{
		Email:        email,
		Name:         currentRetailer.Name, // Preserve name from Google OAuth
		BusinessName: businessName,
		Phone:        phone,
		Address:      address,
	}

	err = s.retailersRepo.UpdateRetailer(retailer)
	if err != nil {
		if err == repositories.ErrRetailerNotFound {
			return &models.Retailer{}, err
		}
		return &models.Retailer{}, fmt.Errorf("service error updating retailer: %w", err)
	}

	// Fetch updated retailer to return complete data
	updatedRetailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return &models.Retailer{}, fmt.Errorf("service error fetching updated retailer: %w", err)
	}

	return updatedRetailer, nil
}

// IsOnboarded checks if a retailer has completed onboarding (has business_name, phone, and address)
func (s *RetailersService) IsOnboarded(email string) (bool, error) {
	retailer, err := s.GetRetailerByEmail(email)
	if err != nil {
		return false, err
	}

	return retailer.BusinessName != "" && retailer.Phone != "" && retailer.Address != "", nil
}

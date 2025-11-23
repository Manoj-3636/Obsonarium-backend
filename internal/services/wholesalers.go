package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type WholesalersService struct {
	wholesalersRepo repositories.IWholesalersRepo
}

func NewWholesalersService(wholesalersRepo repositories.IWholesalersRepo) *WholesalersService {
	return &WholesalersService{
		wholesalersRepo: wholesalersRepo,
	}
}

func (s *WholesalersService) GetWholesaler(id int) (*models.Wholesaler, error) {
	wholesaler, err := s.wholesalersRepo.GetWholesalerByID(id)
	if err != nil {
		if err == repositories.ErrWholesalerNotFound {
			return &models.Wholesaler{}, err
		}
		return &models.Wholesaler{}, fmt.Errorf("service error fetching wholesaler: %w", err)
	}

	return wholesaler, nil
}

func (s *WholesalersService) GetWholesalerByEmail(email string) (*models.Wholesaler, error) {
	wholesaler, err := s.wholesalersRepo.GetWholesalerByEmail(email)
	if err != nil {
		if err == repositories.ErrWholesalerNotFound {
			return &models.Wholesaler{}, err
		}
		return &models.Wholesaler{}, fmt.Errorf("service error fetching wholesaler: %w", err)
	}

	return wholesaler, nil
}

func (s *WholesalersService) UpdateWholesaler(email string, businessName, phone, address, streetAddress, city, state, postalCode, country string, latitude, longitude *float64) (*models.Wholesaler, error) {
	// First, get the current wholesaler to preserve the name (which comes from Google OAuth)
	currentWholesaler, err := s.GetWholesalerByEmail(email)
	if err != nil {
		return &models.Wholesaler{}, fmt.Errorf("service error fetching wholesaler: %w", err)
	}

	wholesaler := &models.Wholesaler{
		Email:         email,
		Name:          currentWholesaler.Name, // Preserve name from Google OAuth
		BusinessName:  businessName,
		Phone:         phone,
		Address:       address,
		StreetAddress: streetAddress,
		City:          city,
		State:         state,
		PostalCode:    postalCode,
		Country:       country,
		Latitude:      latitude,
		Longitude:     longitude,
	}

	err = s.wholesalersRepo.UpdateWholesaler(wholesaler)
	if err != nil {
		if err == repositories.ErrWholesalerNotFound {
			return &models.Wholesaler{}, err
		}
		return &models.Wholesaler{}, fmt.Errorf("service error updating wholesaler: %w", err)
	}

	// Fetch updated wholesaler to return complete data
	updatedWholesaler, err := s.wholesalersRepo.GetWholesalerByEmail(email)
	if err != nil {
		return &models.Wholesaler{}, fmt.Errorf("service error fetching updated wholesaler: %w", err)
	}

	return updatedWholesaler, nil
}

// IsOnboarded checks if a wholesaler has completed onboarding (has business_name, phone, and address)
func (s *WholesalersService) IsOnboarded(email string) (bool, error) {
	wholesaler, err := s.GetWholesalerByEmail(email)
	if err != nil {
		return false, err
	}

	return wholesaler.BusinessName != "" && wholesaler.Phone != "" && wholesaler.Address != "", nil
}

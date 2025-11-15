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

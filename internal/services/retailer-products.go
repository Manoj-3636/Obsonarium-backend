package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type RetailerProductsService struct {
	productsRepo repositories.IRetailerProductsRepo
}

func NewRetailerProductsService(productsRepo repositories.IRetailerProductsRepo) *RetailerProductsService {
	return &RetailerProductsService{
		productsRepo: productsRepo,
	}
}

func (s *RetailerProductsService) GetProducts() ([]models.RetailerProduct, error) {
	products, err := s.productsRepo.GetProducts()
	if err != nil {
		return nil, fmt.Errorf("service error fetching products: %w", err)
	}

	return products, nil
}

func (s *RetailerProductsService) GetProduct(id int) (*models.RetailerProduct, error) {
	product, err := s.productsRepo.GetProduct(id)
	if err != nil {
		if err == repositories.ErrProductNotFound {
			return &models.RetailerProduct{}, err
		}
		return &models.RetailerProduct{}, fmt.Errorf("service error fetching product: %w", err)
	}

	return product, nil
}


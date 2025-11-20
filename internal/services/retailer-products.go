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

func (s *RetailerProductsService) GetProducts(q string) ([]models.RetailerProduct, error) {
	if q != "" {
		// full-text search
		return s.productsRepo.SearchProducts(q)
	}

	// default 10 products
	return s.productsRepo.GetProducts()
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

func (s *RetailerProductsService) GetProductsByRetailerID(retailerID int) ([]models.RetailerProduct, error) {
	products, err := s.productsRepo.GetProductsByRetailerID(retailerID)
	if err != nil {
		return nil, fmt.Errorf("service error fetching products by retailer ID: %w", err)
	}

	return products, nil
}

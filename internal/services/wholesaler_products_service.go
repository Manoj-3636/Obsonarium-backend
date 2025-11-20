package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type WholesalerProductsService struct {
	productsRepo repositories.IWholesalerProductRepository
}

func NewWholesalerProductsService(productsRepo repositories.IWholesalerProductRepository) *WholesalerProductsService {
	return &WholesalerProductsService{
		productsRepo: productsRepo,
	}
}

func (s *WholesalerProductsService) GetProducts(q string) ([]models.WholesalerProduct, error) {
	if q != "" {
		// full-text search
		return s.productsRepo.SearchProducts(q)
	}

	// default 10 products
	return s.productsRepo.GetProducts()
}

func (s *WholesalerProductsService) GetProduct(id int) (*models.WholesalerProduct, error) {
	product, err := s.productsRepo.GetProduct(id)
	if err != nil {
		if err == repositories.ErrWholesalerProductNotFound {
			return &models.WholesalerProduct{}, err
		}
		return &models.WholesalerProduct{}, fmt.Errorf("service error fetching product: %w", err)
	}

	return product, nil
}

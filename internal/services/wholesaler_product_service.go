package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type WholesalerProductService struct {
	productRepo repositories.IWholesalerProductRepository
}

func NewWholesalerProductService(productRepo repositories.IWholesalerProductRepository) *WholesalerProductService {
	return &WholesalerProductService{
		productRepo: productRepo,
	}
}

func (s *WholesalerProductService) GetProductsByWholesalerID(wholesalerID int) ([]models.WholesalerProduct, error) {
	products, err := s.productRepo.GetProductsByWholesalerID(wholesalerID)
	if err != nil {
		return nil, fmt.Errorf("service error fetching products by wholesaler ID: %w", err)
	}
	return products, nil
}

func (s *WholesalerProductService) GetProductByIDForWholesaler(productID int, wholesalerID int) (*models.WholesalerProduct, error) {
	product, err := s.productRepo.GetProductByIDForWholesaler(productID, wholesalerID)
	if err != nil {
		if err == repositories.ErrWholesalerProductNotFound {
			return &models.WholesalerProduct{}, err
		}
		return &models.WholesalerProduct{}, fmt.Errorf("service error fetching product by ID for wholesaler: %w", err)
	}
	return product, nil
}

func (s *WholesalerProductService) CreateProduct(product *models.WholesalerProduct) (*models.WholesalerProduct, error) {
	createdProduct, err := s.productRepo.CreateProduct(product)
	if err != nil {
		return &models.WholesalerProduct{}, fmt.Errorf("service error creating product: %w", err)
	}
	return createdProduct, nil
}

func (s *WholesalerProductService) UpdateProduct(product *models.WholesalerProduct) (*models.WholesalerProduct, error) {
	updatedProduct, err := s.productRepo.UpdateProduct(product)
	if err != nil {
		if err == repositories.ErrWholesalerProductNotFound {
			return &models.WholesalerProduct{}, err
		}
		return &models.WholesalerProduct{}, fmt.Errorf("service error updating product: %w", err)
	}
	return updatedProduct, nil
}

func (s *WholesalerProductService) DeleteProduct(productID int, wholesalerID int) error {
	err := s.productRepo.DeleteProduct(productID, wholesalerID)
	if err != nil {
		if err == repositories.ErrWholesalerProductNotFound {
			return err
		}
		return fmt.Errorf("service error deleting product: %w", err)
	}
	return nil
}

package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type ProductRepository interface {
	GetProductsByRetailerID(retailerID int) ([]models.RetailerProduct, error)
	GetProductByIDForRetailer(productID int, retailerID int) (*models.RetailerProduct, error)
	CreateProduct(product *models.RetailerProduct) (*models.RetailerProduct, error)
	UpdateProduct(product *models.RetailerProduct) (*models.RetailerProduct, error)
	DeleteProduct(productID int, retailerID int) error
}

type ProductService struct {
	productRepo ProductRepository
}

func NewProductService(productRepo ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

func (s *ProductService) GetProductsByRetailer(retailerID int) ([]models.RetailerProduct, error) {
	products, err := s.productRepo.GetProductsByRetailerID(retailerID)
	if err != nil {
		return nil, fmt.Errorf("service error fetching products: %w", err)
	}

	return products, nil
}

func (s *ProductService) GetProductByID(productID int, retailerID int) (*models.RetailerProduct, error) {
	product, err := s.productRepo.GetProductByIDForRetailer(productID, retailerID)
	if err != nil {
		if err == repositories.ErrProductNotFound {
			return &models.RetailerProduct{}, err
		}
		return &models.RetailerProduct{}, fmt.Errorf("service error fetching product: %w", err)
	}

	return product, nil
}

func (s *ProductService) CreateProduct(product *models.RetailerProduct) (*models.RetailerProduct, error) {
	createdProduct, err := s.productRepo.CreateProduct(product)
	if err != nil {
		return &models.RetailerProduct{}, fmt.Errorf("service error creating product: %w", err)
	}

	return createdProduct, nil
}

func (s *ProductService) UpdateProduct(product *models.RetailerProduct) (*models.RetailerProduct, error) {
	updatedProduct, err := s.productRepo.UpdateProduct(product)
	if err != nil {
		if err == repositories.ErrProductNotFound {
			return &models.RetailerProduct{}, err
		}
		return &models.RetailerProduct{}, fmt.Errorf("service error updating product: %w", err)
	}

	return updatedProduct, nil
}

func (s *ProductService) DeleteProduct(productID int, retailerID int) error {
	err := s.productRepo.DeleteProduct(productID, retailerID)
	if err != nil {
		if err == repositories.ErrProductNotFound {
			return err
		}
		return fmt.Errorf("service error deleting product: %w", err)
	}

	return nil
}


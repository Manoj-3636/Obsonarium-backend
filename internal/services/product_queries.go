package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"fmt"
)

type ProductQueriesService struct {
	queriesRepo repositories.IProductQueriesRepo
}

func NewProductQueriesService(queriesRepo repositories.IProductQueriesRepo) *ProductQueriesService {
	return &ProductQueriesService{
		queriesRepo: queriesRepo,
	}
}

func (s *ProductQueriesService) GetQueriesByRetailerID(retailerID int) ([]models.ProductQuery, error) {
	queries, err := s.queriesRepo.GetQueriesByRetailerID(retailerID)
	if err != nil {
		return nil, fmt.Errorf("service error fetching queries by retailer ID: %w", err)
	}
	return queries, nil
}

func (s *ProductQueriesService) GetQueriesByProductID(productID int) ([]models.ProductQuery, error) {
	queries, err := s.queriesRepo.GetQueriesByProductID(productID)
	if err != nil {
		return nil, fmt.Errorf("service error fetching queries by product ID: %w", err)
	}
	return queries, nil
}

func (s *ProductQueriesService) CreateQuery(query *models.ProductQuery) (*models.ProductQuery, error) {
	createdQuery, err := s.queriesRepo.CreateQuery(query)
	if err != nil {
		return nil, fmt.Errorf("service error creating query: %w", err)
	}
	return createdQuery, nil
}

func (s *ProductQueriesService) ResolveQuery(queryID int, responseText string) (*models.ProductQuery, error) {
	resolvedQuery, err := s.queriesRepo.ResolveQuery(queryID, responseText)
	if err != nil {
		if err == repositories.ErrProductQueryNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("service error resolving query: %w", err)
	}
	return resolvedQuery, nil
}


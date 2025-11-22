package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"errors"
	"fmt"
)

type ConsumerOrdersService struct {
	OrdersRepo *repositories.ConsumerOrdersRepository
}

func NewConsumerOrdersService(ordersRepo *repositories.ConsumerOrdersRepository) *ConsumerOrdersService {
	return &ConsumerOrdersService{
		OrdersRepo: ordersRepo,
	}
}

// GetActiveOrdersByRetailerID gets active orders for a retailer (excludes delivered/rejected items)
func (s *ConsumerOrdersService) GetActiveOrdersByRetailerID(retailerID int) ([]models.ConsumerOrder, error) {
	orders, err := s.OrdersRepo.GetActiveOrdersByRetailerID(retailerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}
	return orders, nil
}

// GetHistoryOrdersByRetailerID gets completed orders for history (delivered/rejected items only)
func (s *ConsumerOrdersService) GetHistoryOrdersByRetailerID(retailerID int) ([]models.ConsumerOrder, error) {
	orders, err := s.OrdersRepo.GetHistoryOrdersByRetailerID(retailerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history orders: %w", err)
	}
	return orders, nil
}

// UpdateOrderItemStatus updates the status of an order item
func (s *ConsumerOrdersService) UpdateOrderItemStatus(itemID int, retailerID int, status string) error {
	// Validate status - retailers cannot set items to "pending" as it's the initial state only
	validStatuses := map[string]bool{
		"accepted":  true,
		"rejected":  true,
		"shipped":   true,
		"delivered": true,
	}

	if !validStatuses[status] {
		return errors.New("invalid status")
	}

	// Explicitly reject attempts to set status to pending
	if status == "pending" {
		return errors.New("cannot set status to pending - it is the initial state only")
	}

	err := s.OrdersRepo.UpdateOrderItemStatus(itemID, retailerID, status)
	if err != nil {
		return fmt.Errorf("failed to update order item status: %w", err)
	}

	return nil
}

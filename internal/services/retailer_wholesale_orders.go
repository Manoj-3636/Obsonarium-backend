package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"errors"
	"fmt"
)

type RetailerWholesaleOrdersService struct {
	OrdersRepo *repositories.RetailerWholesaleOrdersRepository
}

func NewRetailerWholesaleOrdersService(ordersRepo *repositories.RetailerWholesaleOrdersRepository) *RetailerWholesaleOrdersService {
	return &RetailerWholesaleOrdersService{
		OrdersRepo: ordersRepo,
	}
}

// GetActiveOrdersByWholesalerID gets active orders for a wholesaler (excludes delivered/rejected items)
func (s *RetailerWholesaleOrdersService) GetActiveOrdersByWholesalerID(wholesalerID int) ([]models.RetailerWholesaleOrder, error) {
	orders, err := s.OrdersRepo.GetActiveOrdersByWholesalerID(wholesalerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}
	return orders, nil
}

// GetHistoryOrdersByWholesalerID gets completed orders for history (delivered/rejected items only)
func (s *RetailerWholesaleOrdersService) GetHistoryOrdersByWholesalerID(wholesalerID int) ([]models.RetailerWholesaleOrder, error) {
	orders, err := s.OrdersRepo.GetHistoryOrdersByWholesalerID(wholesalerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history orders: %w", err)
	}
	return orders, nil
}

// GetOrdersByRetailerID gets all orders for a retailer
func (s *RetailerWholesaleOrdersService) GetOrdersByRetailerID(retailerID int) ([]models.RetailerWholesaleOrder, error) {
	orders, err := s.OrdersRepo.GetOrdersByRetailerID(retailerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	return orders, nil
}

// UpdateOrderItemStatus updates the status of an order item
func (s *RetailerWholesaleOrdersService) UpdateOrderItemStatus(itemID int, wholesalerID int, status string) error {
	// Validate status - wholesalers cannot set items to "pending" as it's the initial state only
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

	err := s.OrdersRepo.UpdateOrderItemStatus(itemID, wholesalerID, status)
	if err != nil {
		return fmt.Errorf("failed to update order item status: %w", err)
	}

	return nil
}


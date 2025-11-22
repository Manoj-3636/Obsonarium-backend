package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"errors"
	"fmt"
)

type ConsumerOrdersService struct {
	OrdersRepo   *repositories.ConsumerOrdersRepository
	UsersRepo    repositories.IUsersRepo
	EmailService *EmailService
}

func NewConsumerOrdersService(ordersRepo *repositories.ConsumerOrdersRepository, usersRepo repositories.IUsersRepo, emailService *EmailService) *ConsumerOrdersService {
	return &ConsumerOrdersService{
		OrdersRepo:   ordersRepo,
		UsersRepo:    usersRepo,
		EmailService: emailService,
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

	orderID, consumerID, productName, err := s.OrdersRepo.UpdateOrderItemStatus(itemID, retailerID, status)
	if err != nil {
		return fmt.Errorf("failed to update order item status: %w", err)
	}

	// Send email notification to consumer
	if s.EmailService != nil && s.UsersRepo != nil {
		user, err := s.UsersRepo.GetUserByID(consumerID)
		if err != nil {
			// Log error but don't fail the status update
			fmt.Printf("failed to fetch user for email notification: %v\n", err)
		} else {
			statusText := map[string]string{
				"accepted":  "Accepted",
				"rejected":  "Rejected",
				"shipped":   "Shipped",
				"delivered": "Delivered",
			}
			statusDisplay := statusText[status]
			if statusDisplay == "" {
				statusDisplay = status
			}

			subject := fmt.Sprintf("Order #%d Status Update", orderID)
			body := fmt.Sprintf("Dear customer,\n\nYour order #%d has been updated.\n\nProduct: %s\nNew Status: %s\n\nThank you for shopping with us!", orderID, productName, statusDisplay)

			err = s.EmailService.SendEmail(user.Email, subject, body)
			if err != nil {
				// Log error but don't fail the status update
				fmt.Printf("failed to send email notification: %v\n", err)
			} else {
				fmt.Printf("Email notification sent to consumer %s for order #%d status update\n", user.Email, orderID)
			}
		}
	}

	return nil
}

// GetOrdersByConsumerID gets all orders for a consumer (both ongoing and past)
func (s *ConsumerOrdersService) GetOrdersByConsumerID(consumerID int) ([]models.ConsumerOrder, error) {
	orders, err := s.OrdersRepo.GetOrdersByConsumerID(consumerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	return orders, nil
}

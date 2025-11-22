package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"context"
	"fmt"
)

type CheckoutService struct {
	OrdersRepo   *repositories.ConsumerOrdersRepository
	ProductsRepo *repositories.RetailerProductsRepo
	CartRepo     repositories.ICartRepo
}

func NewCheckoutService(ordersRepo *repositories.ConsumerOrdersRepository, productsRepo *repositories.RetailerProductsRepo, cartRepo repositories.ICartRepo) *CheckoutService {
	return &CheckoutService{
		OrdersRepo:   ordersRepo,
		ProductsRepo: productsRepo,
		CartRepo:     cartRepo,
	}
}

type CheckoutRequest struct {
	AddressID     int        `json:"address_id"`
	PaymentMethod string     `json:"payment_method"`
	CartItems     []CartItem `json:"cart_items"`
}

type CartItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func (s *CheckoutService) ProcessCheckout(ctx context.Context, userID int, req CheckoutRequest) (*models.ConsumerOrder, error) {
	if req.PaymentMethod != "offline" {
		return nil, fmt.Errorf("only offline payment is currently supported")
	}

	var totalAmount float64
	var orderItems []models.ConsumerOrderItem

	// Validate items and calculate total
	for _, item := range req.CartItems {
		product, err := s.ProductsRepo.GetProduct(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product %d: %w", item.ProductID, err)
		}

		lineTotal := product.Price * float64(item.Quantity)
		totalAmount += lineTotal

		orderItems = append(orderItems, models.ConsumerOrderItem{
			ProductID: item.ProductID,
			Qty:       item.Quantity,
			UnitPrice: product.Price,
		})
	}

	order := &models.ConsumerOrder{
		ConsumerID:    userID,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: "pending",
		OrderStatus:   "placed",
		TotalAmount:   totalAmount,
		AddressID:     &req.AddressID,
	}

	createdOrder, err := s.OrdersRepo.CreateOrder(ctx, order, orderItems)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Clear the user's cart after successful order creation
	err = s.CartRepo.ClearCart(userID)
	if err != nil {
		// Log the error but don't fail the order creation
		// The order was already created successfully
		return createdOrder, fmt.Errorf("order created but failed to clear cart: %w", err)
	}

	return createdOrder, nil
}

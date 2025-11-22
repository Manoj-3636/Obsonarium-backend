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
	StripeService *StripeService
	UsersRepo    repositories.IUsersRepo
}

func NewCheckoutService(ordersRepo *repositories.ConsumerOrdersRepository, productsRepo *repositories.RetailerProductsRepo, cartRepo repositories.ICartRepo, stripeService *StripeService, usersRepo repositories.IUsersRepo) *CheckoutService {
	return &CheckoutService{
		OrdersRepo:   ordersRepo,
		ProductsRepo: productsRepo,
		CartRepo:     cartRepo,
		StripeService: stripeService,
		UsersRepo:    usersRepo,
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

type CheckoutResponse struct {
	Order       *models.ConsumerOrder `json:"order,omitempty"`
	SessionURL  string                `json:"session_url,omitempty"`
	Message     string                `json:"message"`
}

func (s *CheckoutService) ProcessCheckout(ctx context.Context, userID int, req CheckoutRequest) (*CheckoutResponse, error) {
	var totalAmount float64
	var orderItems []models.ConsumerOrderItem

	// Validate items, check stock, and calculate total
	for _, item := range req.CartItems {
		product, err := s.ProductsRepo.GetProduct(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product %d: %w", item.ProductID, err)
		}

		// Check stock availability
		if product.Stock_qty < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)", product.Name, product.Stock_qty, item.Quantity)
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

	// Handle offline payment
	if req.PaymentMethod == "offline" {
		createdOrder, err := s.OrdersRepo.CreateOrder(ctx, order, orderItems)
		if err != nil {
			return nil, fmt.Errorf("failed to create order: %w", err)
		}

		// Update stock for each product in the order
		for _, item := range orderItems {
			err := s.ProductsRepo.DecrementStock(item.ProductID, item.Qty)
			if err != nil {
				// Log the error but don't fail the order creation
				// The order is already created, so we'll handle stock issues separately
				fmt.Printf("Warning: Failed to decrement stock for product %d: %v\n", item.ProductID, err)
			}
		}

		// Clear the user's cart after successful order creation
		err = s.CartRepo.ClearCart(userID)
		if err != nil {
			// Log the error but don't fail the order creation
			return &CheckoutResponse{
				Order:   createdOrder,
				Message: "Order placed successfully (cart clearing failed)",
			}, nil
		}

		return &CheckoutResponse{
			Order:   createdOrder,
			Message: "Order placed successfully",
		}, nil
	}

	// Handle online payment (Stripe)
	if req.PaymentMethod == "online" {
		// Create order first
		createdOrder, err := s.OrdersRepo.CreateOrder(ctx, order, orderItems)
		if err != nil {
			return nil, fmt.Errorf("failed to create order: %w", err)
		}

		// Update stock for each product in the order
		for _, item := range orderItems {
			err := s.ProductsRepo.DecrementStock(item.ProductID, item.Qty)
			if err != nil {
				// Log the error but don't fail the order creation
				// The order is already created, so we'll handle stock issues separately
				fmt.Printf("Warning: Failed to decrement stock for product %d: %v\n", item.ProductID, err)
			}
		}

		// Get user email for Stripe
		user, err := s.UsersRepo.GetUserByID(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Get product names for Stripe
		productNames := make(map[int]string)
		for _, item := range orderItems {
			product, err := s.ProductsRepo.GetProduct(item.ProductID)
			if err == nil {
				productNames[item.ProductID] = product.Name
			}
		}

		// Create Stripe checkout session
		successURL := fmt.Sprintf("http://localhost:5173/checkout/success?order_id=%d", createdOrder.ID)
		cancelURL := "http://localhost:5173/checkout/cancel"

		sessionID, sessionURL, err := s.StripeService.CreateCheckoutSession(CreateCheckoutSessionParams{
			OrderID:       createdOrder.ID,
			UserID:        userID,
			Items:         orderItems,
			ProductNames:  productNames,
			TotalAmount:   totalAmount,
			SuccessURL:    successURL,
			CancelURL:     cancelURL,
			CustomerEmail: user.Email,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Stripe checkout session: %w", err)
		}

		// Update order with Stripe session ID in database
		err = s.OrdersRepo.UpdateStripeSessionID(createdOrder.ID, sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to update Stripe session ID: %w", err)
		}
		createdOrder.StripeSessionID = &sessionID

		// Clear the user's cart after creating the Stripe session
		err = s.CartRepo.ClearCart(userID)
		if err != nil {
			// Log the error but don't fail
		}

		return &CheckoutResponse{
			Order:      createdOrder,
			SessionURL: sessionURL,
			Message:    "Redirecting to payment...",
		}, nil
	}

	return nil, fmt.Errorf("invalid payment method: %s", req.PaymentMethod)
}

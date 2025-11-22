package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"context"
	"fmt"
)

type RetailerCheckoutService struct {
	OrdersRepo      *repositories.RetailerWholesaleOrdersRepository
	ProductsRepo    repositories.IWholesalerProductRepository
	CartRepo        repositories.IRetailerCartRepo
	StripeService   *StripeService
	RetailersRepo   repositories.IRetailersRepo
}

func NewRetailerCheckoutService(
	ordersRepo *repositories.RetailerWholesaleOrdersRepository,
	productsRepo repositories.IWholesalerProductRepository,
	cartRepo repositories.IRetailerCartRepo,
	stripeService *StripeService,
	retailersRepo repositories.IRetailersRepo,
) *RetailerCheckoutService {
	return &RetailerCheckoutService{
		OrdersRepo:    ordersRepo,
		ProductsRepo:  productsRepo,
		CartRepo:      cartRepo,
		StripeService: stripeService,
		RetailersRepo: retailersRepo,
	}
}

type RetailerCheckoutRequest struct {
	PaymentMethod string                `json:"payment_method"`
	CartItems     []RetailerCartItem     `json:"cart_items"`
}

type RetailerCartItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type RetailerCheckoutResponse struct {
	Order      *models.RetailerWholesaleOrder `json:"order,omitempty"`
	SessionURL string                          `json:"session_url,omitempty"`
	Message    string                          `json:"message"`
}

func (s *RetailerCheckoutService) ProcessCheckout(ctx context.Context, retailerID int, req RetailerCheckoutRequest) (*RetailerCheckoutResponse, error) {
	var totalAmount float64
	var orderItems []models.RetailerWholesaleOrderItem

	// Validate items and calculate total
	for _, item := range req.CartItems {
		product, err := s.ProductsRepo.GetProduct(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product %d: %w", item.ProductID, err)
		}

		lineTotal := product.Price * float64(item.Quantity)
		totalAmount += lineTotal

		orderItems = append(orderItems, models.RetailerWholesaleOrderItem{
			ProductID: item.ProductID,
			Qty:       item.Quantity,
			UnitPrice: product.Price,
		})
	}

	order := &models.RetailerWholesaleOrder{
		RetailerID:    retailerID,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: "pending",
		OrderStatus:   "placed",
		TotalAmount:   totalAmount,
	}

	// Handle offline payment
	if req.PaymentMethod == "offline" {
		createdOrder, err := s.OrdersRepo.CreateOrder(ctx, order, orderItems)
		if err != nil {
			return nil, fmt.Errorf("failed to create order: %w", err)
		}

		// Clear the retailer's cart after successful order creation
		err = s.CartRepo.ClearCart(retailerID)
		if err != nil {
			// Log the error but don't fail the order creation
			return &RetailerCheckoutResponse{
				Order:   createdOrder,
				Message: "Order placed successfully (cart clearing failed)",
			}, nil
		}

		return &RetailerCheckoutResponse{
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

		// Get retailer email for Stripe
		retailer, err := s.RetailersRepo.GetRetailerByID(retailerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get retailer: %w", err)
		}

		// Get product names for Stripe
		productNames := make(map[int]string)
		for _, item := range orderItems {
			product, err := s.ProductsRepo.GetProduct(item.ProductID)
			if err == nil {
				productNames[item.ProductID] = product.Name
			}
		}

		// Convert order items to Stripe line items format
		var stripeItems []models.ConsumerOrderItem
		for _, item := range orderItems {
			stripeItems = append(stripeItems, models.ConsumerOrderItem{
				ProductID: item.ProductID,
				Qty:       item.Qty,
				UnitPrice: item.UnitPrice,
			})
		}

		// Create Stripe checkout session
		successURL := fmt.Sprintf("http://localhost:5174/wholesale/checkout/success?order_id=%d", createdOrder.ID)
		cancelURL := "http://localhost:5174/wholesale/checkout/cancel"

		sessionID, sessionURL, err := s.StripeService.CreateCheckoutSession(CreateCheckoutSessionParams{
			OrderID:       createdOrder.ID,
			UserID:        retailerID,
			Items:         stripeItems,
			ProductNames:  productNames,
			TotalAmount:   totalAmount,
			SuccessURL:    successURL,
			CancelURL:     cancelURL,
			CustomerEmail: retailer.Email,
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

		// Clear the retailer's cart after creating the Stripe session
		err = s.CartRepo.ClearCart(retailerID)
		if err != nil {
			// Log the error but don't fail
		}

		return &RetailerCheckoutResponse{
			Order:      createdOrder,
			SessionURL: sessionURL,
			Message:    "Redirecting to payment...",
		}, nil
	}

	return nil, fmt.Errorf("invalid payment method: %s", req.PaymentMethod)
}


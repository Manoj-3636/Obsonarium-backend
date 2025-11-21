package services

import (
	"Obsonarium-backend/internal/models"
	"Obsonarium-backend/internal/repositories"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/stripe/stripe-go/v79"
)

type OrdersService struct {
	ordersRepo          repositories.IOrdersRepo
	cartService         CartService
	retailerCartService RetailerCartService
	stripeService       *StripeService
	emailService        *EmailService
	usersRepo           repositories.IUsersRepo
	retailersRepo       repositories.IRetailersRepo
}

func NewOrdersService(ordersRepo repositories.IOrdersRepo, cartService CartService, retailerCartService RetailerCartService, stripeService *StripeService, emailService *EmailService, usersRepo repositories.IUsersRepo, retailersRepo repositories.IRetailersRepo) *OrdersService {
	return &OrdersService{
		ordersRepo:          ordersRepo,
		cartService:         cartService,
		retailerCartService: retailerCartService,
		stripeService:       stripeService,
		emailService:        emailService,
		usersRepo:           usersRepo,
		retailersRepo:       retailersRepo,
	}
}

func (s *OrdersService) CreateConsumerCheckoutByEmail(email, successURL, cancelURL string, addressID int) (string, error) {
	// Get user ID from email
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	return s.CreateConsumerCheckout(user.Id, successURL, cancelURL, addressID)
}

func (s *OrdersService) CreateConsumerCheckout(userID int, successURL, cancelURL string, addressID int) (string, error) {
	// Get cart items for the user
	cartItems, err := s.cartService.GetCartItemsByUserID(userID)
	if err != nil {
		return "", fmt.Errorf("failed to get cart items: %w", err)
	}

	if len(cartItems) == 0 {
		return "", fmt.Errorf("cart is empty")
	}

	// Convert cart items to Stripe line items
	var lineItems []*stripe.CheckoutSessionLineItemParams
	var totalPrice float64
	for _, item := range cartItems {
		price := item.Product.Price * float64(item.Quantity)
		totalPrice += price

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("inr"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(item.Product.Name),
				},
				UnitAmount: stripe.Int64(int64(item.Product.Price * 100)), // Convert to paise
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}

	// Get retailer ID from first cart item (all items should be from same retailer in a single checkout)
	if len(cartItems) == 0 {
		return "", fmt.Errorf("cart is empty")
	}
	retailerID := cartItems[0].Product.Retailer_id

	// Create order items
	orderItems := make([]models.ConsumerOrderItem, len(cartItems))
	for i, item := range cartItems {
		orderItems[i] = models.ConsumerOrderItem{
			ProductId: item.Product_id,
			Quantity:  item.Quantity,
			Price:     item.Product.Price,
		}
	}

	// Create order in database
	order := models.ConsumerOrder{
		RetailerId: retailerID,
		UserId:     userID,
		AddressId:  addressID,
		TotalPrice: totalPrice,
		Status:     models.OrderStatusPending,
		Items:      orderItems,
	}

	if err := s.ordersRepo.CreateConsumerOrder(&order); err != nil {
		return "", fmt.Errorf("failed to create order in db: %w", err)
	}

	// Create Stripe checkout session
	sessionID, sessionURL, err := s.stripeService.CreateCheckoutSession(
		lineItems,
		successURL,
		cancelURL,
		fmt.Sprintf("%d", userID),
		map[string]string{"type": "consumer", "order_id": fmt.Sprintf("%d", order.Id)},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe session: %w", err)
	}

	// Update order with stripe session ID
	if err := s.ordersRepo.UpdateConsumerOrderStripeSession(order.Id, sessionID); err != nil {
		return "", fmt.Errorf("failed to update order with stripe session: %w", err)
	}

	return sessionURL, nil
}

func (s *OrdersService) CreateRetailerCheckoutByEmail(email string, successURL, cancelURL string) (string, error) {
	retailer, err := s.retailersRepo.GetRetailerByEmail(email)
	if err != nil {
		return "", fmt.Errorf("failed to get retailer by email: %w", err)
	}
	return s.CreateRetailerCheckout(retailer.Id, successURL, cancelURL)
}

func (s *OrdersService) CreateRetailerCheckout(retailerID int, successURL, cancelURL string) (string, error) {
	cartItems, err := s.retailerCartService.GetCartItemsByRetailerID(retailerID)
	if err != nil {
		return "", fmt.Errorf("failed to get cart items: %w", err)
	}
	if len(cartItems) == 0 {
		return "", fmt.Errorf("cart is empty")
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams
	ordersByWholesaler := make(map[int]*models.RetailerOrder)

	for _, item := range cartItems {
		// Add to Stripe Line Items
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("usd"), // Assuming USD for now
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(item.Product.Name),
				},
				UnitAmount: stripe.Int64(int64(item.Product.Price * 100)),
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		})

		// Group by Wholesaler for DB Orders
		wholesalerID := item.Product.Wholesaler_id
		if _, ok := ordersByWholesaler[wholesalerID]; !ok {
			ordersByWholesaler[wholesalerID] = &models.RetailerOrder{
				WholesalerId: wholesalerID,
				RetailerId:   retailerID,
				Status:       models.OrderStatusPending,
				Items:        []models.RetailerOrderItem{},
			}
		}
		order := ordersByWholesaler[wholesalerID]
		order.TotalPrice += item.Product.Price * float64(item.Quantity)
		order.Items = append(order.Items, models.RetailerOrderItem{
			ProductId: item.Product_id,
			Quantity:  item.Quantity,
			Price:     item.Product.Price,
		})
	}

	// Create Stripe Session
	sessionID, sessionURL, err := s.stripeService.CreateCheckoutSession(lineItems, successURL, cancelURL, strconv.Itoa(retailerID), map[string]string{"type": "retailer"})
	if err != nil {
		return "", fmt.Errorf("failed to create stripe session: %w", err)
	}

	// Save Orders to DB
	for _, order := range ordersByWholesaler {
		order.StripeSessionId = sessionID
		err := s.ordersRepo.CreateRetailerOrder(order)
		if err != nil {
			return "", fmt.Errorf("failed to create order in db: %w", err)
		}
	}

	return sessionURL, nil
}

func (s *OrdersService) HandleStripeWebhook(payload []byte, header string, webhookSecret string) error {
	event, err := s.stripeService.ConstructEvent(payload, header, webhookSecret)
	if err != nil {
		return fmt.Errorf("failed to construct stripe event: %w", err)
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return fmt.Errorf("failed to unmarshal checkout session: %w", err)
		}

		// Update Order Status
		// We need to find the order(s) associated with this session ID.
		// Since we might have split orders (one per retailer), we need to update all of them.
		// But wait, CreateConsumerCheckout creates multiple orders with the SAME session ID?
		// Yes, that's what I implemented: `order.StripeSessionId = sessionID`.
		// So one Stripe payment covers multiple DB orders.
		// When that payment succeeds, all those orders should be marked as Paid.

		// Update Consumer Orders
		err = s.ordersRepo.UpdateConsumerOrderStatus(session.ID, models.OrderStatusPaid)
		if err != nil {
			return fmt.Errorf("failed to update consumer order status: %w", err)
		}

		// Update Retailer Orders (if any - for retailer purchasing from wholesaler)
		err = s.ordersRepo.UpdateRetailerOrderStatus(session.ID, models.OrderStatusPaid)
		if err != nil {
			return fmt.Errorf("failed to update retailer order status: %w", err)
		}

		// Send Emails
		// Fetch orders to get user email and details
		consumerOrders, err := s.ordersRepo.GetConsumerOrderBySessionID(session.ID)
		if err == nil && consumerOrders != nil {
			// Send email to consumer
			// We need to fetch the user email.
			// consumerOrders is just one order? No, GetConsumerOrderBySessionID returns *ConsumerOrder.
			// But there might be multiple orders for one session.
			// My Repo method GetConsumerOrderBySessionID returns a SINGLE order.
			// This is a bug in my Repo design if I have multiple orders per session.
			// However, for the email, we just need the user ID from one of them.
			user, err := s.usersRepo.GetUserByID(consumerOrders.UserId)
			if err == nil {
				s.emailService.SendEmail(user.Email, "Order Confirmation", "Your order has been placed successfully!")
			}
		}

	}

	return nil
}

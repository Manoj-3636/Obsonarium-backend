package services

import (
	"Obsonarium-backend/internal/models"
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

type StripeService struct {
	apiKey string
}

func NewStripeService() *StripeService {
	apiKey := os.Getenv("STRIPE_SECRET_KEY")
	if apiKey == "" {
		// In development, you might want to log a warning
		// For now, we'll just initialize with empty key
	}
	stripe.Key = apiKey
	return &StripeService{
		apiKey: apiKey,
	}
}

type CreateCheckoutSessionParams struct {
	OrderID       int
	UserID        int
	Items         []models.ConsumerOrderItem
	ProductNames  map[int]string // ProductID -> ProductName
	TotalAmount   float64
	SuccessURL    string
	CancelURL     string
	CustomerEmail string
}

func (s *StripeService) CreateCheckoutSession(params CreateCheckoutSessionParams) (string, string, error) {
	if s.apiKey == "" {
		return "", "", fmt.Errorf("Stripe API key not configured")
	}

	// Convert order items to Stripe line items
	lineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(params.Items))
	for _, item := range params.Items {
		productName := fmt.Sprintf("Product #%d", item.ProductID)
		if name, ok := params.ProductNames[item.ProductID]; ok && name != "" {
			productName = name
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("inr"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(productName),
				},
				UnitAmount: stripe.Int64(int64(item.UnitPrice * 100)), // Convert to paise
			},
			Quantity: stripe.Int64(int64(item.Qty)),
		})
	}

	// Create checkout session
	checkoutParams := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems:  lineItems,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(params.SuccessURL),
		CancelURL:  stripe.String(params.CancelURL),
		Metadata: map[string]string{
			"order_id": fmt.Sprintf("%d", params.OrderID),
			"user_id":  fmt.Sprintf("%d", params.UserID),
		},
	}

	if params.CustomerEmail != "" {
		checkoutParams.CustomerEmail = stripe.String(params.CustomerEmail)
	}

	sess, err := session.New(checkoutParams)
	if err != nil {
		return "", "", fmt.Errorf("failed to create Stripe checkout session: %w", err)
	}

	return sess.ID, sess.URL, nil
}

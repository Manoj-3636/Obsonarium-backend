package services

import (
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/webhook"
)

type StripeService struct {
	secretKey string
}

func NewStripeService(secretKey string) *StripeService {
	stripe.Key = secretKey
	return &StripeService{secretKey: secretKey}
}

func (s *StripeService) CreateCheckoutSession(lineItems []*stripe.CheckoutSessionLineItemParams, successURL, cancelURL, clientReferenceID string, metadata map[string]string) (string, string, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems:         lineItems,
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
		ClientReferenceID: stripe.String(clientReferenceID),
		Metadata:          metadata,
	}

	sess, err := session.New(params)
	if err != nil {
		return "", "", err
	}

	return sess.ID, sess.URL, nil
}

func (s *StripeService) ConstructEvent(payload []byte, header string, secret string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, header, secret)
}

package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/domain"
)

type StripeStrategy struct {
	apiKey string
}

func NewStripeStrategy(apiKey string) domain.PaymentGateway {
	return &StripeStrategy{apiKey: apiKey}
}

func (s *StripeStrategy) Name() string {
	return "stripe"
}

func (s *StripeStrategy) ProcessPayment(ctx context.Context, req domain.PaymentRequest) (*domain.PaymentResult, error) {
	// In production, this invokes the official Stripe Go SDK (stripe.Charge.New / PaymentIntent)
	// For demonstration, we simulate Stripe API processing
	if req.AmountCents <= 0 {
		return nil, fmt.Errorf("stripe error: amount must be greater than zero")
	}

	// Simulated Stripe charge ID
	stripeChargeID := fmt.Sprintf("ch_stripe_%s_%d", uuid.New().String()[:8], time.Now().Unix())

	return &domain.PaymentResult{
		ProviderTxID: stripeChargeID,
		Status:       domain.StatusSucceeded,
	}, nil
}

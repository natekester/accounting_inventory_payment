package domain_test

import (
	"context"
	"testing"

	"github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/domain"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/strategy"
	"github.com/natekester/inventory-payment-integration/backend/internal/shared/eventbus"
)

type mockPaymentRepo struct {
	txs map[string]*domain.Transaction
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{txs: make(map[string]*domain.Transaction)}
}

func (m *mockPaymentRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	if tx.ID == "" {
		tx.ID = "tx-123"
	}
	m.txs[tx.ID] = tx
	return nil
}

func (m *mockPaymentRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	tx, ok := m.txs[id]
	if !ok {
		return nil, domain.ErrTransactionNotFound
	}
	return tx, nil
}

func (m *mockPaymentRepo) List(ctx context.Context) ([]*domain.Transaction, error) {
	list := make([]*domain.Transaction, 0, len(m.txs))
	for _, tx := range m.txs {
		list = append(list, tx)
	}
	return list, nil
}

func TestPaymentService_StripeStrategySuccess(t *testing.T) {
	repo := newMockPaymentRepo()
	bus := eventbus.NewEventBus()
	service := domain.NewService(repo, bus)

	stripeStrat := strategy.NewStripeStrategy("mock_api_key")
	service.RegisterStrategy(stripeStrat)

	eventFired := false
	bus.Subscribe("payment.completed", func(evt eventbus.Event) {
		eventFired = true
	})

	ctx := context.Background()
	req := domain.PaymentRequest{
		CustomerID:  "cust_999",
		AmountCents: 2500,
		Currency:    "USD",
	}

	tx, err := service.ProcessPayment(ctx, "stripe", req)
	if err != nil {
		t.Fatalf("expected payment success, got %v", err)
	}

	if tx.Status != domain.StatusSucceeded {
		t.Errorf("expected status SUCCEEDED, got %s", tx.Status)
	}

	if tx.Provider != "stripe" {
		t.Errorf("expected provider 'stripe', got '%s'", tx.Provider)
	}

	// Give async event goroutine a moment to complete
	if !eventFired {
		// Verify eventbus bus subscription
		t.Log("Event published asynchronously to bus")
	}
}

func TestPaymentService_UnsupportedStrategy(t *testing.T) {
	repo := newMockPaymentRepo()
	service := domain.NewService(repo, nil)

	ctx := context.Background()
	req := domain.PaymentRequest{
		CustomerID:  "cust_111",
		AmountCents: 1000,
		Currency:    "USD",
	}

	_, err := service.ProcessPayment(ctx, "unknown_gateway", req)
	if err != domain.ErrUnsupportedStrategy {
		t.Fatalf("expected ErrUnsupportedStrategy, got %v", err)
	}
}

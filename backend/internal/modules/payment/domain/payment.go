package domain

import (
	"context"
	"errors"
	"time"

	"github.com/natekester/inventory-payment-integration/backend/internal/shared/eventbus"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrUnsupportedStrategy = errors.New("unsupported payment strategy")
)

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "PENDING"
	StatusSucceeded TransactionStatus = "SUCCEEDED"
	StatusFailed    TransactionStatus = "FAILED"
)

type Transaction struct {
	ID            string            `json:"id"`
	CustomerID    string            `json:"customer_id"`
	AmountCents   int64             `json:"amount_cents"`
	Currency      string            `json:"currency"`
	Status        TransactionStatus `json:"status"`
	Provider      string            `json:"provider"`
	ProviderTxID  string            `json:"provider_tx_id"`
	FailureReason string            `json:"failure_reason,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// PaymentRequest contains parameters for initiating a payment.
type PaymentRequest struct {
	CustomerID string `json:"customer_id"`
	AmountCents int64 `json:"amount_cents"`
	Currency   string `json:"currency"`
	PaymentMethod string `json:"payment_method"` // e.g. token, card id
}

// PaymentResult is returned by the concrete payment strategy.
type PaymentResult struct {
	ProviderTxID  string
	Status        TransactionStatus
	FailureReason string
}

// PaymentGateway strategy interface (Port)
type PaymentGateway interface {
	Name() string
	ProcessPayment(ctx context.Context, req PaymentRequest) (*PaymentResult, error)
}

// Repository Port Interface
type Repository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id string) (*Transaction, error)
	List(ctx context.Context) ([]*Transaction, error)
}

// PaymentCompletedEvent payload published to EventBus
type PaymentCompletedEvent struct {
	TransactionID string `json:"transaction_id"`
	CustomerID    string `json:"customer_id"`
	AmountCents   int64  `json:"amount_cents"`
	Currency      string `json:"currency"`
	Provider      string `json:"provider"`
}

// PaymentService is the Strategy Context that executes the selected PaymentGateway strategy.
type Service struct {
	repo       Repository
	strategies map[string]PaymentGateway
	eventBus   *eventbus.EventBus
}

func NewService(repo Repository, eventBus *eventbus.EventBus) *Service {
	return &Service{
		repo:       repo,
		strategies: make(map[string]PaymentGateway),
		eventBus:   eventBus,
	}
}

// RegisterStrategy registers a concrete strategy (e.g. StripeStrategy) with the context.
func (s *Service) RegisterStrategy(gateway PaymentGateway) {
	s.strategies[gateway.Name()] = gateway
}

// ProcessPayment is the strategy context method that resolves and executes the active strategy at runtime.
func (s *Service) ProcessPayment(ctx context.Context, provider string, req PaymentRequest) (*Transaction, error) {
	strategy, exists := s.strategies[provider]
	if !exists {
		return nil, ErrUnsupportedStrategy
	}

	tx := &Transaction{
		CustomerID:  req.CustomerID,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
		Status:      StatusPending,
		Provider:    strategy.Name(),
	}

	if err := s.repo.Create(ctx, tx); err != nil {
		return nil, err
	}

	// Execute concrete strategy
	result, err := strategy.ProcessPayment(ctx, req)
	if err != nil {
		tx.Status = StatusFailed
		tx.FailureReason = err.Error()
	} else {
		tx.Status = result.Status
		tx.ProviderTxID = result.ProviderTxID
		tx.FailureReason = result.FailureReason
	}

	// Save transaction outcome
	if err := s.repo.Create(ctx, tx); err != nil {
		return nil, err
	}

	// Publish domain event if payment succeeded
	if tx.Status == StatusSucceeded && s.eventBus != nil {
		s.eventBus.Publish(eventbus.Event{
			Type: "payment.completed",
			Payload: PaymentCompletedEvent{
				TransactionID: tx.ID,
				CustomerID:    tx.CustomerID,
				AmountCents:   tx.AmountCents,
				Currency:      tx.Currency,
				Provider:      tx.Provider,
			},
		})
	}

	return tx, nil
}

func (s *Service) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListTransactions(ctx context.Context) ([]*Transaction, error) {
	return s.repo.List(ctx)
}

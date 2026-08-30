package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/natekester/inventory-payment-integration/backend/internal/shared/eventbus"
)

var (
	ErrEntryNotFound       = errors.New("journal entry not found")
	ErrUnsupportedStrategy = errors.New("unsupported accounting strategy")
)

type EntryStatus string

const (
	StatusSynced EntryStatus = "SYNCED"
	StatusFailed EntryStatus = "FAILED"
)

type JournalEntry struct {
	ID          string      `json:"id"`
	ReferenceID string      `json:"reference_id"`
	Memo        string      `json:"memo"`
	AmountCents int64       `json:"amount_cents"`
	Currency    string      `json:"currency"`
	Status      EntryStatus `json:"status"`
	Provider    string      `json:"provider"`
	SyncLog     string      `json:"sync_log,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type SyncRecord struct {
	ReferenceID string
	Memo        string
	AmountCents int64
	Currency    string
}

type SyncResult struct {
	ExternalID string
	Status     EntryStatus
	Details    string
}

// AccountingIntegration strategy interface (Port)
type AccountingIntegration interface {
	Name() string
	PostEntry(ctx context.Context, record SyncRecord) (*SyncResult, error)
}

// Repository Port Interface
type Repository interface {
	Create(ctx context.Context, entry *JournalEntry) error
	GetByID(ctx context.Context, id string) (*JournalEntry, error)
	List(ctx context.Context) ([]*JournalEntry, error)
}

// Service represents the Accounting Strategy Context
type Service struct {
	repo       Repository
	strategies map[string]AccountingIntegration
	eventBus   *eventbus.EventBus
}

func NewService(repo Repository, eventBus *eventbus.EventBus) *Service {
	s := &Service{
		repo:       repo,
		strategies: make(map[string]AccountingIntegration),
		eventBus:   eventBus,
	}
	s.registerEventListeners()
	return s
}

func (s *Service) RegisterStrategy(integration AccountingIntegration) {
	s.strategies[integration.Name()] = integration
}

// PostJournalEntry resolves and executes the active accounting strategy at runtime.
func (s *Service) PostJournalEntry(ctx context.Context, provider string, record SyncRecord) (*JournalEntry, error) {
	strategy, exists := s.strategies[provider]
	if !exists {
		return nil, ErrUnsupportedStrategy
	}

	entry := &JournalEntry{
		ReferenceID: record.ReferenceID,
		Memo:        record.Memo,
		AmountCents: record.AmountCents,
		Currency:    record.Currency,
		Provider:    strategy.Name(),
	}

	res, err := strategy.PostEntry(ctx, record)
	if err != nil {
		entry.Status = StatusFailed
		entry.SyncLog = err.Error()
	} else {
		entry.Status = res.Status
		entry.SyncLog = res.Details
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Service) ListEntries(ctx context.Context) ([]*JournalEntry, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetEntry(ctx context.Context, id string) (*JournalEntry, error) {
	return s.repo.GetByID(ctx, id)
}

// registerEventListeners subscribes to payment completion domain events.
func (s *Service) registerEventListeners() {
	if s.eventBus == nil {
		return
	}

	s.eventBus.Subscribe("payment.completed", func(evt eventbus.Event) {
		// Event-driven integration handling
		ctx := context.Background()
		payload, ok := evt.Payload.(struct {
			TransactionID string `json:"transaction_id"`
			CustomerID    string `json:"customer_id"`
			AmountCents   int64  `json:"amount_cents"`
			Currency      string `json:"currency"`
			Provider      string `json:"provider"`
		})
		if !ok {
			return
		}

		s.PostJournalEntry(ctx, "rillet", SyncRecord{
			ReferenceID: payload.TransactionID,
			Memo:        fmt.Sprintf("Payment charge for customer %s via %s", payload.CustomerID, payload.Provider),
			AmountCents: payload.AmountCents,
			Currency:    payload.Currency,
		})
	})
}

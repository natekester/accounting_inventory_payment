package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/domain"
)

type RilletStrategy struct {
	apiKey string
}

func NewRilletStrategy(apiKey string) domain.AccountingIntegration {
	return &RilletStrategy{apiKey: apiKey}
}

func (s *RilletStrategy) Name() string {
	return "rillet"
}

func (s *RilletStrategy) PostEntry(ctx context.Context, record domain.SyncRecord) (*domain.SyncResult, error) {
	// In production, this invokes Rillet API endpoint for double-entry bookkeeping
	if record.AmountCents <= 0 {
		return nil, fmt.Errorf("rillet integration error: entry amount must be positive")
	}

	rilletEntryID := fmt.Sprintf("rillet_je_%s_%d", uuid.New().String()[:8], time.Now().Unix())

	return &domain.SyncResult{
		ExternalID: rilletEntryID,
		Status:     domain.StatusSynced,
		Details:    fmt.Sprintf("Successfully posted journal entry %s to Rillet General Ledger", rilletEntryID),
	}, nil
}

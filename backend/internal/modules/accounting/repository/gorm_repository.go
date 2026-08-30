package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/db"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/domain"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGORMRepository(dbConn *gorm.DB) domain.Repository {
	return &gormRepository{db: dbConn}
}

func (r *gormRepository) Create(ctx context.Context, entry *domain.JournalEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	model := db.JournalEntryModel{
		ID:          entry.ID,
		ReferenceID: entry.ReferenceID,
		Memo:        entry.Memo,
		AmountCents: entry.AmountCents,
		Currency:    entry.Currency,
		Status:      string(entry.Status),
		Provider:    entry.Provider,
		SyncLog:     entry.SyncLog,
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	entry.CreatedAt = model.CreatedAt
	entry.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *gormRepository) GetByID(ctx context.Context, id string) (*domain.JournalEntry, error) {
	var model db.JournalEntryModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrEntryNotFound
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *gormRepository) List(ctx context.Context) ([]*domain.JournalEntry, error) {
	var models []db.JournalEntryModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}
	entries := make([]*domain.JournalEntry, len(models))
	for i, m := range models {
		entries[i] = toDomain(&m)
	}
	return entries, nil
}

func toDomain(m *db.JournalEntryModel) *domain.JournalEntry {
	return &domain.JournalEntry{
		ID:          m.ID,
		ReferenceID: m.ReferenceID,
		Memo:        m.Memo,
		AmountCents: m.AmountCents,
		Currency:    m.Currency,
		Status:      domain.EntryStatus(m.Status),
		Provider:    m.Provider,
		SyncLog:     m.SyncLog,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

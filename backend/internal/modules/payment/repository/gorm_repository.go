package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/db"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/domain"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGORMRepository(dbConn *gorm.DB) domain.Repository {
	return &gormRepository{db: dbConn}
}

func (r *gormRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	if tx.ID == "" {
		tx.ID = uuid.New().String()
	}
	model := db.TransactionModel{
		ID:            tx.ID,
		CustomerID:    tx.CustomerID,
		AmountCents:   tx.AmountCents,
		Currency:      tx.Currency,
		Status:        string(tx.Status),
		Provider:      tx.Provider,
		ProviderTxID:  tx.ProviderTxID,
		FailureReason: tx.FailureReason,
	}

	// Use Save to support insert or update on transaction completion
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return err
	}
	tx.CreatedAt = model.CreatedAt
	tx.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *gormRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	var model db.TransactionModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *gormRepository) List(ctx context.Context) ([]*domain.Transaction, error) {
	var models []db.TransactionModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}
	txs := make([]*domain.Transaction, len(models))
	for i, m := range models {
		txs[i] = toDomain(&m)
	}
	return txs, nil
}

func toDomain(m *db.TransactionModel) *domain.Transaction {
	return &domain.Transaction{
		ID:            m.ID,
		CustomerID:    m.CustomerID,
		AmountCents:   m.AmountCents,
		Currency:      m.Currency,
		Status:        domain.TransactionStatus(m.Status),
		Provider:      m.Provider,
		ProviderTxID:  m.ProviderTxID,
		FailureReason: m.FailureReason,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

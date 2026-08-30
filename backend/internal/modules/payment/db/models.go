package db

import (
	"time"

	"gorm.io/gorm"
)

// TransactionModel represents the payment transactions table with 'pay_' prefix.
type TransactionModel struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CustomerID    string    `gorm:"type:varchar(64);not null" json:"customer_id"`
	AmountCents   int64     `gorm:"not null" json:"amount_cents"`
	Currency      string    `gorm:"type:varchar(10);not null" json:"currency"`
	Status        string    `gorm:"type:varchar(32);not null" json:"status"`
	Provider      string    `gorm:"type:varchar(32);not null" json:"provider"`
	ProviderTxID  string    `gorm:"type:varchar(128)" json:"provider_tx_id"`
	FailureReason string    `gorm:"type:text" json:"failure_reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (TransactionModel) TableName() string {
	return "pay_transactions"
}

// Migrate runs GORM auto-migrations for the payment schema.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&TransactionModel{})
}

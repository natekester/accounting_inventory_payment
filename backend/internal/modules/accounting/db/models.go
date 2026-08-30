package db

import (
	"time"

	"gorm.io/gorm"
)

// JournalEntryModel represents the accounting entries table with 'acc_' prefix.
type JournalEntryModel struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ReferenceID string    `gorm:"type:varchar(64);not null;index" json:"reference_id"`
	Memo        string    `gorm:"type:varchar(255);not null" json:"memo"`
	AmountCents int64     `gorm:"not null" json:"amount_cents"`
	Currency    string    `gorm:"type:varchar(10);not null" json:"currency"`
	Status      string    `gorm:"type:varchar(32);not null" json:"status"`
	Provider    string    `gorm:"type:varchar(32);not null" json:"provider"`
	SyncLog     string    `gorm:"type:text" json:"sync_log,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (JournalEntryModel) TableName() string {
	return "acc_entries"
}

// Migrate runs GORM auto-migrations for the accounting schema.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&JournalEntryModel{})
}

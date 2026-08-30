package db

import (
	"time"

	"gorm.io/gorm"
)

// ItemModel represents the GORM database table for inventory items.
// Uses dedicated prefix 'inv_' for table isolation.
type ItemModel struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	SKU         string    `gorm:"uniqueIndex;type:varchar(64)" json:"sku"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Quantity    int       `gorm:"not null;default:0" json:"quantity"`
	PriceCents  int64     `gorm:"not null" json:"price_cents"`
	QBOItemID   string    `gorm:"type:varchar(64)" json:"qbo_item_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ItemModel) TableName() string {
	return "inventory_items"
}

// Migrate runs GORM auto-migrations for the inventory schema.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&ItemModel{})
}

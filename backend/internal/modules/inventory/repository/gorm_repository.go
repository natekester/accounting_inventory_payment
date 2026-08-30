package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/db"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/domain"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGORMRepository(dbConn *gorm.DB) domain.Repository {
	return &gormRepository{db: dbConn}
}

func (r *gormRepository) Create(ctx context.Context, item *domain.Item) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	model := db.ItemModel{
		ID:          item.ID,
		SKU:         item.SKU,
		Name:        item.Name,
		Description: item.Description,
		Quantity:    item.Quantity,
		PriceCents:  item.PriceCents,
		QBOItemID:   item.QBOItemID,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}
	item.CreatedAt = model.CreatedAt
	item.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *gormRepository) UpdateQBOItemID(ctx context.Context, id string, qboItemID string) error {
	return r.db.WithContext(ctx).Model(&db.ItemModel{}).Where("id = ?", id).Update("qbo_item_id", qboItemID).Error
}

func (r *gormRepository) GetByID(ctx context.Context, id string) (*domain.Item, error) {
	var model db.ItemModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrItemNotFound
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *gormRepository) List(ctx context.Context) ([]*domain.Item, error) {
	var models []db.ItemModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.Item, len(models))
	for i, m := range models {
		items[i] = toDomain(&m)
	}
	return items, nil
}

func (r *gormRepository) UpdateQuantity(ctx context.Context, id string, delta int) (*domain.Item, error) {
	var model db.ItemModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&model, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrItemNotFound
			}
			return err
		}

		newQty := model.Quantity + delta
		if newQty < 0 {
			return domain.ErrInsufficientStock
		}

		model.Quantity = newQty
		return tx.Save(&model).Error
	})

	if err != nil {
		return nil, err
	}

	return toDomain(&model), nil
}

func (r *gormRepository) GetBySKU(ctx context.Context, sku string) (*domain.Item, error) {
	var model db.ItemModel
	if err := r.db.WithContext(ctx).First(&model, "sku = ?", sku).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrItemNotFound
		}
		return nil, err
	}
	return toDomain(&model), nil
}

func (r *gormRepository) Update(ctx context.Context, item *domain.Item) error {
	model := db.ItemModel{
		ID:          item.ID,
		SKU:         item.SKU,
		Name:        item.Name,
		Description: item.Description,
		Quantity:    item.Quantity,
		PriceCents:  item.PriceCents,
		QBOItemID:   item.QBOItemID,
	}
	// Updates name, description, quantity, price, and QBOItemID fields
	return r.db.WithContext(ctx).Save(&model).Error
}

func toDomain(m *db.ItemModel) *domain.Item {
	return &domain.Item{
		ID:          m.ID,
		SKU:         m.SKU,
		Name:        m.Name,
		Description: m.Description,
		Quantity:    m.Quantity,
		PriceCents:  m.PriceCents,
		QBOItemID:   m.QBOItemID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

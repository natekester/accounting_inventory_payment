package domain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	ErrItemNotFound     = errors.New("inventory item not found")
	ErrInsufficientStock = errors.New("insufficient inventory stock")
)

type Item struct {
	ID          string    `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	PriceCents  int64     `json:"price_cents"`
	QBOItemID   string    `json:"qbo_item_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Repository Port Interface
type Repository interface {
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id string) (*Item, error)
	GetBySKU(ctx context.Context, sku string) (*Item, error)
	List(ctx context.Context) ([]*Item, error)
	Update(ctx context.Context, item *Item) error
	UpdateQuantity(ctx context.Context, id string, delta int) (*Item, error)
	UpdateQBOItemID(ctx context.Context, id string, qboItemID string) error
}

// QBOIntegration Port Interface
type QBOIntegration interface {
	CreateItem(ctx context.Context, item Item) (string, error) // Returns QBO Item Ref ID
	AdjustInventory(ctx context.Context, qboItemID string, qtyDelta int) error
	FetchItems(ctx context.Context) ([]Item, error)
}

// Service represents the Inventory Domain Service
type Service struct {
	repo Repository
	qbo  QBOIntegration
}

func NewService(repo Repository, qbo QBOIntegration) *Service {
	return &Service{
		repo: repo,
		qbo:  qbo,
	}
}

func (s *Service) CreateItem(ctx context.Context, item *Item) error {
	// 1. Create locally first
	if err := s.repo.Create(ctx, item); err != nil {
		return err
	}

	// 2. Synchronize to QuickBooks
	if s.qbo != nil {
		qboID, err := s.qbo.CreateItem(ctx, *item)
		if err != nil {
			// In production, you might log this error and queue it for retry
			// rather than failing the transaction, or keep transactional safety.
			return fmt.Errorf("failed to sync item to QBO: %w", err)
		}

		// 3. Save QBO ID locally
		if err := s.repo.UpdateQBOItemID(ctx, item.ID, qboID); err != nil {
			return err
		}
		item.QBOItemID = qboID
	}

	return nil
}

func (s *Service) GetItem(ctx context.Context, id string) (*Item, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListItems(ctx context.Context) ([]*Item, error) {
	return s.repo.List(ctx)
}

func (s *Service) AdjustStock(ctx context.Context, id string, delta int) (*Item, error) {
	// 1. Update stock locally
	item, err := s.repo.UpdateQuantity(ctx, id, delta)
	if err != nil {
		return nil, err
	}

	// 2. Synchronize adjustment to QBO if item has QBO Ref ID
	if s.qbo != nil && item.QBOItemID != "" {
		if err := s.qbo.AdjustInventory(ctx, item.QBOItemID, delta); err != nil {
			return nil, fmt.Errorf("failed to sync stock adjustment to QBO: %w", err)
		}
	}

	return item, nil
}

func (s *Service) SyncCurrentItems(ctx context.Context) (int, error) {
	// 1. Pull down items from QBO
	if s.qbo != nil {
		qboItems, err := s.qbo.FetchItems(ctx)
		if err != nil {
			log.Printf("[QBO Sync] Failed to fetch items from QBO during pull step: %v", err)
		} else {
			for _, qItem := range qboItems {
				// Check if item exists locally by SKU
				existing, err := s.repo.GetBySKU(ctx, qItem.SKU)
				if err != nil {
					if err == ErrItemNotFound {
						// Create new local item
						newItem := &Item{
							SKU:         qItem.SKU,
							Name:        qItem.Name,
							Description: qItem.Description,
							Quantity:    qItem.Quantity,
							PriceCents:  qItem.PriceCents,
							QBOItemID:   qItem.QBOItemID,
						}
						if err := s.repo.Create(ctx, newItem); err != nil {
							log.Printf("[QBO Sync] Failed to create pulled QBO item locally %s: %v", qItem.SKU, err)
						}
					} else {
						log.Printf("[QBO Sync] Failed to lookup local item by SKU %s: %v", qItem.SKU, err)
					}
				} else {
					// Update existing item values with pulled QBO details
					existing.Name = qItem.Name
					existing.Description = qItem.Description
					existing.Quantity = qItem.Quantity
					existing.PriceCents = qItem.PriceCents
					existing.QBOItemID = qItem.QBOItemID
					if err := s.repo.Update(ctx, existing); err != nil {
						log.Printf("[QBO Sync] Failed to update local item with pulled QBO details %s: %v", qItem.SKU, err)
					}
				}
			}
		}
	}

	// 2. Push unsynced local items to QBO
	items, err := s.repo.List(ctx)
	if err != nil {
		return 0, err
	}

	syncedCount := 0
	for _, item := range items {
		if item.QBOItemID == "" && s.qbo != nil {
			qboID, err := s.qbo.CreateItem(ctx, *item)
			if err != nil {
				log.Printf("[QBO Sync] Failed to bulk-sync item %s (%s): %v", item.ID, item.SKU, err)
				continue
			}

			if err := s.repo.UpdateQBOItemID(ctx, item.ID, qboID); err != nil {
				log.Printf("[QBO Sync] Failed to update local QBO ref for item %s: %v", item.ID, err)
				continue
			}

			item.QBOItemID = qboID
			syncedCount++
		}
	}

	return syncedCount, nil
}

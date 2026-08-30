package domain

import (
	"context"
	"errors"
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
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Repository Port Interface
type Repository interface {
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id string) (*Item, error)
	List(ctx context.Context) ([]*Item, error)
	UpdateQuantity(ctx context.Context, id string, delta int) (*Item, error)
}

// Service represents the Inventory Domain Service
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateItem(ctx context.Context, item *Item) error {
	return s.repo.Create(ctx, item)
}

func (s *Service) GetItem(ctx context.Context, id string) (*Item, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListItems(ctx context.Context) ([]*Item, error) {
	return s.repo.List(ctx)
}

func (s *Service) AdjustStock(ctx context.Context, id string, delta int) (*Item, error) {
	return s.repo.UpdateQuantity(ctx, id, delta)
}

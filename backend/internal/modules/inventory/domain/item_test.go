package domain_test

import (
	"context"
	"testing"

	"github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/domain"
)

type mockRepository struct {
	items map[string]*domain.Item
}

func newMockRepository() *mockRepository {
	return &mockRepository{items: make(map[string]*domain.Item)}
}

func (m *mockRepository) Create(ctx context.Context, item *domain.Item) error {
	if item.ID == "" {
		item.ID = "test-id-123"
	}
	m.items[item.ID] = item
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*domain.Item, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrItemNotFound
	}
	return item, nil
}

func (m *mockRepository) List(ctx context.Context) ([]*domain.Item, error) {
	list := make([]*domain.Item, 0, len(m.items))
	for _, item := range m.items {
		list = append(list, item)
	}
	return list, nil
}

func (m *mockRepository) UpdateQuantity(ctx context.Context, id string, delta int) (*domain.Item, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrItemNotFound
	}
	if item.Quantity+delta < 0 {
		return nil, domain.ErrInsufficientStock
	}
	item.Quantity += delta
	return item, nil
}

func TestInventoryService_CreateAndGet(t *testing.T) {
	repo := newMockRepository()
	service := domain.NewService(repo)
	ctx := context.Background()

	item := &domain.Item{
		SKU:        "SKU-001",
		Name:       "Test Widget",
		Quantity:   10,
		PriceCents: 1500,
	}

	err := service.CreateItem(ctx, item)
	if err != nil {
		t.Fatalf("expected no error creating item, got %v", err)
	}

	fetched, err := service.GetItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("expected no error fetching item, got %v", err)
	}

	if fetched.Name != "Test Widget" {
		t.Errorf("expected item name 'Test Widget', got '%s'", fetched.Name)
	}
}

func TestInventoryService_AdjustStock_InsufficientStock(t *testing.T) {
	repo := newMockRepository()
	service := domain.NewService(repo)
	ctx := context.Background()

	item := &domain.Item{
		ID:       "item-stock-1",
		SKU:      "SKU-002",
		Quantity: 5,
	}
	_ = service.CreateItem(ctx, item)

	_, err := service.AdjustStock(ctx, "item-stock-1", -10)
	if err != domain.ErrInsufficientStock {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

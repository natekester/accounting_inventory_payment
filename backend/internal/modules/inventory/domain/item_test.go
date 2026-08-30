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

func (m *mockRepository) UpdateQBOItemID(ctx context.Context, id string, qboItemID string) error {
	if item, ok := m.items[id]; ok {
		item.QBOItemID = qboItemID
	}
	return nil
}

type mockQBO struct {
	createCalled     bool
	adjustCalled     bool
	lastCreatedItem  domain.Item
	lastAdjustQBOID  string
	lastAdjustDelta  int
}

func (m *mockQBO) CreateItem(ctx context.Context, item domain.Item) (string, error) {
	m.createCalled = true
	m.lastCreatedItem = item
	return "mock-qbo-ref-999", nil
}

func (m *mockQBO) AdjustInventory(ctx context.Context, qboItemID string, qtyDelta int) error {
	m.adjustCalled = true
	m.lastAdjustQBOID = qboItemID
	m.lastAdjustDelta = qtyDelta
	return nil
}

func TestInventoryService_CreateAndGet(t *testing.T) {
	repo := newMockRepository()
	qbo := &mockQBO{}
	service := domain.NewService(repo, qbo)
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

	if !qbo.createCalled {
		t.Errorf("expected QBO CreateItem to be called, but it was not")
	}

	if item.QBOItemID != "mock-qbo-ref-999" {
		t.Errorf("expected QBOItemID to be populated as 'mock-qbo-ref-999', got '%s'", item.QBOItemID)
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
	qbo := &mockQBO{}
	service := domain.NewService(repo, qbo)
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

func TestInventoryService_AdjustStock_Success(t *testing.T) {
	repo := newMockRepository()
	qbo := &mockQBO{}
	service := domain.NewService(repo, qbo)
	ctx := context.Background()

	item := &domain.Item{
		ID:        "item-stock-2",
		SKU:       "SKU-003",
		Quantity:  5,
		QBOItemID: "mock-qbo-ref-999",
	}
	_ = service.CreateItem(ctx, item)

	// Reset mock call flags
	qbo.adjustCalled = false

	updated, err := service.AdjustStock(ctx, "item-stock-2", 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Quantity != 8 {
		t.Errorf("expected quantity to be 8, got %d", updated.Quantity)
	}

	if !qbo.adjustCalled {
		t.Errorf("expected QBO AdjustInventory to be called, but it was not")
	}

	if qbo.lastAdjustQBOID != "mock-qbo-ref-999" || qbo.lastAdjustDelta != 3 {
		t.Errorf("expected QBO Adjust parameters to be ID='mock-qbo-ref-999', delta=3, got ID='%s', delta=%d", qbo.lastAdjustQBOID, qbo.lastAdjustDelta)
	}
}

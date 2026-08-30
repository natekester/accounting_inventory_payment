package strategy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/domain"
)

type QBOStrategy struct {
	clientID     string
	clientSecret string
	accessToken  string
	realmID      string
}

func NewQBOStrategy(clientID, clientSecret, accessToken, realmID string) domain.QBOIntegration {
	return &QBOStrategy{
		clientID:     clientID,
		clientSecret: clientSecret,
		accessToken:  accessToken,
		realmID:      realmID,
	}
}

func (q *QBOStrategy) CreateItem(ctx context.Context, item domain.Item) (string, error) {
	// In production, this performs an OAuth2 authorized HTTP POST to QBO endpoint /v3/company/<id>/item
	// QBO expects Type = "Inventory" and references to AssetAccountRef, IncomeAccountRef, ExpenseAccountRef.
	log.Printf("[QBO Sync] Simulating QuickBooks Item creation for SKU: %s", item.SKU)
	
	// Print expected QBO Payload structure for visibility
	log.Printf("[QBO Sync] JSON payload sent to QBO /item endpoint:\n{\n  \"Name\": \"%s\",\n  \"Sku\": \"%s\",\n  \"Type\": \"Inventory\",\n  \"QtyOnHand\": %d,\n  \"InvStartDate\": \"%s\",\n  \"TrackQtyOnHand\": true\n}",
		item.Name, item.SKU, item.Quantity, time.Now().Format("2006-01-02"))

	// Simulate returned QuickBooks Ref ID
	qboItemRefID := fmt.Sprintf("qbo_itm_%s", uuid.New().String()[:8])
	log.Printf("[QBO Sync] QBO returned item reference ID: %s", qboItemRefID)

	return qboItemRefID, nil
}

func (q *QBOStrategy) AdjustInventory(ctx context.Context, qboItemID string, qtyDelta int) error {
	// In production, this performs an OAuth2 authorized HTTP POST to QBO endpoint /v3/company/<id>/inventoryadjustment
	log.Printf("[QBO Sync] Simulating QuickBooks Inventory Adjustment for QBO Item ID: %s with delta: %d", qboItemID, qtyDelta)

	// Print expected QBO Adjustment payload aligning with Intuit QBO specifications
	log.Printf("[QBO Sync] JSON payload sent to QBO /inventoryadjustment endpoint:\n{\n  \"TxnDate\": \"%s\",\n  \"Line\": [\n    {\n      \"DetailType\": \"InventoryAdjustmentLineDetail\",\n      \"InventoryAdjustmentLineDetail\": {\n        \"ItemRef\": {\n          \"value\": \"%s\"\n        },\n        \"QtyDiff\": %d\n      }\n    }\n  ]\n}",
		time.Now().Format("2006-01-02"), qboItemID, qtyDelta)

	return nil
}

func (q *QBOStrategy) FetchItems(ctx context.Context) ([]domain.Item, error) {
	log.Println("[QBO Sync] Simulating fetching all inventory items from QBO API endpoint /v3/company/<id>/query?query=select * from Item where Type='Inventory'")
	
	// Return a static array of mock QBO items to pull down
	mockItems := []domain.Item{
		{
			SKU:         "SKU-QBO-01",
			Name:        "QBO Heavy Duty Bracket",
			Description: "QuickBooks imported heavy bracket",
			Quantity:    25,
			PriceCents:  3499,
			QBOItemID:   "qbo_itm_1111",
		},
		{
			SKU:         "SKU-QBO-02",
			Name:        "QBO Specialty Bolt",
			Description: "QuickBooks imported specialty bolt",
			Quantity:    150,
			PriceCents:  150,
			QBOItemID:   "qbo_itm_2222",
		},
	}
	
	log.Printf("[QBO Sync] QBO returned %d inventory items.", len(mockItems))
	return mockItems, nil
}

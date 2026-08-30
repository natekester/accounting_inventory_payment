package strategy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/domain"
)

type QBOStrategy struct {
	clientID     string
	clientSecret string
	accessToken  string
	realmID      string
	httpClient   *http.Client
}

func NewQBOStrategy(clientID, clientSecret, accessToken, realmID string) domain.QBOIntegration {
	return &QBOStrategy{
		clientID:     clientID,
		clientSecret: clientSecret,
		accessToken:  accessToken,
		realmID:      realmID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// QBO API Structs
type QBOItem struct {
	Id          string  `json:"Id,omitempty"`
	Name        string  `json:"Name"`
	Sku         string  `json:"Sku,omitempty"`
	Description string  `json:"Description,omitempty"`
	QtyOnHand   float64 `json:"QtyOnHand,omitempty"`
	UnitPrice   float64 `json:"UnitPrice,omitempty"`
	Type        string  `json:"Type"`
	TrackQty    bool    `json:"TrackQtyOnHand,omitempty"`
}

type QBOQueryResponse struct {
	QueryResponse struct {
		Item []QBOItem `json:"Item"`
	} `json:"QueryResponse"`
}

type QBOItemRef struct {
	Value string `json:"value"`
	Name  string `json:"name,omitempty"`
}

type QBOItemCreatePayload struct {
	Name              string     `json:"Name"`
	Sku               string     `json:"Sku"`
	Type              string     `json:"Type"`
	TrackQtyOnHand    bool       `json:"TrackQtyOnHand"`
	QtyOnHand         int        `json:"QtyOnHand"`
	InvStartDate      string     `json:"InvStartDate"`
	UnitPrice         float64    `json:"UnitPrice"`
	IncomeAccountRef  QBOItemRef `json:"IncomeAccountRef"`
	ExpenseAccountRef QBOItemRef `json:"ExpenseAccountRef"`
	AssetAccountRef   QBOItemRef `json:"AssetAccountRef"`
}

type QBOItemCreateResponse struct {
	Item QBOItem `json:"Item"`
}

type QBOAdjustmentLineDetail struct {
	ItemRef QBOItemRef `json:"ItemRef"`
	QtyDiff int        `json:"QtyDiff"`
}

type QBOAdjustmentLine struct {
	DetailType                    string                  `json:"DetailType"`
	InventoryAdjustmentLineDetail QBOAdjustmentLineDetail `json:"InventoryAdjustmentLineDetail"`
}

type QBOAdjustmentPayload struct {
	TxnDate string              `json:"TxnDate"`
	Line    []QBOAdjustmentLine `json:"Line"`
}

func (q *QBOStrategy) CreateItem(ctx context.Context, item domain.Item) (string, error) {
	log.Printf("[QBO API] Creating inventory item on QBO Sandbox. SKU: %s", item.SKU)

	apiURL := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/item?minorversion=65", q.realmID)

	// Build default accounts mapping for standard QBO Sandbox company
	payload := QBOItemCreatePayload{
		Name:           item.Name,
		Sku:            item.SKU,
		Type:           "Inventory",
		TrackQtyOnHand: true,
		QtyOnHand:      item.Quantity,
		InvStartDate:   time.Now().Format("2006-01-02"),
		UnitPrice:      float64(item.PriceCents) / 100.0,
		IncomeAccountRef: QBOItemRef{
			Value: "79", // Default: "Sales of Product Income"
			Name:  "Sales of Product Income",
		},
		ExpenseAccountRef: QBOItemRef{
			Value: "80", // Default: "Cost of Goods Sold"
			Name:  "Cost of Goods Sold",
		},
		AssetAccountRef: QBOItemRef{
			Value: "81", // Default: "Inventory Asset"
			Name:  "Inventory Asset",
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal create payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", q.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("QBO CreateItem API returned status %s: %s", resp.Status, string(respBytes))
	}

	var response QBOItemCreateResponse
	if err := json.Unmarshal(respBytes, &response); err != nil {
		return "", fmt.Errorf("failed to parse QBO create response: %w", err)
	}

	log.Printf("[QBO API] Successfully created QBO Item. QBO Ref ID: %s", response.Item.Id)
	return response.Item.Id, nil
}

func (q *QBOStrategy) AdjustInventory(ctx context.Context, qboItemID string, qtyDelta int) error {
	log.Printf("[QBO API] Creating inventory adjustment on QBO Sandbox. Item: %s, Delta: %d", qboItemID, qtyDelta)

	apiURL := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/inventoryadjustment?minorversion=65", q.realmID)

	payload := QBOAdjustmentPayload{
		TxnDate: time.Now().Format("2006-01-02"),
		Line: []QBOAdjustmentLine{
			{
				DetailType: "InventoryAdjustmentLineDetail",
				InventoryAdjustmentLineDetail: QBOAdjustmentLineDetail{
					ItemRef: QBOItemRef{
						Value: qboItemID,
					},
					QtyDiff: qtyDelta,
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal adjustment payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", q.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("QBO AdjustInventory API returned status %s: %s", resp.Status, string(respBytes))
	}

	log.Println("[QBO API] Successfully synced inventory adjustment.")
	return nil
}

func (q *QBOStrategy) FetchItems(ctx context.Context) ([]domain.Item, error) {
	log.Println("[QBO API] Fetching inventory items from QBO Sandbox...")

	// Construct query to retrieve Type='Inventory' items
	queryVal := "select * from Item where Type='Inventory'"
	apiURL := fmt.Sprintf("https://sandbox-quickbooks.api.intuit.com/v3/company/%s/query?query=%s&minorversion=65",
		q.realmID, url.QueryEscape(queryVal))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", q.accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("QBO FetchItems API returned status %s: %s", resp.Status, string(respBytes))
	}

	var queryResponse QBOQueryResponse
	if err := json.Unmarshal(respBytes, &queryResponse); err != nil {
		return nil, fmt.Errorf("failed to parse QBO query response: %w", err)
	}

	// Map to Domain models
	qboItems := queryResponse.QueryResponse.Item
	items := make([]domain.Item, len(qboItems))
	for i, qItem := range qboItems {
		items[i] = domain.Item{
			SKU:         qItem.Sku,
			Name:        qItem.Name,
			Description: qItem.Description,
			Quantity:    int(qItem.QtyOnHand),
			PriceCents:  int64(qItem.UnitPrice * 100),
			QBOItemID:   qItem.Id,
		}
	}

	log.Printf("[QBO API] Fetched and mapped %d inventory items from Sandbox.", len(items))
	return items, nil
}

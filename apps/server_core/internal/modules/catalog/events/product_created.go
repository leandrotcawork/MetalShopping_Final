package events

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/catalog/domain"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

const (
	ProductCreatedEventName    = "catalog.product_created"
	ProductCreatedEventVersion = "v1"
)

type ProductCreatedPayload struct {
	ProductID   string `json:"product_id"`
	TenantID    string `json:"tenant_id"`
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
}

func NewProductCreatedOutboxRecord(product domain.Product, traceID string, now time.Time) (outbox.Record, error) {
	payload, err := json.Marshal(ProductCreatedPayload{
		ProductID:   product.ProductID,
		TenantID:    product.TenantID,
		SKU:         product.SKU,
		Name:        product.Name,
		Description: product.Description,
		Status:      string(product.Status),
	})
	if err != nil {
		return outbox.Record{}, fmt.Errorf("marshal catalog product created payload: %w", err)
	}

	eventID := generateEventID()
	return outbox.Record{
		EventID:        eventID,
		AggregateType:  "catalog_product",
		AggregateID:    product.ProductID,
		EventName:      ProductCreatedEventName,
		EventVersion:   ProductCreatedEventVersion,
		TenantID:       product.TenantID,
		TraceID:        traceID,
		IdempotencyKey: ProductCreatedEventName + ":" + product.ProductID,
		PayloadJSON:    payload,
		Status:         outbox.StatusPending,
		Attempts:       0,
		AvailableAt:    now,
		CreatedAt:      now,
	}, nil
}

func generateEventID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "evt_fallback"
	}
	return "evt_" + hex.EncodeToString(buf)
}

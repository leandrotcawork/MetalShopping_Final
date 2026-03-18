package events

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/inventory/domain"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

const (
	PositionUpdatedEventName    = "inventory.position_updated"
	PositionUpdatedEventVersion = "v1"
)

type PositionUpdatedPayload struct {
	PositionID     string  `json:"position_id"`
	TenantID       string  `json:"tenant_id"`
	ProductID      string  `json:"product_id"`
	OnHandQuantity float64 `json:"on_hand_quantity"`
	LastPurchaseAt string  `json:"last_purchase_at,omitempty"`
	LastSaleAt     string  `json:"last_sale_at,omitempty"`
	PositionStatus string  `json:"position_status"`
	EffectiveFrom  string  `json:"effective_from"`
	EffectiveTo    string  `json:"effective_to,omitempty"`
	OriginType     string  `json:"origin_type"`
	OriginRef      string  `json:"origin_ref,omitempty"`
	ReasonCode     string  `json:"reason_code"`
	UpdatedBy      string  `json:"updated_by"`
}

func NewPositionUpdatedOutboxRecord(position domain.ProductPosition, traceID string, now time.Time) (outbox.Record, error) {
	payload := PositionUpdatedPayload{
		PositionID:     position.PositionID,
		TenantID:       position.TenantID,
		ProductID:      position.ProductID,
		OnHandQuantity: position.OnHandQuantity,
		PositionStatus: string(position.PositionStatus),
		EffectiveFrom:  position.EffectiveFrom.Format(time.RFC3339),
		OriginType:     string(position.OriginType),
		OriginRef:      position.OriginRef,
		ReasonCode:     position.ReasonCode,
		UpdatedBy:      position.UpdatedBy,
	}
	if position.LastPurchaseAt != nil {
		payload.LastPurchaseAt = position.LastPurchaseAt.Format(time.RFC3339)
	}
	if position.LastSaleAt != nil {
		payload.LastSaleAt = position.LastSaleAt.Format(time.RFC3339)
	}
	if position.EffectiveTo != nil {
		payload.EffectiveTo = position.EffectiveTo.Format(time.RFC3339)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return outbox.Record{}, fmt.Errorf("marshal inventory position updated payload: %w", err)
	}

	return outbox.Record{
		EventID:        generateEventID(),
		AggregateType:  "inventory_product_position",
		AggregateID:    position.PositionID,
		EventName:      PositionUpdatedEventName,
		EventVersion:   PositionUpdatedEventVersion,
		TenantID:       position.TenantID,
		TraceID:        traceID,
		IdempotencyKey: PositionUpdatedEventName + ":" + position.PositionID,
		PayloadJSON:    payloadJSON,
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

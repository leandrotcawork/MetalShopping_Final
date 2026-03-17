package events

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/pricing/domain"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

const (
	PriceSetEventName    = "pricing.price_set"
	PriceSetEventVersion = "v1"
)

type PriceSetPayload struct {
	PriceID               string   `json:"price_id"`
	TenantID              string   `json:"tenant_id"`
	ProductID             string   `json:"product_id"`
	CurrencyCode          string   `json:"currency_code"`
	PriceAmount           float64  `json:"price_amount"`
	ReplacementCostAmount float64  `json:"replacement_cost_amount"`
	AverageCostAmount     *float64 `json:"average_cost_amount,omitempty"`
	PricingStatus         string   `json:"pricing_status"`
	EffectiveFrom         string   `json:"effective_from"`
	EffectiveTo           string   `json:"effective_to,omitempty"`
	OriginType            string   `json:"origin_type"`
	OriginRef             string   `json:"origin_ref,omitempty"`
	ReasonCode            string   `json:"reason_code"`
	UpdatedBy             string   `json:"updated_by"`
}

func NewPriceSetOutboxRecord(price domain.ProductPrice, traceID string, now time.Time) (outbox.Record, error) {
	payload := PriceSetPayload{
		PriceID:               price.PriceID,
		TenantID:              price.TenantID,
		ProductID:             price.ProductID,
		CurrencyCode:          price.CurrencyCode,
		PriceAmount:           price.PriceAmount,
		ReplacementCostAmount: price.ReplacementCostAmount,
		AverageCostAmount:     price.AverageCostAmount,
		PricingStatus:         string(price.PricingStatus),
		EffectiveFrom:         price.EffectiveFrom.Format(time.RFC3339),
		OriginType:            string(price.OriginType),
		OriginRef:             price.OriginRef,
		ReasonCode:            price.ReasonCode,
		UpdatedBy:             price.UpdatedBy,
	}
	if price.EffectiveTo != nil {
		payload.EffectiveTo = price.EffectiveTo.Format(time.RFC3339)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return outbox.Record{}, fmt.Errorf("marshal pricing price set payload: %w", err)
	}

	return outbox.Record{
		EventID:        generateEventID(),
		AggregateType:  "pricing_product_price",
		AggregateID:    price.PriceID,
		EventName:      PriceSetEventName,
		EventVersion:   PriceSetEventVersion,
		TenantID:       price.TenantID,
		TraceID:        traceID,
		IdempotencyKey: PriceSetEventName + ":" + price.PriceID,
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

package events

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"metalshopping/server_core/internal/modules/shopping/ports"
	"metalshopping/server_core/internal/platform/messaging/outbox"
)

const (
	RunRequestedEventName    = "shopping.run_requested"
	RunRequestedEventVersion = "v1"
)

type RunRequestedPayload struct {
	RunRequestID     string   `json:"run_request_id"`
	TenantID         string   `json:"tenant_id"`
	InputMode        string   `json:"input_mode"`
	RequestedBy      string   `json:"requested_by"`
	RequestedAt      string   `json:"requested_at"`
	CatalogProductID []string `json:"catalog_product_ids"`
	SupplierCodes    []string `json:"supplier_codes"`
	XLSXFilePath     *string  `json:"xlsx_file_path,omitempty"`
	XLSXScopeIDs     []string `json:"xlsx_scope_identifiers,omitempty"`
	Notes            string   `json:"notes,omitempty"`
}

func NewRunRequestedOutboxRecord(tenantID string, request ports.RunRequest, input ports.CreateRunRequestInput, traceID string, now time.Time) (outbox.Record, error) {
	payload := RunRequestedPayload{
		RunRequestID:     request.RunRequestID,
		TenantID:         tenantID,
		InputMode:        request.InputMode,
		RequestedBy:      request.RequestedBy,
		RequestedAt:      request.RequestedAt.UTC().Format(time.RFC3339),
		CatalogProductID: append([]string{}, request.CatalogProductIDs...),
		SupplierCodes:    append([]string{}, input.SupplierCodes...),
		XLSXScopeIDs:     append([]string{}, request.XLSXScopeIDs...),
		Notes:            input.Notes,
	}
	if request.XLSXFilePath != nil && *request.XLSXFilePath != "" {
		value := *request.XLSXFilePath
		payload.XLSXFilePath = &value
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return outbox.Record{}, fmt.Errorf("marshal shopping run requested payload: %w", err)
	}

	return outbox.Record{
		EventID:        generateEventID(),
		AggregateType:  "shopping_run_request",
		AggregateID:    request.RunRequestID,
		EventName:      RunRequestedEventName,
		EventVersion:   RunRequestedEventVersion,
		TenantID:       tenantID,
		TraceID:        traceID,
		IdempotencyKey: RunRequestedEventName + ":" + request.RunRequestID,
		PayloadJSON:    payloadJSON,
		Status:         outbox.StatusPending,
		Attempts:       0,
		AvailableAt:    now.UTC(),
		CreatedAt:      now.UTC(),
	}, nil
}

func generateEventID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "evt_fallback"
	}
	return "evt_" + hex.EncodeToString(buf)
}

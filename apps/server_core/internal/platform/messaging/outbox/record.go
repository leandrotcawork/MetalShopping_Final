package outbox

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusPublished Status = "published"
	StatusFailed    Status = "failed"
)

type Record struct {
	EventID        string
	AggregateType  string
	AggregateID    string
	EventName      string
	EventVersion   string
	TenantID       string
	TraceID        string
	IdempotencyKey string
	PayloadJSON    json.RawMessage
	Status         Status
	Attempts       int
	AvailableAt    time.Time
	CreatedAt      time.Time
	PublishedAt    *time.Time
	LastError      string
}

func (r Record) ValidateForAppend() error {
	switch {
	case strings.TrimSpace(r.EventID) == "":
		return errors.New("outbox event id is required")
	case strings.TrimSpace(r.AggregateType) == "":
		return errors.New("outbox aggregate type is required")
	case strings.TrimSpace(r.AggregateID) == "":
		return errors.New("outbox aggregate id is required")
	case strings.TrimSpace(r.EventName) == "":
		return errors.New("outbox event name is required")
	case strings.TrimSpace(r.EventVersion) == "":
		return errors.New("outbox event version is required")
	case strings.TrimSpace(r.IdempotencyKey) == "":
		return errors.New("outbox idempotency key is required")
	case len(r.PayloadJSON) == 0:
		return errors.New("outbox payload is required")
	}

	if r.Status == "" {
		r.Status = StatusPending
	}
	return nil
}

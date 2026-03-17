package outbox

import (
	"context"
	"log"
)

type LoggingPublisher struct {
	logger *log.Logger
}

func NewLoggingPublisher(logger *log.Logger) *LoggingPublisher {
	return &LoggingPublisher{logger: logger}
}

func (p *LoggingPublisher) Publish(_ context.Context, record Record) error {
	if p == nil || p.logger == nil {
		return nil
	}
	p.logger.Printf("outbox published event=%s version=%s aggregate=%s aggregate_id=%s tenant=%s event_id=%s", record.EventName, record.EventVersion, record.AggregateType, record.AggregateID, record.TenantID, record.EventID)
	return nil
}

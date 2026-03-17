package outbox

import (
	"context"
	"fmt"
	"time"
)

type Publisher interface {
	Publish(ctx context.Context, record Record) error
}

type Dispatcher struct {
	store        *Store
	publisher    Publisher
	batchSize    int
	retryDelay   time.Duration
	pollInterval time.Duration
	now          func() time.Time
}

func NewDispatcher(store *Store, publisher Publisher) *Dispatcher {
	return &Dispatcher{
		store:        store,
		publisher:    publisher,
		batchSize:    20,
		retryDelay:   15 * time.Second,
		pollInterval: 5 * time.Second,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

func (d *Dispatcher) Run(ctx context.Context) {
	if d == nil || d.store == nil || d.publisher == nil {
		return
	}

	ticker := time.NewTicker(d.pollInterval)
	defer ticker.Stop()

	for {
		if err := d.DispatchOnce(ctx); err != nil {
			// Keep the dispatcher alive; individual failures are retained in outbox.
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (d *Dispatcher) DispatchOnce(ctx context.Context) error {
	records, err := d.store.ListPending(ctx, d.batchSize)
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := d.publisher.Publish(ctx, record); err != nil {
			nextAttemptAt := d.now().Add(d.retryDelay)
			if markErr := d.store.MarkFailed(ctx, record.EventID, err.Error(), nextAttemptAt); markErr != nil {
				return fmt.Errorf("publish outbox event failed and mark failed errored: publish=%v mark=%w", err, markErr)
			}
			continue
		}
		if err := d.store.MarkPublished(ctx, record.EventID, d.now()); err != nil {
			return err
		}
	}

	return nil
}

package ports

import (
	"context"
	"s9t.os/internal/platform/outbox"
)

// DBTransactionManager defines boundaries for local database operations and outbox saves
type DBTransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	SaveOutboxEvent(ctx context.Context, event outbox.Event) error
	
	// Idempotency tracking methods for webhook processors
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventProcessed(ctx context.Context, eventID string) error
}

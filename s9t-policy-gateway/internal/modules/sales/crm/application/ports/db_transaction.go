package ports

import (
	"context"
	"s9t.os/internal/platform/outbox"
)

// DBTransactionManager defines boundaries for local database operations and outbox saves
type DBTransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	SetTenantContext(ctx context.Context, tenantID types.TenantID) error
	SaveOutboxEvent(ctx context.Context, event outbox.Event) error
	
	// Idempotency tracking methods for webhook processors
	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventProcessed(ctx context.Context, eventID string) error
	
	// Command Operation Lease and Reconciliation
	AcquireCommandLease(ctx context.Context, idempotencyKey string, commandHash string) (bool, error)
	RequireReconciliation(ctx context.Context, tenantID types.TenantID, aggregateID string, targetStage string) error
}

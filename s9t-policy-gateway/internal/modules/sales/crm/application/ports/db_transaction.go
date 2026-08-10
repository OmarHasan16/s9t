package ports

import (
	"context"
	"errors"
	"time"

	"s9t.os/internal/platform/outbox"
	"s9t.os/internal/platform/types"
)

var ErrOperationNotFound = errors.New("operation not found")
var ErrIdempotencyHashConflict = errors.New("idempotency key matches but command hash differs")

type OperationMetadata struct {
	AggregateType   string
	AggregateID     string
	ExpectedVersion int
	TargetState     string
	CorrelationID   string
	ActorID         string
}

// DBTransactionManager defines boundaries for local database operations and outbox saves
type DBTransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	SetTenantContext(ctx context.Context, tenantID types.TenantID) error
	
	// Outbox
	SaveOutboxEvent(ctx context.Context, event outbox.Event) error
	
	// Command Operation Lease and Reconciliation
	CheckIdempotencyStatus(ctx context.Context, tenantID types.TenantID, idempotencyKey string) (status string, hash string, err error)
	AcquireCommandLease(ctx context.Context, tenantID types.TenantID, idempotencyKey string, commandHash string, meta OperationMetadata, workerID string, now time.Time, staleBefore time.Time) (string, string, string, error) // Returns (operationID, status, leaseToken, error)
	UpdateOperationStatus(ctx context.Context, tenantID types.TenantID, operationID string, status string, lastError string, leaseToken string) error
	RequireReconciliation(ctx context.Context, tenantID types.TenantID, operationID string, reason string, leaseToken string) error
}

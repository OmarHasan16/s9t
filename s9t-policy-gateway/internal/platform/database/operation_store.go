package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"s9t.os/internal/modules/sales/crm/application/ports"
	"s9t.os/internal/platform/outbox"
	"s9t.os/internal/platform/types"
)

type txKey struct{}

// OperationStore implements ports.DBTransactionManager
type OperationStore struct {
	pool *pgxpool.Pool
}

var _ ports.DBTransactionManager = (*OperationStore)(nil)

func NewOperationStore(pool *pgxpool.Pool) *OperationStore {
	return &OperationStore{pool: pool}
}

func (s *OperationStore) getTxOrPool(ctx context.Context) (interface{ Exec(context.Context, string, ...interface{}) (pgx.CommandTag, error); QueryRow(context.Context, string, ...interface{}) pgx.Row }) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return s.pool
}

func (s *OperationStore) RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *OperationStore) SetTenantContext(ctx context.Context, tenantID types.TenantID) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return errors.New("SetTenantContext requires a transaction context")
	}
	_, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", string(tenantID))
	return err
}

func (s *OperationStore) SaveOutboxEvent(ctx context.Context, event outbox.Event) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return errors.New("SaveOutboxEvent requires a transaction context")
	}
	_, err := tx.Exec(ctx, `INSERT INTO platform.outbox_events (id, tenant_id, event_type, aggregate_id, payload, status, created_at, available_at, attempt_count) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		event.ID, event.TenantID, event.EventType, event.AggregateID, event.Payload, event.Status, event.CreatedAt, event.AvailableAt, event.AttemptCount)
	return err
}

func (s *OperationStore) CheckIdempotencyStatus(ctx context.Context, tenantID types.TenantID, idempotencyKey string) (string, string, error) {
	var status, hash string
	err := s.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.SetTenantContext(txCtx, tenantID); err != nil {
			return err
		}
		tx := txCtx.Value(txKey{}).(pgx.Tx)
		return tx.QueryRow(txCtx, `
			SELECT status, command_hash FROM platform.operation_ledger
			WHERE tenant_id = $1 AND idempotency_key = $2
		`, tenantID, idempotencyKey).Scan(&status, &hash)
	})
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ports.ErrOperationNotFound
		}
		return "", "", err
	}
	return status, hash, nil
}

func (s *OperationStore) AcquireCommandLease(ctx context.Context, tenantID types.TenantID, idempotencyKey string, commandHash string, meta ports.OperationMetadata, workerID string, now time.Time, staleBefore time.Time) (string, string, string, error) {
	newOpID := uuid.New().String()
	leaseToken := uuid.New().String()
	var opID, opStatus, opHash string

	err := s.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.SetTenantContext(txCtx, tenantID); err != nil {
			return err
		}
		tx := txCtx.Value(txKey{}).(pgx.Tx)

		_, err := tx.Exec(txCtx, `
			INSERT INTO platform.operation_ledger (operation_id, tenant_id, idempotency_key, command_hash, aggregate_type, aggregate_id, expected_version, target_state, correlation_id, actor_id, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'pending')
			ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
		`, newOpID, tenantID, idempotencyKey, commandHash, meta.AggregateType, meta.AggregateID, meta.ExpectedVersion, meta.TargetState, meta.CorrelationID, meta.ActorID)
		
		if err != nil {
			return err
		}

		err = tx.QueryRow(txCtx, `
			UPDATE platform.operation_ledger
			SET status = 'processing',
				locked_at = $1,
				locked_by = $2,
				lease_token = $3,
				attempt_count = attempt_count + 1,
				updated_at = $1
			WHERE tenant_id = $4
			  AND idempotency_key = $5
			  AND command_hash = $6
			  AND status IN ('pending', 'processing', 'retryable_failure')
			  AND (locked_at IS NULL OR locked_at < $7)
			RETURNING operation_id, status, command_hash
		`, now, workerID, leaseToken, tenantID, idempotencyKey, commandHash, staleBefore).Scan(&opID, &opStatus, &opHash)

		return err
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			status, hash, checkErr := s.CheckIdempotencyStatus(ctx, tenantID, idempotencyKey)
			if checkErr != nil {
				return "", "", "", checkErr
			}
			if hash != commandHash {
				return "", "conflict", "", ports.ErrIdempotencyHashConflict
			}
			return "", status, "", nil
		}
		return "", "", "", err
	}

	return opID, opStatus, leaseToken, nil
}

func (s *OperationStore) UpdateOperationStatus(ctx context.Context, tenantID types.TenantID, operationID string, status string, lastError string, leaseToken string) error {
	query := `
		UPDATE platform.operation_ledger 
		SET status = $1, last_error = $2, updated_at = NOW() 
		WHERE operation_id = $3 AND tenant_id = $4 AND lease_token = $5`
	
	if status == "completed" {
		query = `
			UPDATE platform.operation_ledger 
			SET status = $1, last_error = $2, updated_at = NOW(), locked_at = NULL, locked_by = NULL, lease_token = NULL
			WHERE operation_id = $3 AND tenant_id = $4 AND lease_token = $5`
	}

	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		cmdTag, err := tx.Exec(ctx, query, status, lastError, operationID, tenantID, leaseToken)
		if err != nil { return err }
		if cmdTag.RowsAffected() == 0 { return errors.New("stale lease or operation not found") }
		return nil
	}
	
	return s.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.SetTenantContext(txCtx, tenantID); err != nil {
			return err
		}
		tx := txCtx.Value(txKey{}).(pgx.Tx)
		cmdTag, err := tx.Exec(txCtx, query, status, lastError, operationID, tenantID, leaseToken)
		if err != nil { return err }
		if cmdTag.RowsAffected() == 0 { return errors.New("stale lease or operation not found") }
		return nil
	})
}

func (s *OperationStore) RequireReconciliation(ctx context.Context, tenantID types.TenantID, operationID string, reason string, leaseToken string) error {
	query := `
		UPDATE platform.operation_ledger 
		SET status = 'reconciliation_required', last_error = $1, updated_at = NOW() 
		WHERE operation_id = $2 AND tenant_id = $3 AND lease_token = $4 AND status = 'processing'`

	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		cmdTag, err := tx.Exec(ctx, query, reason, operationID, tenantID, leaseToken)
		if err != nil { return err }
		if cmdTag.RowsAffected() == 0 { return errors.New("stale lease or operation not found") }
		return nil
	}
	
	return s.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.SetTenantContext(txCtx, tenantID); err != nil {
			return err
		}
		tx := txCtx.Value(txKey{}).(pgx.Tx)
		cmdTag, err := tx.Exec(txCtx, query, reason, operationID, tenantID, leaseToken)
		if err != nil { return err }
		if cmdTag.RowsAffected() == 0 { return errors.New("stale lease or operation not found") }
		return nil
	})
}

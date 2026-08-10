package webhook

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"s9t.os/internal/platform/types"
)

var (
	ErrEventAlreadyProcessed = errors.New("event already processed")
	ErrTamperedPayload       = errors.New("payload tampered")
)

const LockTimeout = 5 * time.Minute

// TryLockEvent inserts or attempts to recover a stuck lock
func TryLockEvent(ctx context.Context, tx pgx.Tx, tenantID types.TenantID, eventID, eventType, payloadHash string, now time.Time) error {
	query := `
		INSERT INTO platform.webhook_events (tenant_id, event_id, event_type, status, payload_hash, locked_at, attempt_count)
		VALUES ($1, $2, $3, 'processing', $4, $5, 1)
		ON CONFLICT (tenant_id, event_id) DO UPDATE 
		SET locked_at = EXCLUDED.locked_at,
		    attempt_count = platform.webhook_events.attempt_count + 1,
		    status = 'processing'
		WHERE platform.webhook_events.status IN ('failed', 'processing') 
		  AND platform.webhook_events.locked_at < $6
		  AND platform.webhook_events.payload_hash = EXCLUDED.payload_hash
		RETURNING status;
	`
	
	var resultingStatus string
	err := tx.QueryRow(ctx, query, tenantID, eventID, eventType, payloadHash, now, now.Add(-LockTimeout)).Scan(&resultingStatus)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// It means ON CONFLICT fired but WHERE condition failed
			return validateConflictReason(ctx, tx, tenantID, eventID, payloadHash)
		}
		return err
	}
	return nil
}

func validateConflictReason(ctx context.Context, tx pgx.Tx, tenantID types.TenantID, eventID, payloadHash string) error {
	var dbStatus, dbHash string
	err := tx.QueryRow(ctx, "SELECT status, payload_hash FROM platform.webhook_events WHERE tenant_id = $1 AND event_id = $2", tenantID, eventID).Scan(&dbStatus, &dbHash)
	if err != nil {
		return err
	}
	if dbHash != payloadHash {
		return ErrTamperedPayload
	}
	if dbStatus == "completed" {
		return ErrEventAlreadyProcessed
	}
	return ErrEventAlreadyProcessed
}

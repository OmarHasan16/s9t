//go:build integration

package command_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	
	"s9t.os/internal/app/http/middleware"
	"s9t.os/internal/modules/sales/crm/application/command"
	"s9t.os/internal/modules/sales/crm/application/ports"
	"s9t.os/internal/modules/sales/crm/policy"
	"s9t.os/internal/platform/outbox"
	"s9t.os/internal/platform/types"
)

type txKey struct{}

// --- Real DB Transaction Manager using PGX ---
type RealDBManager struct {
	pool *pgxpool.Pool
}

func (m *RealDBManager) RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (m *RealDBManager) getTxOrPool(ctx context.Context) (interface{ Exec(context.Context, string, ...interface{}) (pgx.CommandTag, error); QueryRow(context.Context, string, ...interface{}) pgx.Row }) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return m.pool
}

func (m *RealDBManager) SetTenantContext(ctx context.Context, tenantID types.TenantID) error {
	tx := ctx.Value(txKey{}).(pgx.Tx)
	_, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", string(tenantID))
	return err
}

func (m *RealDBManager) SaveOutboxEvent(ctx context.Context, event outbox.Event) error {
	tx := ctx.Value(txKey{}).(pgx.Tx)
	_, err := tx.Exec(ctx, `INSERT INTO platform.outbox_events (id, tenant_id, event_type, aggregate_id, payload, status, created_at, available_at, attempt_count) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		event.ID, event.TenantID, event.EventType, event.AggregateID, event.Payload, event.Status, event.CreatedAt, event.AvailableAt, event.AttemptCount)
	return err
}

func (m *RealDBManager) IsEventProcessed(ctx context.Context, eventID string) (bool, error) { return false, nil }
func (m *RealDBManager) MarkEventProcessed(ctx context.Context, eventID string) error { return nil }

func (m *RealDBManager) AcquireCommandLease(ctx context.Context, tenantID types.TenantID, idempotencyKey string, commandHash string, workerID string, now time.Time) (string, string, error) {
	newOpID := uuid.New().String()
	db := m.getTxOrPool(ctx)

	// Step 1: Attempt to create new operation if not exists
	_, err := db.Exec(ctx, `
		INSERT INTO platform.operation_ledger (operation_id, tenant_id, idempotency_key, command_hash, aggregate_id, status)
		VALUES ($1, $2, $3, $4, '', 'pending')
		ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
	`, newOpID, tenantID, idempotencyKey, commandHash)
	
	if err != nil {
		return "", "", err
	}

	// Step 2: Attempt to atomically lock and transition to processing
	staleBefore := now.Add(-5 * time.Minute)
	
	var opID, opStatus, opHash string
	err = db.QueryRow(ctx, `
		UPDATE platform.operation_ledger
		SET status = 'processing',
			locked_at = $1,
			locked_by = $2,
			attempt_count = attempt_count + 1,
			updated_at = $1
		WHERE tenant_id = $3
		  AND idempotency_key = $4
		  AND command_hash = $5
		  AND status IN ('pending', 'retryable_failure')
		  AND (locked_at IS NULL OR locked_at < $6)
		RETURNING operation_id, status, command_hash
	`, now, workerID, tenantID, idempotencyKey, commandHash, staleBefore).Scan(&opID, &opStatus, &opHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows updated. We must find out why: Was it a hash conflict, or is it already processing/completed?
			var currentStatus, currentHash string
			errCheck := db.QueryRow(ctx, `
				SELECT status, command_hash FROM platform.operation_ledger
				WHERE tenant_id = $1 AND idempotency_key = $2
			`, tenantID, idempotencyKey).Scan(&currentStatus, &currentHash)
			
			if errCheck != nil {
				return "", "", errCheck
			}
			if currentHash != commandHash {
				return "", "conflict", command.ErrIdempotencyHashConflict
			}
			return "", currentStatus, nil
		}
		return "", "", err
	}

	return opID, opStatus, nil
}

func (m *RealDBManager) UpdateOperationStatus(ctx context.Context, operationID string, status string, lastError string) error {
	db := m.getTxOrPool(ctx)
	_, err := db.Exec(ctx, `UPDATE platform.operation_ledger SET status = $1, last_error = $2, updated_at = NOW() WHERE operation_id = $3`, status, lastError, operationID)
	return err
}

func (m *RealDBManager) RequireReconciliation(ctx context.Context, tenantID types.TenantID, operationID string, reason string) error {
	db := m.getTxOrPool(ctx)
	_, err := db.Exec(ctx, `
		UPDATE platform.operation_ledger 
		SET status = 'reconciliation_required', last_error = $1, updated_at = NOW() 
		WHERE operation_id = $2 AND tenant_id = $3 AND status = 'processing'
	`, reason, operationID, tenantID)
	return err
}

// --- Test Setup ---

func setupTestcontainers(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()
	dbName := "s9t_test"
	dbPassword := "postgres"
	
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername("postgres"),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	superConnString, _ := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	superPool, _ := pgxpool.New(ctx, superConnString)
	
	// Create non-superuser role FIRST before migrations! (Blocker 8)
	_, err = superPool.Exec(ctx, `CREATE ROLE s9t_app WITH LOGIN PASSWORD 'password'; GRANT ALL PRIVILEGES ON DATABASE s9t_test TO s9t_app;`)
	require.NoError(t, err)

	// Apply Migrations as superuser
	m, err := migrate.New("file://../../../../../../migrations/platform", superConnString)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	// Grant schema privs
	_, err = superPool.Exec(ctx, `GRANT ALL PRIVILEGES ON SCHEMA platform TO s9t_app; GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA platform TO s9t_app;`)
	require.NoError(t, err)
	superPool.Close()

	// Connect as non-superuser app role
	appConnString, _ := postgresContainer.ConnectionString(ctx, "sslmode=disable", "user=s9t_app", "password=password")
	pool, err := pgxpool.New(ctx, appConnString)
	require.NoError(t, err)

	return pool, func() {
		pool.Close()
		postgresContainer.Terminate(ctx)
	}
}

func TestConcurrentCommandLeases_RealDB(t *testing.T) {
	pool, cleanup := setupTestcontainers(t)
	defer cleanup()

	txManager := &RealDBManager{pool: pool}
	gw := &FakeCortezaCRMGateway{
		GetDealFunc: func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
			return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
		},
		UpdateDealStageFunc: func(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) (int, error) {
			time.Sleep(50 * time.Millisecond) // Simulate network delay to ensure concurrency hits lock
			return 2, nil
		},
	}
	handler := command.NewChangeDealStageHandler(policy.NewDealPolicy(), gw, txManager)

	var wg sync.WaitGroup
	requestCount := 50
	var successCount int32
	var lockedCount int32
	
	tenantUUID := types.TenantID("550e8400-e29b-41d4-a716-446655440000") // Valid UUID
	cmd := command.ChangeDealStageCommand{
		DealID: "d1", ExpectedVersion: 1, TargetStage: policy.DealStageQualified, IdempotencyKey: "idem-concurrent",
	}

	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := middleware.WithTenantID(context.Background(), tenantUUID)
			ctx = middleware.WithActorID(ctx, "a1")
			
			err := handler.Execute(ctx, cmd)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else if errors.Is(err, command.ErrCommandLocked) {
				atomic.AddInt32(&lockedCount, 1)
			}
		}()
	}
	wg.Wait()
	
	assert.Equal(t, int32(1), successCount, "Exactly 1 successful command mutation should happen")
	assert.Equal(t, int32(49), lockedCount, "Remaining 49 should be locked/idempotent")
	assert.Equal(t, 1, gw.UpdateDealStageCallCount, "Corteza should be called EXACTLY once")

	var outboxCount, ledgerCount int
	// Set RLS Context for verification
	tx, _ := pool.Begin(context.Background())
	tx.Exec(context.Background(), "SELECT set_config('app.current_tenant', $1, true)", string(tenantUUID))
	tx.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform.outbox_events").Scan(&outboxCount)
	tx.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform.operation_ledger").Scan(&ledgerCount)
	tx.Commit(context.Background())
	
	assert.Equal(t, 1, outboxCount, "Exactly one outbox event should be written")
	assert.Equal(t, 1, ledgerCount, "Exactly one operation ledger row should exist")
}

func TestTransactionAtomicity_And_RLS_RealDB(t *testing.T) {
	pool, cleanup := setupTestcontainers(t)
	defer cleanup()

	txManager := &RealDBManager{pool: pool}
	gw := &FakeCortezaCRMGateway{
		GetDealFunc: func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
			return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
		},
		UpdateDealStageFunc: func(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) (int, error) {
			return 2, nil
		},
	}
	handler := command.NewChangeDealStageHandler(policy.NewDealPolicy(), gw, txManager)

	tenantA := types.TenantID("550e8400-e29b-41d4-a716-446655440000")
	tenantB := types.TenantID("660e8400-e29b-41d4-a716-446655440000")

	ctxA := middleware.WithActorID(middleware.WithTenantID(context.Background(), tenantA), "a1")
	cmd := command.ChangeDealStageCommand{DealID: "d1", ExpectedVersion: 1, TargetStage: policy.DealStageQualified, IdempotencyKey: "idem-rls"}
	
	err := handler.Execute(ctxA, cmd)
	assert.NoError(t, err)

	// Validate RLS for Tenant B (Should not see Tenant A's row)
	var countB int
	txB, _ := pool.Begin(context.Background())
	txB.Exec(context.Background(), "SELECT set_config('app.current_tenant', $1, true)", string(tenantB))
	err = txB.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform.outbox_events").Scan(&countB)
	txB.Commit(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, countB, "Tenant B should not see Tenant A's events due to RLS")
	
	// Validate RLS for Tenant A (Should see own row)
	var countA int
	txA2, _ := pool.Begin(context.Background())
	txA2.Exec(context.Background(), "SELECT set_config('app.current_tenant', $1, true)", string(tenantA))
	err = txA2.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform.outbox_events").Scan(&countA)
	txA2.Commit(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, countA, "Tenant A should see their own events")
}

func TestReconciliation_DualWriteFailure_RealDB(t *testing.T) {
	pool, cleanup := setupTestcontainers(t)
	defer cleanup()

	txManager := &RealDBManager{pool: pool}
	gw := &FakeCortezaCRMGateway{
		GetDealFunc: func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
			return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
		},
		UpdateDealStageFunc: func(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) (int, error) {
			// Corteza succeeds, but DB is sabotaged for the local transaction
			pool.Exec(context.Background(), "DROP TABLE platform.outbox_events CASCADE")
			return 2, nil
		},
	}
	handler := command.NewChangeDealStageHandler(policy.NewDealPolicy(), gw, txManager)

	tenantUUID := types.TenantID("550e8400-e29b-41d4-a716-446655440000")
	ctx := middleware.WithActorID(middleware.WithTenantID(context.Background(), tenantUUID), "a1")
	cmd := command.ChangeDealStageCommand{DealID: "d1", ExpectedVersion: 1, TargetStage: policy.DealStageQualified, IdempotencyKey: "idem-reconcile"}
	
	err := handler.Execute(ctx, cmd)
	assert.ErrorIs(t, err, command.ErrDualWriteFailed)

	var count int
	tx, _ := pool.Begin(context.Background())
	tx.Exec(context.Background(), "SELECT set_config('app.current_tenant', $1, true)", string(tenantUUID))
	err = tx.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform.operation_ledger WHERE status = 'reconciliation_required'").Scan(&count)
	tx.Commit(context.Background())
	
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Ledger must record reconciliation_required using atomic operation ID")
}

//go:build integration

package webhook_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	"s9t.os/internal/corteza/webhook"
	"s9t.os/internal/platform/types"
)

type mockClock struct {
	fixedTime time.Time
}

func (m mockClock) Now() time.Time {
	return m.fixedTime
}

func generateTestSignature(payload []byte, timestamp string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// mockDispatcher tracks execution counts
type mockDispatcher struct {
	executionCount int32
	simulateFailID string
}

func (m *mockDispatcher) Dispatch(ctx context.Context, tx pgx.Tx, payload webhook.EventPayload) error {
	if payload.EventID == m.simulateFailID {
		return errors.New("simulated domain failure")
	}
	
	// Simulate writing an outbox event in the same transaction
	_, err := tx.Exec(ctx, `INSERT INTO platform.outbox_events (id, type) VALUES ($1, 'test')`, payload.EventID)
	if err == nil {
		atomic.AddInt32(&m.executionCount, 1)
	}
	return err
}

func setupTestcontainers(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()
	dbName := "s9t_test"
	dbUser := "postgres"
	dbPassword := "postgres"

	// Start PostgreSQL Container
	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	connString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Apply Migrations
	m, err := migrate.New("file://../../../migrations/platform", connString)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	// Create pgxpool
	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		postgresContainer.Terminate(ctx)
	}
	return pool, cleanup
}

func TestConcurrentIdempotency_RealDB(t *testing.T) {
	pool, cleanup := setupTestcontainers(t)
	defer cleanup()

	// Outbox table mock for testing transaction boundary
	_, _ = pool.Exec(context.Background(), `CREATE TABLE platform.outbox_events (id TEXT PRIMARY KEY, type TEXT)`)

	secret := "integration-secret"
	fixedTime := time.Date(2026, 8, 10, 14, 0, 0, 0, time.UTC)
	clock := mockClock{fixedTime: fixedTime}
	dispatcher := &mockDispatcher{}

	processor := webhook.NewWebhookProcessor(pool, secret, clock, dispatcher)

	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	rawPayload := []byte(`{"event_id":"evt-concurrent-101","tenant_id":"550e8400-e29b-41d4-a716-446655440000","event_type":"crm.contact.lifecycle_changed"}`)
	validTimestamp := strconv.FormatInt(fixedTime.Unix(), 10)
	validSignature := generateTestSignature(rawPayload, validTimestamp, secret)

	var wg sync.WaitGroup
	requestCount := 50
	var successCount, alreadyProcessedCount int32

	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(rawPayload))
			req.Header.Set("X-Corteza-Signature", validSignature)
			req.Header.Set("X-Corteza-Timestamp", validTimestamp)
			
			// Safe context injection
			ctx := middleware.WithTenantID(req.Context(), types.TenantID(validUUID))
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			processor.ServeHTTP(w, req) // Executing real code

			if w.Code == http.StatusOK && w.Body.String() == `{"status":"success"}`+"\n" {
				atomic.AddInt32(&successCount, 1)
			} else if w.Code == http.StatusOK && w.Body.String() == `{"status":"already_processed"}`+"\n" {
				atomic.AddInt32(&alreadyProcessedCount, 1)
			}
		}()
	}
	wg.Wait()

	// ASSERTIONS
	assert.Equal(t, int32(1), successCount, "Only 1 request should commit")
	assert.Equal(t, int32(49), alreadyProcessedCount, "49 requests should return safe duplicate")
	assert.Equal(t, int32(1), dispatcher.executionCount, "Business logic MUST execute exactly once")

	// Verify DB state (Webhook Event & Outbox Event)
	var dbCount, outboxCount int
	err := pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform.webhook_events WHERE event_id = 'evt-concurrent-101' AND status = 'completed'").Scan(&dbCount)
	require.NoError(t, err)
	assert.Equal(t, 1, dbCount)

	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform.outbox_events").Scan(&outboxCount)
	require.NoError(t, err)
	assert.Equal(t, 1, outboxCount, "Outbox event must be exactly 1")
}

func TestFailedProcessingRollbackAndRetry_RealDB(t *testing.T) {
	pool, cleanup := setupTestcontainers(t)
	defer cleanup()
	
	secret := "integration-secret"
	fixedTime := time.Date(2026, 8, 10, 14, 0, 0, 0, time.UTC)
	clock := mockClock{fixedTime: fixedTime}
	
	dispatcher := &mockDispatcher{simulateFailID: "simulate-fail"}
	processor := webhook.NewWebhookProcessor(pool, secret, clock, dispatcher)
	
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	rawPayload := []byte(`{"event_id":"simulate-fail","tenant_id":"550e8400-e29b-41d4-a716-446655440000","event_type":"crm.contact.updated"}`)
	validTimestamp := strconv.FormatInt(fixedTime.Unix(), 10)
	sig := generateTestSignature(rawPayload, validTimestamp, secret)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(rawPayload))
	req.Header.Set("X-Corteza-Signature", sig)
	req.Header.Set("X-Corteza-Timestamp", validTimestamp)
	ctx := middleware.WithTenantID(req.Context(), types.TenantID(validUUID))
	processor.ServeHTTP(httptest.NewRecorder(), req.WithContext(ctx))

	// Verify DB: Rollback occurred, event lock shouldn't exist because transaction aborted
	var count int
	_ = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform.webhook_events WHERE event_id = 'simulate-fail'").Scan(&count)
	assert.Equal(t, 0, count, "Event should NOT be in DB because business logic failed and rolled back")

	// Retry should be possible because lock doesn't exist
	dispatcher.simulateFailID = "" // Remove failure simulation
	w2 := httptest.NewRecorder()
	processor.ServeHTTP(w2, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, w2.Code)
}

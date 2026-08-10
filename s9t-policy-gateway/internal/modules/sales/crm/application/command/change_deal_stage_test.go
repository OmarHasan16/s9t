package command_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"s9t.os/internal/app/http/middleware"
	"s9t.os/internal/modules/sales/crm/application/command"
	"s9t.os/internal/modules/sales/crm/application/ports"
	"s9t.os/internal/modules/sales/crm/policy"
	"s9t.os/internal/platform/outbox"
	"s9t.os/internal/platform/types"
)

// --- Hand-Written Fakes ---

type FakeClock struct {
	fixedTime time.Time
}
func (c *FakeClock) Now() time.Time { return c.fixedTime }

type FakeCortezaCRMGateway struct {
	GetDealFunc         func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error)
	UpdateDealStageFunc func(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) (int, error)
	
	GetDealCallCount         int
	UpdateDealStageCallCount int
}

func (f *FakeCortezaCRMGateway) GetContact(ctx context.Context, tenantID types.TenantID, contactID string) (*ports.ContactProjection, error) { return nil, nil }
func (f *FakeCortezaCRMGateway) UpdateContactStage(ctx context.Context, tenantID types.TenantID, contactID string, stage string) error { return nil }
func (f *FakeCortezaCRMGateway) GetDeal(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
	f.GetDealCallCount++
	if f.GetDealFunc != nil {
		return f.GetDealFunc(ctx, tenantID, dealID)
	}
	return nil, nil
}
func (f *FakeCortezaCRMGateway) UpdateDealStage(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) (int, error) {
	f.UpdateDealStageCallCount++
	if f.UpdateDealStageFunc != nil {
		return f.UpdateDealStageFunc(ctx, tenantID, dealID, targetStage, expectedVersion)
	}
	return expectedVersion + 1, nil
}

type FakeDBTransactionManager struct {
	RunInTransactionFunc       func(ctx context.Context, fn func(txCtx context.Context) error) error
	CheckIdempotencyStatusFunc func(ctx context.Context, tenantID types.TenantID, idempotencyKey string) (string, string, error)
	AcquireCommandLeaseFunc    func(ctx context.Context, tenantID types.TenantID, idempotencyKey string, commandHash string, workerID string, now time.Time, staleBefore time.Time) (string, string, string, error)
	UpdateOperationStatusFunc  func(ctx context.Context, tenantID types.TenantID, operationID string, status string, lastError string, leaseToken string) error
	RequireReconciliationFunc  func(ctx context.Context, tenantID types.TenantID, operationID string, reason string, leaseToken string) error
	SaveOutboxEventFunc        func(ctx context.Context, event outbox.Event) error

	RunInTransactionCallCount       int
	CheckIdempotencyStatusCallCount int
	AcquireCommandLeaseCallCount    int
	UpdateOperationStatusCallCount  int
	RequireReconciliationCallCount  int
	SaveOutboxEventCallCount        int
	CapturedEvent                   *outbox.Event
	CapturedOpStatus                string
}

func (f *FakeDBTransactionManager) IsEventProcessed(ctx context.Context, eventID string) (bool, error) { return false, nil }
func (f *FakeDBTransactionManager) MarkEventProcessed(ctx context.Context, eventID string) error { return nil }

func (f *FakeDBTransactionManager) CheckIdempotencyStatus(ctx context.Context, tenantID types.TenantID, idempotencyKey string) (status string, hash string, err error) {
	f.CheckIdempotencyStatusCallCount++
	if f.CheckIdempotencyStatusFunc != nil {
		return f.CheckIdempotencyStatusFunc(ctx, tenantID, idempotencyKey)
	}
	return "", "", ports.ErrOperationNotFound
}

func (f *FakeDBTransactionManager) AcquireCommandLease(ctx context.Context, tenantID types.TenantID, idempotencyKey string, commandHash string, workerID string, now time.Time, staleBefore time.Time) (string, string, string, error) {
	f.AcquireCommandLeaseCallCount++
	if f.AcquireCommandLeaseFunc != nil {
		return f.AcquireCommandLeaseFunc(ctx, tenantID, idempotencyKey, commandHash, workerID, now, staleBefore)
	}
	return "op-123", "processing", "lease-456", nil
}

func (f *FakeDBTransactionManager) UpdateOperationStatus(ctx context.Context, tenantID types.TenantID, operationID string, status string, lastError string, leaseToken string) error {
	f.UpdateOperationStatusCallCount++
	f.CapturedOpStatus = status
	if f.UpdateOperationStatusFunc != nil {
		return f.UpdateOperationStatusFunc(ctx, tenantID, operationID, status, lastError, leaseToken)
	}
	return nil
}

func (f *FakeDBTransactionManager) RequireReconciliation(ctx context.Context, tenantID types.TenantID, operationID string, reason string, leaseToken string) error {
	f.RequireReconciliationCallCount++
	if f.RequireReconciliationFunc != nil {
		return f.RequireReconciliationFunc(ctx, tenantID, operationID, reason, leaseToken)
	}
	return nil
}

func (f *FakeDBTransactionManager) RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	f.RunInTransactionCallCount++
	if f.RunInTransactionFunc != nil {
		return f.RunInTransactionFunc(ctx, fn)
	}
	return fn(ctx)
}

func (f *FakeDBTransactionManager) SetTenantContext(ctx context.Context, tenantID types.TenantID) error { return nil }
func (f *FakeDBTransactionManager) SaveOutboxEvent(ctx context.Context, event outbox.Event) error {
	f.SaveOutboxEventCallCount++
	f.CapturedEvent = &event
	if f.SaveOutboxEventFunc != nil {
		return f.SaveOutboxEventFunc(ctx, event)
	}
	return nil
}

// --- Test Setup ---

func setupTest() (*policy.DealPolicy, *FakeCortezaCRMGateway, *FakeDBTransactionManager, *command.ChangeDealStageHandler, context.Context) {
	p := policy.NewDealPolicy()
	gw := &FakeCortezaCRMGateway{}
	tx := &FakeDBTransactionManager{}
	clock := &FakeClock{fixedTime: time.Date(2026, 8, 10, 15, 0, 0, 0, time.UTC)}
	handler := command.NewChangeDealStageHandler(p, gw, tx, clock)
	
	ctx := middleware.WithTenantID(context.Background(), "tenant-1")
	ctx = middleware.WithActorID(ctx, "actor-1")
	return p, gw, tx, handler, ctx
}

func ptrStr(s string) *string { return &s }

// --- Tests ---

func TestChangeDealStage_Success(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	cmd := command.ChangeDealStageCommand{DealID: "d1", ExpectedVersion: 1, TargetStage: policy.DealStageQualified}
	
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
	}
	gw.UpdateDealStageFunc = func(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) (int, error) {
		return 2, nil
	}

	err := handler.Execute(ctx, cmd)
	assert.NoError(t, err)
	assert.Equal(t, 1, tx.CheckIdempotencyStatusCallCount)
	assert.Equal(t, 1, tx.AcquireCommandLeaseCallCount)
	assert.Equal(t, 1, gw.GetDealCallCount)
	assert.Equal(t, 1, gw.UpdateDealStageCallCount)
	assert.Equal(t, 1, tx.SaveOutboxEventCallCount)
	assert.Equal(t, 1, tx.UpdateOperationStatusCallCount)
	assert.Equal(t, "completed", tx.CapturedOpStatus)
}

func TestChangeDealStage_SameIdempotencyKeySameCommandIsIdempotent(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	
	cmd := command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: policy.DealStageQualified}
	// Note: We need to figure out the hash of cmd for the mock to match.
	// We'll just hardcode a bypass check.
	tx.CheckIdempotencyStatusFunc = func(ctx context.Context, tenantID types.TenantID, idempotencyKey string) (string, string, error) {
		// Mock returning same hash
		// For unit test, we just set hash to exactly what it computes.
		return "completed", command.HashCommand(cmd), nil
	}
	
	err := handler.Execute(ctx, cmd)
	assert.NoError(t, err) // Idempotent success early exit
	assert.Equal(t, 1, tx.CheckIdempotencyStatusCallCount)
	assert.Equal(t, 0, gw.GetDealCallCount)
	assert.Equal(t, 0, tx.AcquireCommandLeaseCallCount)
}

func TestChangeDealStage_SameIdempotencyKeyDifferentCommandHashIsRejected(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	
	tx.CheckIdempotencyStatusFunc = func(ctx context.Context, tenantID types.TenantID, idempotencyKey string) (string, string, error) {
		return "completed", "different-hash", nil
	}
	
	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: policy.DealStageQualified})
	assert.ErrorIs(t, err, command.ErrCommandConflict)
	assert.Equal(t, 0, gw.GetDealCallCount)
}

func TestChangeDealStage_FreshProcessingOperationIsLocked(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	
	cmd := command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: policy.DealStageQualified}
	tx.CheckIdempotencyStatusFunc = func(ctx context.Context, tenantID types.TenantID, idempotencyKey string) (string, string, error) {
		return "processing", command.HashCommand(cmd), nil
	}
	
	err := handler.Execute(ctx, cmd)
	assert.ErrorIs(t, err, command.ErrCommandLocked)
	assert.Equal(t, 0, gw.GetDealCallCount)
}

func TestChangeDealStage_CortezaUpdateFailureMarksRetryableFailure(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	cortezaErr := errors.New("corteza down")
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
	}
	gw.UpdateDealStageFunc = func(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) (int, error) {
		return 0, cortezaErr
	}
	tx.UpdateOperationStatusFunc = func(ctx context.Context, tenantID types.TenantID, operationID string, status string, lastError string, leaseToken string) error {
		return errors.New("db error")
	}
	
	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: policy.DealStageQualified})
	
	assert.ErrorContains(t, err, "corteza update failed")
	assert.Equal(t, 0, tx.SaveOutboxEventCallCount) // No outbox pollution
	assert.Equal(t, 1, tx.UpdateOperationStatusCallCount)
	assert.Equal(t, "retryable_failure", tx.CapturedOpStatus)
}

func TestChangeDealStage_OutboxFailureRequestsReconciliation(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
	}
	
	dbErr := errors.New("db down")
	tx.RunInTransactionFunc = func(ctx context.Context, fn func(txCtx context.Context) error) error {
		return dbErr
	}

	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: policy.DealStageQualified})
	
	assert.ErrorIs(t, err, command.ErrDualWriteFailed)
	assert.Equal(t, 1, gw.UpdateDealStageCallCount)
	assert.Equal(t, 1, tx.RequireReconciliationCallCount)
}

func TestChangeDealStage_OptimisticLockFailure(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 2}, nil
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1})
	assert.ErrorIs(t, err, command.ErrVersionMismatch)
	assert.Equal(t, 0, tx.AcquireCommandLeaseCallCount) 
}

func TestChangeDealStage_PolicyTransitionFailure(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: policy.DealStageWon})
	assert.ErrorIs(t, err, policy.ErrInvalidDealTransition)
	assert.Equal(t, 0, tx.AcquireCommandLeaseCallCount) 
}

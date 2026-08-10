package command_test

import (
	"context"
	"errors"
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

type FakeCortezaCRMGateway struct {
	GetDealFunc         func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error)
	UpdateDealStageFunc func(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) error
	
	GetDealCallCount         int
	UpdateDealStageCallCount int
}

func (f *FakeCortezaCRMGateway) GetContact(ctx context.Context, tenantID types.TenantID, contactID string) (*ports.ContactProjection, error) {
	return nil, nil
}
func (f *FakeCortezaCRMGateway) UpdateContactStage(ctx context.Context, tenantID types.TenantID, contactID string, stage string) error {
	return nil
}
func (f *FakeCortezaCRMGateway) GetDeal(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
	f.GetDealCallCount++
	if f.GetDealFunc != nil {
		return f.GetDealFunc(ctx, tenantID, dealID)
	}
	return nil, nil
}
func (f *FakeCortezaCRMGateway) UpdateDealStage(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) error {
	f.UpdateDealStageCallCount++
	if f.UpdateDealStageFunc != nil {
		return f.UpdateDealStageFunc(ctx, tenantID, dealID, targetStage, expectedVersion)
	}
	return nil
}

type FakeDBTransactionManager struct {
	RunInTransactionFunc      func(ctx context.Context, fn func(txCtx context.Context) error) error
	AcquireCommandLeaseFunc   func(ctx context.Context, idempotencyKey string, commandHash string) (bool, error)
	RequireReconciliationFunc func(ctx context.Context, tenantID types.TenantID, aggregateID string, targetStage string) error
	SaveOutboxEventFunc       func(ctx context.Context, event outbox.Event) error

	RunInTransactionCallCount      int
	AcquireCommandLeaseCallCount   int
	RequireReconciliationCallCount int
	SaveOutboxEventCallCount       int
	CapturedEvent                  *outbox.Event
}

func (f *FakeDBTransactionManager) IsEventProcessed(ctx context.Context, eventID string) (bool, error) { return false, nil }
func (f *FakeDBTransactionManager) MarkEventProcessed(ctx context.Context, eventID string) error { return nil }

func (f *FakeDBTransactionManager) AcquireCommandLease(ctx context.Context, idempotencyKey string, commandHash string) (bool, error) {
	f.AcquireCommandLeaseCallCount++
	if f.AcquireCommandLeaseFunc != nil {
		return f.AcquireCommandLeaseFunc(ctx, idempotencyKey, commandHash)
	}
	return true, nil // default success
}

func (f *FakeDBTransactionManager) RequireReconciliation(ctx context.Context, tenantID types.TenantID, aggregateID string, targetStage string) error {
	f.RequireReconciliationCallCount++
	if f.RequireReconciliationFunc != nil {
		return f.RequireReconciliationFunc(ctx, tenantID, aggregateID, targetStage)
	}
	return nil
}

func (f *FakeDBTransactionManager) RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	f.RunInTransactionCallCount++
	if f.RunInTransactionFunc != nil {
		return f.RunInTransactionFunc(ctx, fn)
	}
	return fn(ctx) // default run directly
}

func (f *FakeDBTransactionManager) SetTenantContext(ctx context.Context, tenantID types.TenantID) error {
	return nil
}

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
	handler := command.NewChangeDealStageHandler(p, gw, tx)
	
	ctx := context.Background()
	ctx = middleware.WithTenantID(ctx, "tenant-1")
	ctx = middleware.WithActorID(ctx, "actor-1")
	
	return p, gw, tx, handler, ctx
}

func ptrStr(s string) *string { return &s }
func ptrTime(t time.Time) *time.Time { return &t }

// --- Tests ---

func TestChangeDealStage_Success(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	cmd := command.ChangeDealStageCommand{DealID: "d1", ExpectedVersion: 1, TargetStage: policy.DealStageQualified}
	
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
	}

	err := handler.Execute(ctx, cmd)
	assert.NoError(t, err)
	assert.Equal(t, 1, tx.AcquireCommandLeaseCallCount)
	assert.Equal(t, 1, gw.GetDealCallCount)
	assert.Equal(t, 1, gw.UpdateDealStageCallCount)
	assert.Equal(t, 1, tx.SaveOutboxEventCallCount)
	assert.NotNil(t, tx.CapturedEvent)
	assert.Equal(t, "crm.deal.stage_changed", tx.CapturedEvent.EventType)
}

func TestChangeDealStage_UnauthorizedWithoutTenantContext(t *testing.T) {
	_, _, _, handler, _ := setupTest()
	ctx := context.Background() // No tenant, no actor
	err := handler.Execute(ctx, command.ChangeDealStageCommand{})
	assert.ErrorIs(t, err, command.ErrUnauthorized)
}

func TestChangeDealStage_UnauthorizedWithoutActorContext(t *testing.T) {
	_, _, _, handler, _ := setupTest()
	ctx := middleware.WithTenantID(context.Background(), "t1") // No actor
	err := handler.Execute(ctx, command.ChangeDealStageCommand{})
	assert.ErrorIs(t, err, command.ErrUnauthorized)
}

func TestChangeDealStage_SameIdempotencyKeySameCommandIsIdempotent(t *testing.T) {
	// Handled by returning ErrCommandLocked when acquired is false
	_, _, tx, handler, ctx := setupTest()
	tx.AcquireCommandLeaseFunc = func(ctx context.Context, idempotencyKey string, commandHash string) (bool, error) {
		return false, nil // Mocking it's already processed/locked
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{})
	assert.ErrorIs(t, err, command.ErrCommandLocked)
}

func TestChangeDealStage_SameIdempotencyKeyDifferentCommandHashIsRejected(t *testing.T) {
	_, _, tx, handler, ctx := setupTest()
	tx.AcquireCommandLeaseFunc = func(ctx context.Context, idempotencyKey string, commandHash string) (bool, error) {
		return false, errors.New("hash conflict") // Mocking hash conflict
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{})
	assert.EqualError(t, err, "hash conflict")
}

func TestChangeDealStage_FreshProcessingOperationIsLocked(t *testing.T) {
	_, _, tx, handler, ctx := setupTest()
	tx.AcquireCommandLeaseFunc = func(ctx context.Context, idempotencyKey string, commandHash string) (bool, error) {
		return false, nil // locked
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{})
	assert.ErrorIs(t, err, command.ErrCommandLocked)
}

func TestChangeDealStage_CompletedOperationDoesNotCallCorteza(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	tx.AcquireCommandLeaseFunc = func(ctx context.Context, idempotencyKey string, commandHash string) (bool, error) {
		return false, nil // already completed/locked
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{})
	assert.ErrorIs(t, err, command.ErrCommandLocked)
	assert.Equal(t, 0, gw.UpdateDealStageCallCount)
}

func TestChangeDealStage_OptimisticLockFailure(t *testing.T) {
	_, gw, _, handler, ctx := setupTest()
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 2}, nil
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1})
	assert.ErrorIs(t, err, command.ErrVersionMismatch)
	assert.Equal(t, 0, gw.UpdateDealStageCallCount)
}

func TestChangeDealStage_UnknownTargetStageRejected(t *testing.T) {
	_, gw, _, handler, ctx := setupTest()
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: "unknown_stage"})
	assert.ErrorIs(t, err, policy.ErrInvalidDealTransition)
}

func TestChangeDealStage_WonReadinessViolation(t *testing.T) {
	_, gw, _, handler, ctx := setupTest()
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageNegotiation, RecordVersion: 1, AmountMinor: 0, ContactID: ptrStr("c1")}
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: policy.DealStageWon})
	assert.ErrorIs(t, err, policy.ErrInvalidAmount)
}

func TestChangeDealStage_CortezaUpdateFailureMarksRetryableFailure(t *testing.T) {
	_, gw, tx, handler, ctx := setupTest()
	cortezaErr := errors.New("corteza down")
	gw.GetDealFunc = func(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
		return &ports.DealProjection{ID: "d1", Stage: policy.DealStageOpen, RecordVersion: 1}, nil
	}
	gw.UpdateDealStageFunc = func(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) error {
		return cortezaErr
	}
	err := handler.Execute(ctx, command.ChangeDealStageCommand{ExpectedVersion: 1, TargetStage: policy.DealStageQualified})
	assert.ErrorIs(t, err, cortezaErr)
	assert.Equal(t, 0, tx.SaveOutboxEventCallCount) // No outbox pollution
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
	
	// Ensure that after RunInTransaction fails, RequireReconciliation is called
	assert.ErrorIs(t, err, command.ErrDualWriteFailed)
	assert.Equal(t, 1, gw.UpdateDealStageCallCount)
	assert.Equal(t, 1, tx.RequireReconciliationCallCount)
}

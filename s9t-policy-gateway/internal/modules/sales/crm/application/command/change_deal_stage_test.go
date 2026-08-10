package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"s9t.os/internal/modules/sales/crm/application/command"
	"s9t.os/internal/modules/sales/crm/application/ports"
	"s9t.os/internal/modules/sales/crm/policy"
	"s9t.os/internal/platform/outbox"
	"s9t.os/internal/platform/types"
)

// --- Mocks ---

type MockCortezaCRMGateway struct {
	mock.Mock
}

func (m *MockCortezaCRMGateway) GetContact(ctx context.Context, tenantID types.TenantID, contactID string) (*ports.ContactProjection, error) {
	args := m.Called(ctx, tenantID, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.ContactProjection), args.Error(1)
}

func (m *MockCortezaCRMGateway) UpdateContactStage(ctx context.Context, tenantID types.TenantID, contactID string, stage string) error {
	args := m.Called(ctx, tenantID, contactID, stage)
	return args.Error(0)
}

func (m *MockCortezaCRMGateway) GetDeal(ctx context.Context, tenantID types.TenantID, dealID string) (*ports.DealProjection, error) {
	args := m.Called(ctx, tenantID, dealID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ports.DealProjection), args.Error(1)
}

func (m *MockCortezaCRMGateway) UpdateDealStage(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) error {
	args := m.Called(ctx, tenantID, dealID, targetStage, expectedVersion)
	return args.Error(0)
}

type MockDBTransactionManager struct {
	mock.Mock
}

func (m *MockDBTransactionManager) RunInTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	// Call the provided function with the passed context to simulate a transaction block
	err := fn(ctx)
	// We also record the call so we can assert it was made
	args := m.Called(ctx)
	if args.Error(0) != nil {
		return args.Error(0)
	}
	return err
}

func (m *MockDBTransactionManager) SetTenantContext(ctx context.Context, tenantID types.TenantID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockDBTransactionManager) SaveOutboxEvent(ctx context.Context, event outbox.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockDBTransactionManager) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	args := m.Called(ctx, eventID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDBTransactionManager) MarkEventProcessed(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

// --- Tests ---

func setupTest() (*policy.DealPolicy, *MockCortezaCRMGateway, *MockDBTransactionManager, *command.ChangeDealStageHandler) {
	p := policy.NewDealPolicy()
	gw := new(MockCortezaCRMGateway)
	tx := new(MockDBTransactionManager)
	handler := command.NewChangeDealStageHandler(p, gw, tx)
	return p, gw, tx, handler
}

func ptrStr(s string) *string {
	return &s
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func TestChangeDealStage_Success(t *testing.T) {
	_, gw, tx, handler := setupTest()
	ctx := context.Background()
	tenantID := types.TenantID("tenant-123")
	dealID := "deal-1"

	cmd := command.ChangeDealStageCommand{
		DealID:          dealID,
		TenantID:        tenantID,
		ExpectedVersion: 1,
		TargetStage:     policy.DealStageQualified,
		ActorID:         "actor-1",
		CorrelationID:   "corr-1",
	}

	deal := &ports.DealProjection{
		ID:            dealID,
		Stage:         policy.DealStageOpen,
		RecordVersion: 1,
	}

	gw.On("GetDeal", ctx, tenantID, dealID).Return(deal, nil)
	gw.On("UpdateDealStage", ctx, tenantID, dealID, policy.DealStageQualified, 1).Return(nil)
	tx.On("RunInTransaction", ctx).Return(nil)
	tx.On("SetTenantContext", ctx, tenantID).Return(nil)
	tx.On("SaveOutboxEvent", ctx, mock.AnythingOfType("outbox.Event")).Return(nil)

	err := handler.Execute(ctx, cmd)

	assert.NoError(t, err)
	gw.AssertExpectations(t)
	tx.AssertExpectations(t)
}

func TestChangeDealStage_OptimisticLockFailure(t *testing.T) {
	_, gw, tx, handler := setupTest()
	ctx := context.Background()

	cmd := command.ChangeDealStageCommand{
		DealID:          "deal-1",
		TenantID:        "tenant-123",
		ExpectedVersion: 1, // Client expects version 1
		TargetStage:     policy.DealStageQualified,
	}

	deal := &ports.DealProjection{
		ID:            "deal-1",
		Stage:         policy.DealStageOpen,
		RecordVersion: 2, // Backend has version 2
	}

	gw.On("GetDeal", ctx, cmd.TenantID, cmd.DealID).Return(deal, nil)

	err := handler.Execute(ctx, cmd)

	assert.ErrorIs(t, err, command.ErrVersionMismatch)
	gw.AssertExpectations(t)
	tx.AssertNotCalled(t, "RunInTransaction", mock.Anything)
	gw.AssertNotCalled(t, "UpdateDealStage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestChangeDealStage_PolicyTransitionFailure(t *testing.T) {
	_, gw, tx, handler := setupTest()
	ctx := context.Background()

	cmd := command.ChangeDealStageCommand{
		DealID:          "deal-1",
		TenantID:        "tenant-123",
		ExpectedVersion: 1,
		TargetStage:     policy.DealStageWon, // Invalid transition from Open
	}

	deal := &ports.DealProjection{
		ID:            "deal-1",
		Stage:         policy.DealStageOpen, // Currently open
		RecordVersion: 1,
	}

	gw.On("GetDeal", ctx, cmd.TenantID, cmd.DealID).Return(deal, nil)

	err := handler.Execute(ctx, cmd)

	assert.ErrorIs(t, err, policy.ErrInvalidDealTransition)
	gw.AssertExpectations(t)
	tx.AssertNotCalled(t, "RunInTransaction", mock.Anything)
}

func TestChangeDealStage_WonReadinessViolation(t *testing.T) {
	_, gw, tx, handler := setupTest()
	ctx := context.Background()

	cmd := command.ChangeDealStageCommand{
		DealID:          "deal-1",
		TenantID:        "tenant-123",
		ExpectedVersion: 1,
		TargetStage:     policy.DealStageWon,
	}

	deal := &ports.DealProjection{
		ID:            "deal-1",
		Stage:         policy.DealStageNegotiation, // Valid transition, but missing requirements
		RecordVersion: 1,
		AmountMinor:   0, // Readiness violation
		ContactID:     ptrStr("contact-1"),
		CompanyID:     nil,
	}

	gw.On("GetDeal", ctx, cmd.TenantID, cmd.DealID).Return(deal, nil)

	err := handler.Execute(ctx, cmd)

	assert.ErrorIs(t, err, policy.ErrInvalidAmount)
	gw.AssertExpectations(t)
	tx.AssertNotCalled(t, "RunInTransaction", mock.Anything)
}

func TestChangeDealStage_CortezaGatewayRejection(t *testing.T) {
	_, gw, tx, handler := setupTest()
	ctx := context.Background()

	cmd := command.ChangeDealStageCommand{
		DealID:          "deal-1",
		TenantID:        "tenant-123",
		ExpectedVersion: 1,
		TargetStage:     policy.DealStageQualified,
	}

	deal := &ports.DealProjection{
		ID:            "deal-1",
		Stage:         policy.DealStageOpen,
		RecordVersion: 1,
	}

	cortezaErr := errors.New("corteza api error or concurrent lock")

	gw.On("GetDeal", ctx, cmd.TenantID, cmd.DealID).Return(deal, nil)
	gw.On("UpdateDealStage", ctx, cmd.TenantID, cmd.DealID, policy.DealStageQualified, 1).Return(cortezaErr)

	err := handler.Execute(ctx, cmd)

	assert.ErrorIs(t, err, cortezaErr)
	gw.AssertExpectations(t)
	tx.AssertNotCalled(t, "RunInTransaction", mock.Anything)
}

func TestChangeDealStage_TransactionFailure(t *testing.T) {
	_, gw, tx, handler := setupTest()
	ctx := context.Background()
	tenantID := types.TenantID("tenant-123")

	cmd := command.ChangeDealStageCommand{
		DealID:          "deal-1",
		TenantID:        tenantID,
		ExpectedVersion: 1,
		TargetStage:     policy.DealStageQualified,
	}

	deal := &ports.DealProjection{
		ID:            "deal-1",
		Stage:         policy.DealStageOpen,
		RecordVersion: 1,
	}

	dbErr := errors.New("database connection lost")

	gw.On("GetDeal", ctx, cmd.TenantID, cmd.DealID).Return(deal, nil)
	gw.On("UpdateDealStage", ctx, cmd.TenantID, cmd.DealID, policy.DealStageQualified, 1).Return(nil)
	
	// We simulate the transaction failing
	tx.On("RunInTransaction", ctx).Return(dbErr)
	tx.On("SetTenantContext", ctx, tenantID).Return(nil)
	tx.On("SaveOutboxEvent", ctx, mock.Anything).Return(nil)

	err := handler.Execute(ctx, cmd)

	// Since we mock RunInTransaction to just return the error it receives (if we set it), 
	// wait, our mock implementation above returns fn(ctx) error OR args.Error(0).
	// If tx.On("RunInTransaction", ctx).Return(dbErr) is set, args.Error(0) is dbErr.
	assert.ErrorIs(t, err, dbErr)
	gw.AssertExpectations(t)
	tx.AssertExpectations(t)
}

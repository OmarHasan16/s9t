package command

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"s9t.os/internal/app/http/middleware"
	"s9t.os/internal/modules/sales/crm/application/ports"
	"s9t.os/internal/modules/sales/crm/domain/valueobject"
	"s9t.os/internal/modules/sales/crm/policy"
	"s9t.os/internal/platform/outbox"
)

var (
	ErrVersionMismatch         = errors.New("optimistic lock failed: deal version is stale")
	ErrUnauthorized            = errors.New("unauthorized: missing required context")
	ErrCommandLocked           = errors.New("command is already being processed")
	ErrOperationConflict       = errors.New("idempotency key reused with different command payload")
	ErrDualWriteFailed         = errors.New("corteza update succeeded but local outbox failed: requires reconciliation")
	ErrReconciliationPending   = errors.New("operation requires reconciliation")
	ErrOperationDeadLetter     = errors.New("operation failed permanently")
)

type ChangeDealStageCommand struct {
	DealID          string
	ExpectedVersion int
	TargetStage     string
	CorrelationID   string
	IdempotencyKey  string
}

type ChangeDealStageHandler struct {
	policy        *policy.DealPolicy
	crmGateway    ports.CortezaCRMGateway
	dbTransaction ports.DBTransactionManager
	clock         ports.Clock
	leaseDuration time.Duration
	workerID      string
}

func NewChangeDealStageHandler(p *policy.DealPolicy, gw ports.CortezaCRMGateway, tx ports.DBTransactionManager, clock ports.Clock, leaseDuration time.Duration) *ChangeDealStageHandler {
	if leaseDuration <= 0 {
		leaseDuration = 5 * time.Minute // default
	}
	return &ChangeDealStageHandler{policy: p, crmGateway: gw, dbTransaction: tx, clock: clock, leaseDuration: leaseDuration, workerID: "gateway-" + uuid.New().String()}
}

func HashCommand(cmd ChangeDealStageCommand) (string, error) {
	b, err := json.Marshal(cmd)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func (h *ChangeDealStageHandler) Execute(ctx context.Context, cmd ChangeDealStageCommand) error {
	// 1. Auth identity validate
	tenantID, ok := middleware.TenantIDFromContext(ctx)
	if !ok || tenantID == "" {
		return ErrUnauthorized
	}
	actorID, ok := middleware.ActorIDFromContext(ctx)
	if !ok || actorID == "" {
		return ErrUnauthorized
	}

	cmdHash, err := HashCommand(cmd)
	if err != nil {
		return fmt.Errorf("hash command failed: %w", err)
	}

	// 2. Existing key/hash read-only lookup
	status, hash, err := h.dbTransaction.CheckIdempotencyStatus(ctx, tenantID, cmd.IdempotencyKey)
	if err != nil {
		if !errors.Is(err, ports.ErrOperationNotFound) {
			return fmt.Errorf("check idempotency status: %w", err)
		}
	} else {
		if hash != cmdHash {
			// 4. Hash mismatch -> conflict, row unchanged
			return ErrOperationConflict
		}
		
		switch status {
		case "completed":
			return nil
		case "processing":
			return ErrCommandLocked
		case "reconciliation_required":
			return ErrReconciliationPending
		case "conflict":
			return ErrOperationConflict
		case "dead_letter":
			return ErrOperationDeadLetter
		case "retryable_failure", "pending":
			// proceed to validate & reclaim
		default:
			return fmt.Errorf("unknown idempotency status: %s", status)
		}
	}

	// 6. Deal/version/domain validation
	deal, err := h.crmGateway.GetDeal(ctx, tenantID, cmd.DealID)
	if err != nil {
		return fmt.Errorf("corteza fetch failed: %w", err)
	}

	if deal.RecordVersion != cmd.ExpectedVersion {
		return ErrVersionMismatch
	}

	if err := h.policy.CanTransition(deal.Stage, cmd.TargetStage); err != nil {
		return err
	}
	if cmd.TargetStage == policy.DealStageWon {
		amount, err := valueobject.NewMoney(deal.AmountMinor, deal.Currency)
		if err != nil {
			return fmt.Errorf("invalid deal money: %w", err)
		}
		if err := h.policy.EnsureWonReadiness(deal.ContactID, deal.CompanyID, amount, deal.ExpectedCloseDate); err != nil {
			return err
		}
	}

	// 7. New operation reserve or retryable operation claim
	now := h.clock.Now().UTC()
	staleBefore := now.Add(-h.leaseDuration)
	
	meta := ports.OperationMetadata{
		AggregateType:   "crm.deal",
		AggregateID:     cmd.DealID,
		ExpectedVersion: cmd.ExpectedVersion,
		TargetState:     cmd.TargetStage,
		CorrelationID:   cmd.CorrelationID,
		ActorID:         actorID,
	}

	opID, opStatus, leaseToken, err := h.dbTransaction.AcquireCommandLease(ctx, tenantID, cmd.IdempotencyKey, cmdHash, meta, h.workerID, now, staleBefore)
	if err != nil {
		if errors.Is(err, ports.ErrIdempotencyHashConflict) {
			return ErrOperationConflict
		}
		return fmt.Errorf("acquire operation lease: %w", err)
	}

	if opStatus != "processing" {
		// If another worker beat us to it, or it completed just now
		switch opStatus {
		case "completed":
			return nil
		case "reconciliation_required":
			return ErrReconciliationPending
		case "conflict":
			return ErrOperationConflict
		case "dead_letter":
			return ErrOperationDeadLetter
		default:
			return ErrCommandLocked
		}
	}

	// 8. Corteza update
	actualVersion, err := h.crmGateway.UpdateDealStage(ctx, tenantID, cmd.DealID, cmd.TargetStage, cmd.ExpectedVersion)
	if err != nil {
		statusErr := h.dbTransaction.UpdateOperationStatus(ctx, tenantID, opID, "retryable_failure", err.Error(), leaseToken)
		if statusErr != nil {
			return fmt.Errorf("corteza update failed: %w, also status update failed: %v", err, statusErr)
		}
		return fmt.Errorf("corteza update failed: %w", err)
	}

	eventType := "crm.deal.stage_changed"
	if cmd.TargetStage == policy.DealStageWon {
		eventType = "crm.deal.won"
	} else if cmd.TargetStage == policy.DealStageLost {
		eventType = "crm.deal.lost"
	}

	eventPayload, err := json.Marshal(map[string]interface{}{
		"deal_id":        cmd.DealID,
		"old_stage":      deal.Stage,
		"new_stage":      cmd.TargetStage,
		"amount_minor":   deal.AmountMinor,
		"currency":       deal.Currency,
		"actor_id":       actorID,
		"correlation_id": cmd.CorrelationID,
		"actual_version": actualVersion,
	})
	if err != nil {
		if statusErr := h.dbTransaction.UpdateOperationStatus(ctx, tenantID, opID, "retryable_failure", "marshal deal event failed", leaseToken); statusErr != nil {
			return fmt.Errorf("marshal deal event: %w, also status update failed: %v", err, statusErr)
		}
		return fmt.Errorf("marshal deal event: %w", err) 
	}

	// 9. Outbox + operation completion in same DB tx
	err = h.dbTransaction.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := h.dbTransaction.SetTenantContext(txCtx, tenantID); err != nil {
			return err
		}

		event := outbox.Event{
			ID:           uuid.New().String(),
			TenantID:     string(tenantID),
			EventType:    eventType,
			AggregateID:  cmd.DealID,
			Payload:      eventPayload,
			Status:       outbox.StatusPending,
			CreatedAt:    now,
			AvailableAt:  now,
			AttemptCount: 0,
		}

		if err := h.dbTransaction.SaveOutboxEvent(txCtx, event); err != nil {
			return err
		}
		
		return h.dbTransaction.UpdateOperationStatus(txCtx, tenantID, opID, "completed", "", leaseToken)
	})

	if err != nil {
		// exact operationID + tenant-scoped reconciliation
		if reconErr := h.dbTransaction.RequireReconciliation(ctx, tenantID, opID, "outbox persistence failed", leaseToken); reconErr != nil {
			return fmt.Errorf("dual-write failed: %w; reconciliation recording failed: %v", err, reconErr)
		}
		return ErrDualWriteFailed
	}

	return nil
}

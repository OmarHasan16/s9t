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
	ErrVersionMismatch      = errors.New("optimistic lock failed: deal version is stale")
	ErrUnauthorized         = errors.New("unauthorized: missing required context")
	ErrCommandLocked        = errors.New("command is already being processed")
	ErrCommandConflict      = errors.New("idempotency key reused with different command payload")
	ErrDualWriteFailed      = errors.New("corteza update succeeded but local outbox failed: requires reconciliation")
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
}

func NewChangeDealStageHandler(p *policy.DealPolicy, gw ports.CortezaCRMGateway, tx ports.DBTransactionManager) *ChangeDealStageHandler {
	return &ChangeDealStageHandler{policy: p, crmGateway: gw, dbTransaction: tx}
}

func hashCommand(cmd ChangeDealStageCommand) string {
	b, _ := json.Marshal(cmd)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (h *ChangeDealStageHandler) Execute(ctx context.Context, cmd ChangeDealStageCommand) error {
	// 1. Context Verification
	tenantID, ok := middleware.TenantIDFromContext(ctx)
	if !ok || tenantID == "" {
		return ErrUnauthorized
	}
	actorID, ok := middleware.ActorIDFromContext(ctx)
	if !ok || actorID == "" {
		return ErrUnauthorized
	}

	// 2. Fetch current projection from Corteza
	deal, err := h.crmGateway.GetDeal(ctx, tenantID, cmd.DealID)
	if err != nil {
		return fmt.Errorf("corteza fetch failed: %w", err)
	}

	// 3. Optimistic Locking / Version Check
	if deal.RecordVersion != cmd.ExpectedVersion {
		return ErrVersionMismatch
	}

	// 4. Pure Policy Validation
	if err := h.policy.CanTransition(deal.Stage, cmd.TargetStage); err != nil {
		return err
	}
	if cmd.TargetStage == policy.DealStageWon {
		amount, _ := valueobject.NewMoney(deal.AmountMinor, deal.Currency)
		if err := h.policy.EnsureWonReadiness(deal.ContactID, deal.CompanyID, amount, deal.ExpectedCloseDate); err != nil {
			return err
		}
	}

	// 5. Idempotency Reserve/Claim
	cmdHash := hashCommand(cmd)
	opID, opStatus, err := h.dbTransaction.AcquireCommandLease(ctx, cmd.IdempotencyKey, cmdHash)
	if err != nil {
		// Handled by DB constraint violation logic internally, mapped to conflict
		return ErrCommandConflict
	}
	
	switch opStatus {
	case "completed":
		return nil // Idempotent success
	case "processing":
		return ErrCommandLocked
	case "retryable_failure":
		// Resume execution
	case "fresh":
		// Proceed
	}

	// 6. EXTERNAL API CALL
	actualVersion, err := h.crmGateway.UpdateDealStage(ctx, tenantID, cmd.DealID, cmd.TargetStage, cmd.ExpectedVersion)
	if err != nil {
		// Update lease on failure
		_ = h.dbTransaction.UpdateOperationStatus(ctx, opID, "retryable_failure", err.Error())
		return fmt.Errorf("corteza update failed: %w", err)
	}

	// 7. Event Payload Generation
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
		"actual_version": actualVersion, // Blocker 4 fix
	})
	if err != nil {
		_ = h.dbTransaction.UpdateOperationStatus(ctx, opID, "retryable_failure", "marshal deal event failed")
		return fmt.Errorf("marshal deal event: %w", err) // Blocker 5 fix
	}

	// 8. LOCAL DB TRANSACTION: Write Intent/Outbox & Update Ledger Atomically
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
			CreatedAt:    time.Now().UTC(),
			AvailableAt:  time.Now().UTC(),
			AttemptCount: 0,
		}

		if err := h.dbTransaction.SaveOutboxEvent(txCtx, event); err != nil {
			return err
		}
		
		// Mark Operation Completed in same tx
		return h.dbTransaction.UpdateOperationStatus(txCtx, opID, "completed", "")
	})

	if err != nil {
		// 9. Dual-Write Reconciliation Catch
		_ = h.dbTransaction.RequireReconciliation(ctx, tenantID, cmd.DealID, cmd.TargetStage)
		return ErrDualWriteFailed
	}

	return nil
}

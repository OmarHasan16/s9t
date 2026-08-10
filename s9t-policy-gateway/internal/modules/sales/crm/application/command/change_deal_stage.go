package command

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"s9t.os/internal/modules/sales/crm/application/ports"
	"s9t.os/internal/modules/sales/crm/domain/valueobject"
	"s9t.os/internal/modules/sales/crm/policy"
	"s9t.os/internal/platform/outbox"
	"s9t.os/internal/platform/types"
)

var ErrVersionMismatch = errors.New("optimistic lock failed: deal version is stale")

type ChangeDealStageCommand struct {
	DealID          string
	TenantID        types.TenantID
	ExpectedVersion int
	TargetStage     string
	ActorID         types.ActorID
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

func (h *ChangeDealStageHandler) Execute(ctx context.Context, cmd ChangeDealStageCommand) error {
	// 1. Fetch current projection from Corteza
	deal, err := h.crmGateway.GetDeal(ctx, cmd.TenantID, cmd.DealID)
	if err != nil {
		return err
	}

	// 2. Optimistic Locking / Version Check
	if deal.RecordVersion != cmd.ExpectedVersion {
		return ErrVersionMismatch
	}

	// 3. Pure Policy Validation (No side-effects)
	if err := h.policy.CanTransition(deal.Stage, cmd.TargetStage); err != nil {
		return err
	}

	// Specialized policy for Won
	if cmd.TargetStage == policy.DealStageWon {
		amount, _ := valueobject.NewMoney(deal.AmountMinor, deal.Currency)
		if err := h.policy.EnsureWonReadiness(deal.ContactID, deal.CompanyID, amount, deal.ExpectedCloseDate); err != nil {
			return err
		}
	}

	// 4. EXTERNAL API CALL: Update Corteza FIRST (Non-transactional boundary)
	err = h.crmGateway.UpdateDealStage(ctx, cmd.TenantID, cmd.DealID, cmd.TargetStage, cmd.ExpectedVersion)
	if err != nil {
		// If Corteza rejects (e.g. concurrent mutation), we abort. No local DB pollution.
		return err
	}

	// 5. Determine Specific Domain Event
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
		"actor_id":       cmd.ActorID,
		"correlation_id": cmd.CorrelationID,
	})
	if err != nil {
		return err // Should realistically never happen on map[string]interface{}
	}

	// 6. LOCAL DB TRANSACTION: Write Intent/Outbox
	return h.dbTransaction.RunInTransaction(ctx, func(txCtx context.Context) error {
		// Enforce Tenant Context
		if err := h.dbTransaction.SetTenantContext(txCtx, cmd.TenantID); err != nil {
			return err
		}

		event := outbox.Event{
			ID:           uuid.New().String(), // In real implementation: inject ID generator
			TenantID:     string(cmd.TenantID),
			EventType:    eventType,
			AggregateID:  cmd.DealID,
			Payload:      eventPayload,
			Status:       outbox.StatusPending,
			CreatedAt:    time.Now().UTC(),
			AvailableAt:  time.Now().UTC(),
			AttemptCount: 0,
		}

		return h.dbTransaction.SaveOutboxEvent(txCtx, event)
	})
}

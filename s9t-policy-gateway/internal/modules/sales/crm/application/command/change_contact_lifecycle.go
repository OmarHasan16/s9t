package command

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid" // Assuming standard UUID usage
	"s9t.os/internal/modules/sales/crm/application/ports"
	"s9t.os/internal/modules/sales/crm/policy"
	"s9t.os/internal/platform/outbox"
	"s9t.os/internal/platform/types"
)

// Added ActorID and CorrelationID for audit and tracing
type ChangeContactLifecycleCommand struct {
	ContactID     string
	TenantID      types.TenantID
	TargetStage   string
	ActorID       types.ActorID // Audit requirement
	CorrelationID string        // Distributed tracing requirement
}

type ChangeContactLifecycleHandler struct {
	policy         *policy.ContactPolicy
	cortezaGateway ports.CortezaCRMGateway
	dbTransaction  ports.DBTransactionManager
}

func NewChangeContactLifecycleHandler(p *policy.ContactPolicy, gw ports.CortezaCRMGateway, tx ports.DBTransactionManager) *ChangeContactLifecycleHandler {
	return &ChangeContactLifecycleHandler{policy: p, cortezaGateway: gw, dbTransaction: tx}
}

func (h *ChangeContactLifecycleHandler) Execute(ctx context.Context, cmd ChangeContactLifecycleCommand) error {
	// 1. Fetch current projection from Corteza
	contact, err := h.cortezaGateway.GetContact(ctx, cmd.TenantID, cmd.ContactID)
	if err != nil {
		return err
	}

	// 2. Pure Policy Validation (No side-effects)
	if err := h.policy.CanTransition(contact.LifecycleStage, cmd.TargetStage); err != nil {
		return err
	}
	if cmd.TargetStage == policy.LifecycleMQL {
		if err := h.policy.EnsureMQLReadiness(contact.Email, contact.Phone); err != nil {
			return err
		}
	}

	// 3. EXTERNAL API CALL: Update Corteza FIRST (Non-transactional)
	err = h.cortezaGateway.UpdateContactStage(ctx, cmd.TenantID, cmd.ContactID, cmd.TargetStage)
	if err != nil {
		// If Corteza fails, we return error. No local DB state is mutated. Safe.
		return err
	}

	// 4. LOCAL DB TRANSACTION: Record the successful transition via Outbox
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"contact_id":     cmd.ContactID,
		"old_stage":      contact.LifecycleStage,
		"new_stage":      cmd.TargetStage,
		"actor_id":       cmd.ActorID,
		"correlation_id": cmd.CorrelationID,
	})

	return h.dbTransaction.RunInTransaction(ctx, func(txCtx context.Context) error {
		// Event generation moved inside transaction closure
		event := outbox.Event{
			ID:           uuid.New().String(),
			EventType:    "crm.contact.lifecycle_changed",
			TenantID:     string(cmd.TenantID),
			AggregateID:  cmd.ContactID,
			Payload:      eventPayload,
			Status:       outbox.StatusPending,
			CreatedAt:    time.Now(),
		}

		// If this fails, Corteza is updated but Go lacks the event.
		// A background Reconciliation Worker will catch this mismatch later.
		return h.dbTransaction.SaveOutboxEvent(txCtx, event)
	})
}

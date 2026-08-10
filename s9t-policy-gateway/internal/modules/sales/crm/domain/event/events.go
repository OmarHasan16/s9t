package event

import (
	"time"

	"s9t.os/internal/modules/sales/crm/domain/valueobject"
	"s9t.os/internal/platform/types"
)

// Base Event Interface
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type ContactCreated struct {
	ContactID valueobject.ContactID
	TenantID  types.TenantID
	Timestamp time.Time
}
func (e ContactCreated) EventName() string { return "ContactCreated" }
func (e ContactCreated) OccurredAt() time.Time { return e.Timestamp }

type DealStageChanged struct {
	DealID     valueobject.DealID
	TenantID   types.TenantID
	OldStageID valueobject.PipelineStageID
	NewStageID valueobject.PipelineStageID
	Timestamp  time.Time
}
func (e DealStageChanged) EventName() string { return "DealStageChanged" }
func (e DealStageChanged) OccurredAt() time.Time { return e.Timestamp }

type DealWon struct {
	DealID    valueobject.DealID
	TenantID  types.TenantID
	Amount    valueobject.Money
	Timestamp time.Time
}
func (e DealWon) EventName() string { return "DealWon" }
func (e DealWon) OccurredAt() time.Time { return e.Timestamp }

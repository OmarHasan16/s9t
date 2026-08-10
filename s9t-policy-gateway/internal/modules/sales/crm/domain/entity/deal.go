package entity

import (
	"errors"
	"time"

	"s9t.os/internal/modules/sales/crm/domain/valueobject"
	"s9t.os/internal/platform/types"
)

var (
	ErrDealAlreadyClosed = errors.New("cannot modify a deal that is already closed (won/lost)")
	ErrInvalidStage      = errors.New("stage does not belong to the current pipeline")
)

type DealStatus string

const (
	DealStatusOpen DealStatus = "open"
	DealStatusWon  DealStatus = "won"
	DealStatusLost DealStatus = "lost"
)

type Deal struct {
	ID                valueobject.DealID
	TenantID          types.TenantID
	Name              string
	ContactID         *valueobject.ContactID
	CompanyID         *valueobject.CompanyID
	PipelineID        valueobject.PipelineID
	StageID           valueobject.PipelineStageID
	Amount            valueobject.Money
	ExpectedCloseDate *time.Time
	OwnerID           types.ActorID
	Status            DealStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ChangeStage safely moves a deal across the pipeline
func (d *Deal) ChangeStage(pipeline *Pipeline, targetStageID valueobject.PipelineStageID) error {
	if d.Status != DealStatusOpen {
		return ErrDealAlreadyClosed
	}
	if !pipeline.HasStage(targetStageID) {
		return ErrInvalidStage
	}

	d.StageID = targetStageID
	d.UpdatedAt = time.Now()
	return nil
}

// MarkWon transitions the deal to Won
func (d *Deal) MarkWon() error {
	if d.Status != DealStatusOpen {
		return ErrDealAlreadyClosed
	}
	d.Status = DealStatusWon
	d.UpdatedAt = time.Now()
	return nil
}

// MarkLost transitions the deal to Lost
func (d *Deal) MarkLost() error {
	if d.Status != DealStatusOpen {
		return ErrDealAlreadyClosed
	}
	d.Status = DealStatusLost
	d.UpdatedAt = time.Now()
	return nil
}

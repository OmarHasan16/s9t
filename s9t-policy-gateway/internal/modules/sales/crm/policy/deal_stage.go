package policy

import (
	"errors"
	"time"

	"s9t.os/internal/modules/sales/crm/domain/valueobject"
)

const (
	DealStageOpen        = "open"
	DealStageQualified   = "qualified"
	DealStageProposal    = "proposal"
	DealStageNegotiation = "negotiation"
	DealStageWon         = "closed_won"
	DealStageLost        = "closed_lost"
)

var (
	ErrInvalidDealTransition = errors.New("invalid deal stage transition")
	ErrMissingAssociation    = errors.New("deal must be associated with a contact or company to be won")
	ErrInvalidAmount         = errors.New("deal amount must be greater than zero to be won")
	ErrCloseDateRequired     = errors.New("expected close date is required")
	ErrDealIsClosed          = errors.New("deal is closed; must be explicitly reopened to 'open' first")
)

type DealPolicy struct{}

func NewDealPolicy() *DealPolicy {
	return &DealPolicy{}
}

// CanTransition evaluates strict pipeline rules
func (p *DealPolicy) CanTransition(current, target string) error {
	if current == target {
		return nil // Idempotent/no-op
	}

	// Rule: Closed deals cannot skip to intermediate stages. They must be reopened to 'open'.
	if (current == DealStageWon || current == DealStageLost) && target != DealStageOpen {
		return ErrDealIsClosed
	}

	// Forward progression matrix
	switch current {
	case DealStageOpen:
		if target != DealStageQualified && target != DealStageLost {
			return ErrInvalidDealTransition
		}
	case DealStageQualified:
		if target != DealStageProposal && target != DealStageLost {
			return ErrInvalidDealTransition
		}
	case DealStageProposal:
		if target != DealStageNegotiation && target != DealStageLost {
			return ErrInvalidDealTransition
		}
	case DealStageNegotiation:
		if target != DealStageWon && target != DealStageLost {
			return ErrInvalidDealTransition
		}
	case DealStageWon, DealStageLost:
		if target != DealStageOpen { // Explicit reopen rule
			return ErrInvalidDealTransition
		}
	default:
		return ErrInvalidDealTransition
	}

	return nil
}

// EnsureWonReadiness checks business constraints before a deal is marked as Won
func (p *DealPolicy) EnsureWonReadiness(contactID, companyID *string, amount valueobject.Money, closeDate *time.Time) error {
	if contactID == nil && companyID == nil {
		return ErrMissingAssociation
	}
	if amount.AmountMinor <= 0 {
		return ErrInvalidAmount
	}
	if closeDate == nil {
		return ErrCloseDateRequired
	}
	return nil
}

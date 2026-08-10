package valueobject

import "errors"

// Typed Identifiers (Protects against passing a DealID to a ContactID)
type ContactID string
type CompanyID string
type DealID string
type PipelineID string
type PipelineStageID string

// Domain Errors
var (
	ErrInvalidLifecycleTransition = errors.New("invalid lifecycle stage transition")
	ErrInvalidEmail               = errors.New("invalid email format")
	ErrNegativeAmount             = errors.New("money amount cannot be negative")
)

// Email Value Object (Self-validating)
type Email struct {
	Address string
}

func NewEmail(address string) (Email, error) {
	if address == "" /* add strict regex validation later */ {
		return Email{}, ErrInvalidEmail
	}
	return Email{Address: address}, nil
}

// Phone Value Object
type Phone struct {
	Number string
}

// Money Value Object (Avoids float precision issues)
type Money struct {
	AmountMinor int64  // e.g., cents
	Currency    string // e.g., "USD", "BDT"
}

func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{AmountMinor: amount, Currency: currency}, nil
}

// Percentage Value Object
type Percentage float64

// LifecycleStage Value Object & Rules
type LifecycleStage string

const (
	LifecycleSubscriber LifecycleStage = "subscriber"
	LifecycleLead       LifecycleStage = "lead"
	LifecycleMQL        LifecycleStage = "mql"
	LifecycleSQL        LifecycleStage = "sql"
	LifecycleCustomer   LifecycleStage = "customer"
)

// CanTransitionFrom enforces domain business rules for lifecycle changes
func (s LifecycleStage) CanTransitionFrom(current LifecycleStage) bool {
	// Example Rule: A Customer cannot be downgraded back to a Lead
	if current == LifecycleCustomer && s != LifecycleCustomer {
		return false
	}
	return true // Add stricter workflow rules as needed
}

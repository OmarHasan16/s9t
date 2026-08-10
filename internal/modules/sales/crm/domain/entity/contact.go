package entity

import (
	"time"

	"s9t.os/internal/modules/sales/crm/domain/valueobject"
	"s9t.os/internal/platform/types"
)

type Contact struct {
	ID             valueobject.ContactID
	TenantID       types.TenantID
	FirstName      string
	LastName       string
	Email          valueobject.Email
	Phone          valueobject.Phone
	LifecycleStage valueobject.LifecycleStage
	CompanyID      *valueobject.CompanyID
	OwnerID        *types.ActorID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// PromoteTo safely transitions the contact's lifecycle stage
func (c *Contact) PromoteTo(stage valueobject.LifecycleStage) error {
	if !stage.CanTransitionFrom(c.LifecycleStage) {
		return valueobject.ErrInvalidLifecycleTransition
	}
	c.LifecycleStage = stage
	c.UpdatedAt = time.Now()
	return nil
}

func (c *Contact) AssignToCompany(companyID valueobject.CompanyID) {
	c.CompanyID = &companyID
	c.UpdatedAt = time.Now()
}

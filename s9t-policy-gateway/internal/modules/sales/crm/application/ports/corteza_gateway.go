package ports

import (
	"context"
	"s9t.os/internal/platform/types"
)

// ContactProjection represents the data we care about from Corteza
type ContactProjection struct {
	ID             string
	LifecycleStage string
	Email          string
	Phone          string
	RecordVersion  int
}

// DealProjection represents the deal data from Corteza
type DealProjection struct {
	ID                string
	Stage             string
	RecordVersion     int
	ContactID         *string
	CompanyID         *string
	AmountMinor       int64
	Currency          string
	ExpectedCloseDate *string
}

// CortezaCRMGateway handles external API communications with Corteza
type CortezaCRMGateway interface {
	GetContact(ctx context.Context, tenantID types.TenantID, contactID string) (*ContactProjection, error)
	UpdateContactStage(ctx context.Context, tenantID types.TenantID, contactID string, stage string) error
	GetDeal(ctx context.Context, tenantID types.TenantID, dealID string) (*DealProjection, error)
	UpdateDealStage(ctx context.Context, tenantID types.TenantID, dealID string, targetStage string, expectedVersion int) (int, error)
}

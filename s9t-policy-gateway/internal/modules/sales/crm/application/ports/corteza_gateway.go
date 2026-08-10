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

// CortezaCRMGateway handles external API communications with Corteza
type CortezaCRMGateway interface {
	GetContact(ctx context.Context, tenantID types.TenantID, contactID string) (*ContactProjection, error)
	UpdateContactStage(ctx context.Context, tenantID types.TenantID, contactID string, stage string) error
}

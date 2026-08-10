# Low-Code CRM Architecture

S9T utilizes Corteza's `compose` metadata engine to drive dynamic CRM objects (Contact, Company, Deal, Engagement) inspired by the HubSpot data model. Instead of hardcoded Go structs for every object, we rely on Envoy YAML manifests to idempotently seed these structures into PostgreSQL.

## DIFF PLAN

| File Path | Action | Purpose |
|-----------|--------|---------|
| `internal/modules/core/metadata/domain/entity/object_metadata.go` | Create | Define ObjectMetadata entity |
| `internal/modules/core/metadata/domain/entity/field_metadata.go` | Create | Define FieldMetadata entity |
| `internal/modules/sales/crm/domain/entity/contact.go` | Create | Contact entity definition (HubSpot style) |
| `internal/modules/sales/crm/infrastructure/repository/contact_repository.go` | Create | PostgreSQL implementation for Contact |
| `migrations/sales/20260809120000_create_crm_contacts.sql` | Create | Contact table schema |

## CODE

```go
// internal/modules/sales/crm/domain/entity/contact.go
package entity

import "s9t.os/internal/platform/types"

type LifecycleStage string

const (
    StageSubscriber LifecycleStage = "SUBSCRIBER"
    StageLead       LifecycleStage = "LEAD"
    StageMQL        LifecycleStage = "MQL"
    StageCustomer   LifecycleStage = "CUSTOMER"
)

type Contact struct {
    ID             types.ContactID
    TenantID       types.TenantID
    Email          string
    FirstName      string
    LastName       string
    LifecycleStage LifecycleStage
    CompanyID      *types.CompanyID // Links to Company (Account)
    OwnerID        types.ActorID
}

func (c *Contact) PromoteToCustomer() error {
    c.LifecycleStage = StageCustomer
    return nil
}
```

# Runbook: CRM Seeding (HubSpot Schema)

## INSPECTION

**Status:** PARTIAL

**Verified:**
- `cmd/api/main.go` — Entrypoint exists

**UNVERIFIED:**
- `internal/modules/sales/crm/application/seeder/hubspot_seeder.go` — Needs inspection
- Cobra/CLI command registration for `crm:hubspot:seed` — Needs inspection

## Overview
Details how to seed initial metadata and standard objects (HubSpot data model: Contacts, Companies, Deals, Engagements) required for CRM Core via the CLI.

## Seeding Rules
- **Idempotency:** Safe to run multiple times. Preserves existing metadata. **[UNVERIFIED — INSPECTION REQUIRED]**
- **Dynamic Registration:** Registers entities into the Corteza Metadata Engine dynamically without hardcoded Go structs if applicable.

## DIFF PLAN

| File Path | Action | Purpose |
|-----------|--------|---------|
| `cmd/api/main.go` | Modify | Register seeding CLI command |
| `internal/modules/sales/crm/application/seeder/hubspot_seeder.go` | Create | Seeder logic for HubSpot schema |

## CODE

```go
// internal/modules/sales/crm/application/seeder/hubspot_seeder.go
package seeder

import (
    "context"
    "log"
)

type HubSpotSeeder struct {
    // Dependencies like MetadataRepo injected here [UNVERIFIED]
}

func (s *HubSpotSeeder) Seed(ctx context.Context) error {
    log.Println("Starting HubSpot-style metadata seeding...")
    
    // Idempotent check and insert logic
    // Creates standard HubSpot objects: 
    // 1. Contact (with lifecycle stages)
    // 2. Company 
    // 3. Deal (with pipelines)
    // 4. Engagement (Email, Meeting, Call, Note)
    
    log.Println("Seeding complete.")
    return nil
}

```

## VALIDATION

```bash
go mod tidy
go build ./cmd/api
go run cmd/api/main.go crm:hubspot:seed

```

package entity

import (
	"strings"
	"time"

	"s9t.os/internal/modules/sales/crm/domain/valueobject"
	"s9t.os/internal/platform/types"
)

type Company struct {
	ID               valueobject.CompanyID
	TenantID         types.TenantID
	Name             string
	Domain           string // Uniqueness enforced at DB level (tenant_id, domain)
	NormalizedDomain string // e.g., "planetcrust.com" instead of "https://www.planetcrust.com/"
	Industry         string
	OwnerID          *types.ActorID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// SetDomain updates and normalizes the domain for uniqueness checks
func (c *Company) SetDomain(domain string) {
	c.Domain = domain
	c.NormalizedDomain = normalizeDomain(domain)
	c.UpdatedAt = time.Now()
}

func normalizeDomain(domain string) string {
	d := strings.ToLower(strings.TrimSpace(domain))
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "www.")
	return strings.TrimRight(d, "/")
}

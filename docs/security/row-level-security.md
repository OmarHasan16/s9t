# Row-Level Security

Corteza's `pkg/rbac` implements row-level security (RLS) dynamically at the service layer by evaluating `rbac_rules` against the `recordID` and `moduleID`. This prevents unauthorized access to individual CRM records without hardcoding SQL RLS policies in PostgreSQL.

For environments enforcing Database-Level RLS (e.g. S9T direct DB integration), it is structured as follows:

```sql
-- migrations/sales/20260809120000_create_crm_contacts.sql
CREATE TABLE sales.crm_contacts (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    lifecycle_stage VARCHAR(50) NOT NULL DEFAULT 'LEAD'
);

-- Enable RLS
ALTER TABLE sales.crm_contacts ENABLE ROW LEVEL SECURITY;

-- Create Tenant Isolation Policy
CREATE POLICY tenant_isolation_policy ON sales.crm_contacts
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

```

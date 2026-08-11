CREATE TABLE platform.operation_ledger (
    operation_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    idempotency_key TEXT NOT NULL,
    command_hash TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    expected_version INT,
    target_state TEXT,
    correlation_id TEXT,
    actor_id TEXT,
    status TEXT NOT NULL, -- 'pending', 'processing', 'completed', 'retryable_failure', 'reconciliation_required'
    attempt_count INT DEFAULT 0,
    last_error TEXT,
    locked_at TIMESTAMP WITH TIME ZONE,
    locked_by TEXT,
    lease_token UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, idempotency_key)
);

ALTER TABLE platform.operation_ledger ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_policy ON platform.operation_ledger
    FOR ALL
    TO s9t_app
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

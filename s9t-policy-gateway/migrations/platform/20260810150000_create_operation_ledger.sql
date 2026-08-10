CREATE TABLE platform.operation_ledger (
    operation_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    idempotency_key TEXT NOT NULL,
    command_hash TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    expected_version INT,
    target_state TEXT,
    status TEXT NOT NULL, -- e.g., 'processing', 'completed', 'retryable_failure'
    attempt_count INT DEFAULT 0,
    last_error TEXT,
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

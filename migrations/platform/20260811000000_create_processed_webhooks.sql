CREATE TABLE platform.processed_webhook_events (
    tenant_id    UUID NOT NULL,
    event_id     VARCHAR(255) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, event_id)
);

-- Enable RLS
ALTER TABLE platform.processed_webhook_events ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY tenant_isolation_policy ON platform.processed_webhook_events
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

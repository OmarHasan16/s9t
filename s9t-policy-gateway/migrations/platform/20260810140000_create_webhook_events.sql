CREATE TABLE platform.webhook_events (
    tenant_id    UUID NOT NULL,
    event_id     VARCHAR(255) NOT NULL,
    event_type   VARCHAR(255) NOT NULL,
    status       VARCHAR(50) NOT NULL, -- 'processing', 'completed', 'failed'
    payload_hash VARCHAR(64) NOT NULL,
    received_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    locked_at    TIMESTAMP WITH TIME ZONE,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_error   TEXT,
    PRIMARY KEY (tenant_id, event_id)
);

ALTER TABLE platform.webhook_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY webhook_events_tenant_select ON platform.webhook_events
    FOR SELECT USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY webhook_events_tenant_insert ON platform.webhook_events
    FOR INSERT WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY webhook_events_tenant_update ON platform.webhook_events
    FOR UPDATE USING (tenant_id = current_setting('app.current_tenant', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);

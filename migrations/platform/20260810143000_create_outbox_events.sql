CREATE TABLE platform.outbox_events (
    id             UUID PRIMARY KEY,
    tenant_id      UUID NOT NULL,
    event_type     TEXT NOT NULL,
    aggregate_id   TEXT NOT NULL,
    payload        JSONB NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending',
    attempt_count  INTEGER NOT NULL DEFAULT 0,
    available_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ,
    last_error     TEXT
);

-- Enable RLS
ALTER TABLE platform.outbox_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY outbox_tenant_isolation ON platform.outbox_events
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid);
    
-- Indexes for Dispatcher performance
CREATE INDEX idx_outbox_pending ON platform.outbox_events (available_at) WHERE status = 'pending';

CREATE TABLE ws_tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID,
    org_id UUID,
    user_id VARCHAR(255) NOT NULL,
    role VARCHAR(100) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ws_tickets_expires_at ON ws_tickets(expires_at);

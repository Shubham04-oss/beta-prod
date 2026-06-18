-- Create custom enum for integration categories
CREATE TYPE integration_category AS ENUM ('commerce', 'crm', 'ticketing', 'ats', 'accounting');

-- Store Unified.to connection IDs mapped to our internal tenants
CREATE TABLE IF NOT EXISTS tenant_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    connection_id TEXT NOT NULL,
    category integration_category NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS on integrations
ALTER TABLE tenant_integrations ENABLE ROW LEVEL SECURITY;

CREATE POLICY "tenant_integrations_isolation" ON tenant_integrations
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

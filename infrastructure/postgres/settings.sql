-- Settings schema for tenant configurations
CREATE TABLE tenant_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id) UNIQUE,
    inventory_allocation_model VARCHAR(50) DEFAULT 'HARD', -- 'SOFT' or 'HARD'
    auto_po_enabled BOOLEAN DEFAULT false,
    default_low_stock_threshold INTEGER DEFAULT 10,
    costing_method VARCHAR(50) DEFAULT 'WAC', -- 'WAC' or 'FIFO'
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_tenant_settings ON tenant_settings 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

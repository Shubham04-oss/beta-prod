-- Phase 3: Unified Commerce Integration

-- 1. Commerce Connections mapping Tenant to external storefronts via Unified.to
CREATE TABLE IF NOT EXISTS commerce_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    unified_connection_id VARCHAR(255) NOT NULL,
    provider VARCHAR(100) NOT NULL, -- e.g., 'shopify', 'woocommerce'
    status VARCHAR(50) DEFAULT 'ACTIVE',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, unified_connection_id)
);

ALTER TABLE commerce_connections ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_commerce_connections ON commerce_connections 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_commerce_connections_tenant_id ON commerce_connections(tenant_id);

-- 2. Dead Letter Queue for Failed Syncs
CREATE TABLE IF NOT EXISTS sync_failures_dlq (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    connection_id UUID REFERENCES commerce_connections(id),
    entity_type VARCHAR(100) NOT NULL, -- e.g., 'product', 'inventory'
    entity_id UUID NOT NULL,
    payload JSONB NOT NULL,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'FAILED', -- 'FAILED', 'RETRYING', 'RESOLVED'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE sync_failures_dlq ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_sync_failures_dlq ON sync_failures_dlq 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_dlq_tenant_id ON sync_failures_dlq(tenant_id);
CREATE INDEX idx_dlq_status ON sync_failures_dlq(status);

-- 3. Mappings for Unified.to external IDs
CREATE TABLE IF NOT EXISTS commerce_item_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    connection_id UUID NOT NULL REFERENCES commerce_connections(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    unified_item_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(connection_id, product_id)
);

ALTER TABLE commerce_item_mappings ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_commerce_item_mappings ON commerce_item_mappings 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_commerce_item_mappings_tenant_id ON commerce_item_mappings(tenant_id);

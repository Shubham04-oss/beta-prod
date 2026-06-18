-- Attribute Groups
CREATE TABLE attribute_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, name)
);

ALTER TABLE attribute_groups ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_attribute_groups ON attribute_groups 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_attribute_groups_tenant_id ON attribute_groups(tenant_id);


-- Attribute Group Links (N:M mapping between groups and attributes)
CREATE TABLE attribute_group_links (
    group_id UUID REFERENCES attribute_groups(id) ON DELETE CASCADE,
    attribute_id UUID REFERENCES attributes(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    PRIMARY KEY (group_id, attribute_id)
);

ALTER TABLE attribute_group_links ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_attribute_group_links ON attribute_group_links 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_attribute_group_links_tenant_id ON attribute_group_links(tenant_id);


-- Bulk Jobs / Async Tasks
CREATE TABLE bulk_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    job_type VARCHAR(50) NOT NULL, -- e.g. 'BULK_UPDATE', 'IMPORT', 'EXPORT'
    status VARCHAR(50) DEFAULT 'PENDING', -- 'PENDING', 'RUNNING', 'COMPLETED', 'FAILED'
    payload_json JSONB,
    total_items INT DEFAULT 0,
    processed_items INT DEFAULT 0,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE bulk_jobs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_bulk_jobs ON bulk_jobs 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_bulk_jobs_tenant_id ON bulk_jobs(tenant_id);

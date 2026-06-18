-- PIM Module Schema (Enterprise EAV pattern)

-- 1. Product Types (Templates)
CREATE TABLE product_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

ALTER TABLE product_types ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_product_types ON product_types 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- 2. Global Attributes (e.g. "Color", "Size")
CREATE TABLE attributes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    type VARCHAR(50) DEFAULT 'TEXT', -- TEXT, BOOLEAN, NUMBER, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, slug)
);

ALTER TABLE attributes ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_attributes ON attributes 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- 3. Attribute Values (e.g. "Red", "Large")
CREATE TABLE attribute_values (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    attribute_id UUID NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    value VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(attribute_id, slug)
);

ALTER TABLE attribute_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_attribute_values ON attribute_values 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- 4. Product Type Attributes (Junction mapping Attributes to Product Types)
CREATE TABLE product_type_attributes (
    product_type_id UUID NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    attribute_id UUID NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    PRIMARY KEY (product_type_id, attribute_id)
);

ALTER TABLE product_type_attributes ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_product_type_attributes ON product_type_attributes 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- 5. Modify Products to link to Product Type (assuming products already exists in init.sql)
ALTER TABLE products ADD COLUMN IF NOT EXISTS product_type_id UUID REFERENCES product_types(id);

-- 6. Variant Attribute Values (Junction mapping a Variant to its specific Attribute Values)
CREATE TABLE variant_attribute_values (
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    attribute_value_id UUID NOT NULL REFERENCES attribute_values(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    PRIMARY KEY (variant_id, attribute_value_id)
);

ALTER TABLE variant_attribute_values ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_variant_attribute_values ON variant_attribute_values 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- 7. Mandatory Multi-Tenant RLS Indexes
CREATE INDEX idx_product_types_tenant_id ON product_types(tenant_id);
CREATE INDEX idx_attributes_tenant_id ON attributes(tenant_id);
CREATE INDEX idx_attribute_values_tenant_id ON attribute_values(tenant_id);
CREATE INDEX idx_product_type_attributes_tenant_id ON product_type_attributes(tenant_id);
CREATE INDEX idx_variant_attribute_values_tenant_id ON variant_attribute_values(tenant_id);

-- name: ListProductsByTenant :many
SELECT 
    p.id, 
    p.title, 
    p.sku, 
    p.category, 
    p.status, 
    p.updated_at
FROM products p
WHERE p.tenant_id = $1 AND p.deleted_at IS NULL
ORDER BY p.updated_at DESC
LIMIT 100;

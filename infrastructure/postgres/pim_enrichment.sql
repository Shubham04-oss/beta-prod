-- Phase 2: PIM Frontend Completeness Enrichment

-- 1. Merchandising, Lifecycle, Flags, SEO, and Quality Score on Products
ALTER TABLE products 
    ADD COLUMN IF NOT EXISTS short_description TEXT,
    ADD COLUMN IF NOT EXISTS long_description TEXT,
    ADD COLUMN IF NOT EXISTS category VARCHAR(255),
    ADD COLUMN IF NOT EXISTS brand VARCHAR(255),
    ADD COLUMN IF NOT EXISTS tags JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS launch_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS discontinue_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS warranty_period VARCHAR(100),
    ADD COLUMN IF NOT EXISTS is_taxable BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS is_returnable BOOLEAN DEFAULT true,
    ADD COLUMN IF NOT EXISTS requires_serial_number BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS seo_title VARCHAR(255),
    ADD COLUMN IF NOT EXISTS seo_description TEXT,
    ADD COLUMN IF NOT EXISTS seo_keywords JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS data_quality_score INTEGER DEFAULT 0;

-- 2. Data Quality Score on Variants
ALTER TABLE product_variants
    ADD COLUMN IF NOT EXISTS data_quality_score INTEGER DEFAULT 0;

-- 3. Product Media Table (Commercetools Standard: attachable to Product or Variant)
CREATE TABLE product_media (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    variant_id UUID REFERENCES product_variants(id) ON DELETE CASCADE,
    url VARCHAR(1024) NOT NULL,
    alt_text VARCHAR(255),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CHECK (product_id IS NOT NULL OR variant_id IS NOT NULL)
);

-- RLS for Product Media
ALTER TABLE product_media ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_product_media ON product_media 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE INDEX idx_media_product_id ON product_media(product_id);
CREATE INDEX idx_media_variant_id ON product_media(variant_id);
CREATE INDEX idx_media_tenant_id ON product_media(tenant_id);

-- 4. Validation Issues (Data Governance)
CREATE TABLE validation_issues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    issue_type VARCHAR(100) NOT NULL, -- e.g. 'MISSING_IMAGE', 'NO_CATEGORY'
    severity VARCHAR(50) DEFAULT 'WARNING', -- 'ERROR', 'WARNING', 'INFO'
    message TEXT NOT NULL,
    resolved BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE validation_issues ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_validation_issues ON validation_issues 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE INDEX idx_validation_issues_tenant_id ON validation_issues(tenant_id);
CREATE INDEX idx_validation_issues_product_id ON validation_issues(product_id);

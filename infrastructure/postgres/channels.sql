-- Sales Channels
CREATE TABLE sales_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    currency CHAR(3) DEFAULT 'USD',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product Channel Listings (Overrides)
CREATE TABLE product_channel_listings (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES sales_channels(id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    is_published BOOLEAN DEFAULT false,
    published_at TIMESTAMP WITH TIME ZONE,
    override_title VARCHAR(255),
    override_description TEXT,
    channel_price NUMERIC(10, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(product_id, channel_id)
);

-- RLS Setup
ALTER TABLE sales_channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE product_channel_listings ENABLE ROW LEVEL SECURITY;

-- Policies
CREATE POLICY tenant_isolation_sales_channels ON sales_channels 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE POLICY tenant_isolation_product_channel_listings ON product_channel_listings 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

-- Indexes
CREATE INDEX idx_sales_channels_tenant_id ON sales_channels(tenant_id);
CREATE INDEX idx_pcl_tenant_id ON product_channel_listings(tenant_id);
CREATE INDEX idx_pcl_channel_id ON product_channel_listings(channel_id);

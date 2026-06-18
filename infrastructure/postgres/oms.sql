-- OMS Database Schema
-- Run this after inventory.sql

-- ─── ENUMS ───────────────────────────────────────────────────────────────────
CREATE TYPE order_status AS ENUM (
    'draft', 'pending_payment', 'payment_authorized', 'confirmed',
    'processing', 'partially_fulfilled', 'fulfilled', 'delivered',
    'completed', 'cancelled', 'return_requested', 'partially_returned',
    'returned', 'refunded', 'partially_refunded', 'failed'
);
CREATE TYPE fulfillment_status AS ENUM (
    'pending', 'assigned', 'picked', 'packed', 'shipped',
    'out_for_delivery', 'delivered', 'cancelled', 'failed'
);
CREATE TYPE return_status AS ENUM (
    'requested', 'authorized', 'in_transit', 'received',
    'inspected', 'restocked', 'rejected', 'completed', 'cancelled'
);
CREATE TYPE return_item_disposition AS ENUM (
    'pending', 'sellable', 'damaged', 'refurbish', 'liquidate', 'scrap'
);
CREATE TYPE refund_status AS ENUM (
    'pending', 'processing', 'succeeded', 'failed', 'cancelled'
);
CREATE TYPE refund_reason AS ENUM (
    'customer_request', 'order_cancelled', 'return_received',
    'duplicate_charge', 'fraud', 'other'
);

-- ─── CUSTOMERS ───────────────────────────────────────────────────────────────
CREATE TABLE customers (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id      UUID NOT NULL REFERENCES organizations(id),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    email       TEXT NOT NULL,
    first_name  TEXT,
    last_name   TEXT,
    phone       TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE(org_id, tenant_id, email)
);
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_customers ON customers
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_customers_tenant ON customers(tenant_id, email) WHERE deleted_at IS NULL;

-- ─── ADDRESSES ───────────────────────────────────────────────────────────────
CREATE TABLE addresses (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id       UUID NOT NULL REFERENCES organizations(id),
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    customer_id  UUID REFERENCES customers(id),
    line1        TEXT NOT NULL,
    line2        TEXT,
    city         TEXT NOT NULL,
    state        TEXT,
    postal_code  TEXT NOT NULL,
    country_code CHAR(2) NOT NULL,
    is_default   BOOLEAN DEFAULT false,
    metadata     JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE addresses ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_addresses ON addresses
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_addresses_customer ON addresses(customer_id, tenant_id);

-- ─── ORDERS ──────────────────────────────────────────────────────────────────
CREATE TABLE orders (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    customer_id         UUID REFERENCES customers(id),
    -- address snapshots (FK to copies, not live address book)
    shipping_address_id UUID REFERENCES addresses(id),
    billing_address_id  UUID REFERENCES addresses(id),
    -- state
    status              order_status NOT NULL DEFAULT 'draft',
    -- financials
    currency            CHAR(3) NOT NULL DEFAULT 'USD',
    subtotal            NUMERIC(19,4) NOT NULL DEFAULT 0,
    discount_total      NUMERIC(19,4) NOT NULL DEFAULT 0,
    shipping_total      NUMERIC(19,4) NOT NULL DEFAULT 0,
    tax_total           NUMERIC(19,4) NOT NULL DEFAULT 0,
    total               NUMERIC(19,4) NOT NULL DEFAULT 0,
    -- payment tracking (decoupled from order status)
    payment_status      TEXT,          -- authorized | captured | voided | refunded
    payment_provider    TEXT,          -- stripe | paypal | stub
    payment_reference   TEXT,          -- external payment intent / charge ID
    -- channel / source
    channel             TEXT,          -- web | mobile | pos | b2b | api
    source_platform     TEXT,          -- shopify | woocommerce | custom (for Unified.to pulled orders)
    sales_channel_id    UUID REFERENCES sales_channels(id),
    -- metadata
    notes               TEXT,
    tags                TEXT[],
    metadata            JSONB,
    -- idempotency (prevents duplicate order on Temporal workflow retry)
    idempotency_key     TEXT UNIQUE,
    -- timestamps
    confirmed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ,
    fulfilled_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_orders ON orders
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_orders_tenant_status ON orders(tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_customer ON orders(customer_id, created_at DESC) WHERE deleted_at IS NULL;
-- Partial index for active orders — most queries target this set
CREATE INDEX idx_orders_active ON orders(tenant_id, updated_at DESC)
    WHERE status NOT IN ('completed','cancelled','refunded','failed') AND deleted_at IS NULL;

-- ─── ORDER LINE ITEMS ─────────────────────────────────────────────────────────
CREATE TABLE order_line_items (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id                  UUID NOT NULL REFERENCES organizations(id),
    tenant_id               UUID NOT NULL REFERENCES tenants(id),
    order_id                UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    variant_id              UUID,               -- soft ref, nullable (custom/manual items)
    -- PRICE + TITLE SNAPSHOT at purchase time — never reference live catalog
    sku                     TEXT,
    product_title           TEXT NOT NULL,
    variant_title           TEXT,
    unit_price              NUMERIC(19,4) NOT NULL,
    option_values           JSONB,              -- snapshot: { color: red, size: L }
    -- quantities (source of truth for fulfillment/return status)
    quantity                INT NOT NULL CHECK (quantity > 0),
    quantity_fulfilled      INT NOT NULL DEFAULT 0,
    quantity_returned       INT NOT NULL DEFAULT 0,
    quantity_cancelled      INT NOT NULL DEFAULT 0,
    -- per-line financials
    discount_total          NUMERIC(19,4) NOT NULL DEFAULT 0,
    tax_total               NUMERIC(19,4) NOT NULL DEFAULT 0,
    line_total              NUMERIC(19,4) NOT NULL,
    requires_shipping       BOOLEAN DEFAULT true,
    metadata                JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE order_line_items ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_order_line_items ON order_line_items
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_order_line_items_order ON order_line_items(order_id);
CREATE INDEX idx_order_line_items_variant ON order_line_items(variant_id) WHERE variant_id IS NOT NULL;

-- ─── FULFILLMENTS ─────────────────────────────────────────────────────────────
CREATE TABLE fulfillments (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    order_id            UUID NOT NULL REFERENCES orders(id),
    location_id         UUID REFERENCES inventory_locations(id), -- multi-location
    status              fulfillment_status NOT NULL DEFAULT 'pending',
    carrier             TEXT,
    tracking_number     TEXT,
    tracking_url        TEXT,
    shipped_at          TIMESTAMPTZ,
    delivered_at        TIMESTAMPTZ,
    estimated_delivery  TIMESTAMPTZ,
    metadata            JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    cancelled_at        TIMESTAMPTZ
);
ALTER TABLE fulfillments ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_fulfillments ON fulfillments
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_fulfillments_order ON fulfillments(order_id);
CREATE INDEX idx_fulfillments_status ON fulfillments(tenant_id, status);

-- ─── FULFILLMENT LINE ITEMS ───────────────────────────────────────────────────
CREATE TABLE fulfillment_line_items (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fulfillment_id      UUID NOT NULL REFERENCES fulfillments(id) ON DELETE CASCADE,
    order_line_item_id  UUID NOT NULL REFERENCES order_line_items(id),
    quantity            INT NOT NULL CHECK (quantity > 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_fli_fulfillment ON fulfillment_line_items(fulfillment_id);
CREATE INDEX idx_fli_order_line_item ON fulfillment_line_items(order_line_item_id);

-- ─── RETURNS ──────────────────────────────────────────────────────────────────
CREATE TABLE returns (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    order_id        UUID NOT NULL REFERENCES orders(id),
    rma_number      TEXT UNIQUE,
    status          return_status NOT NULL DEFAULT 'requested',
    reason          TEXT,
    notes           TEXT,
    refund_amount   NUMERIC(19,4),
    restocking_fee  NUMERIC(19,4) DEFAULT 0,
    received_at     TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE returns ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_returns ON returns
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_returns_order ON returns(order_id);
CREATE INDEX idx_returns_rma ON returns(rma_number) WHERE rma_number IS NOT NULL;

-- ─── RETURN LINE ITEMS ────────────────────────────────────────────────────────
CREATE TABLE return_line_items (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    return_id           UUID NOT NULL REFERENCES returns(id) ON DELETE CASCADE,
    order_line_item_id  UUID NOT NULL REFERENCES order_line_items(id),
    quantity            INT NOT NULL CHECK (quantity > 0),
    reason              TEXT,
    disposition         return_item_disposition DEFAULT 'pending',
    restocked_quantity  INT DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rli_return ON return_line_items(return_id);

-- ─── REFUNDS ──────────────────────────────────────────────────────────────────
CREATE TABLE refunds (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    order_id            UUID NOT NULL REFERENCES orders(id),
    return_id           UUID REFERENCES returns(id),
    status              refund_status NOT NULL DEFAULT 'pending',
    reason              refund_reason NOT NULL,
    amount              NUMERIC(19,4) NOT NULL CHECK (amount > 0),
    currency            CHAR(3) NOT NULL,
    payment_reference   TEXT,
    notes               TEXT,
    processed_at        TIMESTAMPTZ,
    failed_reason       TEXT,
    metadata            JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE refunds ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_refunds ON refunds
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_refunds_order ON refunds(order_id);
CREATE INDEX idx_refunds_pending ON refunds(status) WHERE status IN ('pending','processing');

-- ─── ORDER EVENTS (Audit Log) ─────────────────────────────────────────────────
CREATE TABLE order_events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id      UUID NOT NULL REFERENCES organizations(id),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    order_id    UUID NOT NULL REFERENCES orders(id),
    event_type  TEXT NOT NULL,     -- 'order.placed', 'order.cancelled', 'fulfillment.shipped', etc.
    actor_id    UUID,              -- user_id from authcontext — who triggered this
    actor_role  TEXT,              -- role from authcontext — ADMIN | STAFF | CUSTOMER | SYSTEM
    payload     JSONB,             -- snapshot of relevant data at event time
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE order_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_order_events ON order_events
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_order_events_order ON order_events(order_id, created_at DESC);
CREATE INDEX idx_order_events_actor ON order_events(actor_id) WHERE actor_id IS NOT NULL;

-- ─── OUTBOX EVENTS (For OMS) ──────────────────────────────────────────────────
CREATE TABLE oms_outbox_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    topic           TEXT NOT NULL,
    aggregate_id    UUID NOT NULL,
    aggregate_type  TEXT NOT NULL,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    org_id          UUID NOT NULL REFERENCES organizations(id),
    payload         JSONB NOT NULL,
    metadata        JSONB,           -- correlation_id, trace_id, actor_id
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    retry_count     INT NOT NULL DEFAULT 0
);
CREATE INDEX idx_oms_outbox_unpublished
    ON oms_outbox_events(created_at)
    WHERE published_at IS NULL AND failed_at IS NULL;

-- ─── UNIFIED ORDER MAPPINGS ───────────────────────────────────────────────────
CREATE TABLE commerce_order_mappings (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id              UUID NOT NULL REFERENCES organizations(id),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    connection_id       UUID NOT NULL REFERENCES commerce_connections(id),
    order_id            UUID NOT NULL REFERENCES orders(id),
    unified_order_id    TEXT NOT NULL,
    external_order_id   TEXT NOT NULL,
    last_synced_at      TIMESTAMPTZ,
    sync_direction      TEXT NOT NULL DEFAULT 'inbound', -- inbound | outbound | bidirectional
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (connection_id, external_order_id)
);
ALTER TABLE commerce_order_mappings ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation_order_mappings ON commerce_order_mappings
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
CREATE INDEX idx_order_mappings_order ON commerce_order_mappings(order_id);
CREATE INDEX idx_order_mappings_external ON commerce_order_mappings(connection_id, external_order_id);

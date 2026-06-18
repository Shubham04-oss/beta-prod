-- Phase 6: Inventory Ledger Migration

-- 1. Add cost_price to product_variants for accurate inventory valuation
ALTER TABLE product_variants ADD COLUMN IF NOT EXISTS cost_price NUMERIC(10, 2) DEFAULT 0.00;

-- 2. Create the immutable inventory ledger
DO $$ BEGIN
    CREATE TYPE inventory_transaction_type AS ENUM ('RESTOCK', 'SALE', 'RETURN', 'ADJUSTMENT', 'SHRINKAGE');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS inventory_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    transaction_type inventory_transaction_type NOT NULL,
    quantity_delta INTEGER NOT NULL,
    reference_id VARCHAR(255), -- Could be an Order ID or Manual Adjustment ID
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE inventory_ledger ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_inventory_ledger ON inventory_ledger;
CREATE POLICY tenant_isolation_inventory_ledger ON inventory_ledger 
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);

CREATE INDEX IF NOT EXISTS idx_inventory_ledger_tenant_id ON inventory_ledger(tenant_id);
CREATE INDEX IF NOT EXISTS idx_inventory_ledger_variant_id ON inventory_ledger(variant_id);

-- 3. Create a Trigger Function to automatically roll up the ledger into inventory_levels
CREATE OR REPLACE FUNCTION trigger_rollup_inventory_ledger()
RETURNS TRIGGER AS $$
BEGIN
    -- If it's a regular transaction, we update the available_quantity in inventory_levels
    INSERT INTO inventory_levels (
        org_id, 
        tenant_id, 
        variant_id, 
        location_id, 
        available_quantity, 
        reserved_quantity
    )
    VALUES (
        NEW.org_id,
        NEW.tenant_id,
        NEW.variant_id,
        NEW.location_id,
        NEW.quantity_delta,
        0
    )
    ON CONFLICT (variant_id, location_id) DO UPDATE 
    SET 
        available_quantity = inventory_levels.available_quantity + NEW.quantity_delta,
        updated_at = CURRENT_TIMESTAMP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop trigger if exists, then attach
DROP TRIGGER IF EXISTS trg_inventory_ledger_insert ON inventory_ledger;
CREATE TRIGGER trg_inventory_ledger_insert
AFTER INSERT ON inventory_ledger
FOR EACH ROW
EXECUTE FUNCTION trigger_rollup_inventory_ledger();

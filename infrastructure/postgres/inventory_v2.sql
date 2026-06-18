-- Phase 7: WAC and Procurement
ALTER TABLE inventory_ledger ADD COLUMN IF NOT EXISTS unit_cost NUMERIC(10, 2);

-- Update Trigger Function to calculate WAC
CREATE OR REPLACE FUNCTION trigger_rollup_inventory_ledger()
RETURNS TRIGGER AS $$
DECLARE
    current_qty INTEGER;
    current_cost NUMERIC(10, 2);
    new_wac NUMERIC(10, 2);
BEGIN
    -- WAC Calculation on RESTOCK
    IF NEW.transaction_type = 'RESTOCK' AND NEW.quantity_delta > 0 AND NEW.unit_cost IS NOT NULL THEN
        -- Get current quantity from inventory_levels across all locations for a true average (or just this location depending on accounting preference, we will use total available quantity)
        SELECT COALESCE(SUM(available_quantity), 0) INTO current_qty
        FROM inventory_levels 
        WHERE variant_id = NEW.variant_id;
        
        -- Get current cost_price
        SELECT COALESCE(cost_price, 0) INTO current_cost
        FROM product_variants WHERE id = NEW.variant_id;

        -- Calculate WAC
        IF (current_qty + NEW.quantity_delta) > 0 THEN
            new_wac := ((current_qty * current_cost) + (NEW.quantity_delta * NEW.unit_cost)) / (current_qty + NEW.quantity_delta);
            
            -- Update the variant
            UPDATE product_variants SET cost_price = new_wac WHERE id = NEW.variant_id;
        END IF;
    END IF;

    -- Update inventory_levels
    INSERT INTO inventory_levels (
        org_id, tenant_id, variant_id, location_id, available_quantity, reserved_quantity
    )
    VALUES (
        NEW.org_id, NEW.tenant_id, NEW.variant_id, NEW.location_id, NEW.quantity_delta, 0
    )
    ON CONFLICT (variant_id, location_id) DO UPDATE 
    SET 
        available_quantity = inventory_levels.available_quantity + NEW.quantity_delta,
        updated_at = CURRENT_TIMESTAMP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

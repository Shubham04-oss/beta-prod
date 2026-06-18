-- Core org+tenant RLS hardening for in-scope backend domains.
-- This file is intentionally idempotent so it can be applied to local or remote
-- development databases after the base schema.

DO $$
BEGIN
  IF to_regclass('public.orders') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_orders ON orders;
    DROP POLICY IF EXISTS tenant_org_isolation_orders ON orders;
    ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
    ALTER TABLE orders FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_orders ON orders
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.order_line_items') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_order_line_items ON order_line_items;
    DROP POLICY IF EXISTS tenant_org_isolation_order_line_items ON order_line_items;
    ALTER TABLE order_line_items ENABLE ROW LEVEL SECURITY;
    ALTER TABLE order_line_items FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_order_line_items ON order_line_items
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.fulfillments') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_fulfillments ON fulfillments;
    DROP POLICY IF EXISTS tenant_org_isolation_fulfillments ON fulfillments;
    ALTER TABLE fulfillments ENABLE ROW LEVEL SECURITY;
    ALTER TABLE fulfillments FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_fulfillments ON fulfillments
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.order_events') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_order_events ON order_events;
    DROP POLICY IF EXISTS tenant_org_isolation_order_events ON order_events;
    ALTER TABLE order_events ENABLE ROW LEVEL SECURITY;
    ALTER TABLE order_events FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_order_events ON order_events
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.returns') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_returns ON returns;
    DROP POLICY IF EXISTS tenant_org_isolation_returns ON returns;
    ALTER TABLE returns ENABLE ROW LEVEL SECURITY;
    ALTER TABLE returns FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_returns ON returns
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.refunds') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_refunds ON refunds;
    DROP POLICY IF EXISTS tenant_org_isolation_refunds ON refunds;
    ALTER TABLE refunds ENABLE ROW LEVEL SECURITY;
    ALTER TABLE refunds FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_refunds ON refunds
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.commerce_order_mappings') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_order_mappings ON commerce_order_mappings;
    DROP POLICY IF EXISTS tenant_org_isolation_commerce_order_mappings ON commerce_order_mappings;
    ALTER TABLE commerce_order_mappings ENABLE ROW LEVEL SECURITY;
    ALTER TABLE commerce_order_mappings FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_commerce_order_mappings ON commerce_order_mappings
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.sales_channels') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_sales_channels ON sales_channels;
    DROP POLICY IF EXISTS tenant_org_isolation_sales_channels ON sales_channels;
    ALTER TABLE sales_channels ENABLE ROW LEVEL SECURITY;
    ALTER TABLE sales_channels FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_sales_channels ON sales_channels
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.product_channel_listings') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_product_channel_listings ON product_channel_listings;
    DROP POLICY IF EXISTS tenant_org_isolation_product_channel_listings ON product_channel_listings;
    ALTER TABLE product_channel_listings ENABLE ROW LEVEL SECURITY;
    ALTER TABLE product_channel_listings FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_product_channel_listings ON product_channel_listings
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.tenant_settings') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_tenant_settings ON tenant_settings;
    DROP POLICY IF EXISTS tenant_org_isolation_tenant_settings ON tenant_settings;
    ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;
    ALTER TABLE tenant_settings FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_tenant_settings ON tenant_settings
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.commerce_connections') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_commerce_connections ON commerce_connections;
    DROP POLICY IF EXISTS tenant_org_isolation_commerce_connections ON commerce_connections;
    ALTER TABLE commerce_connections ENABLE ROW LEVEL SECURITY;
    ALTER TABLE commerce_connections FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_commerce_connections ON commerce_connections
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.sync_failures_dlq') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_sync_failures_dlq ON sync_failures_dlq;
    DROP POLICY IF EXISTS tenant_org_isolation_sync_failures_dlq ON sync_failures_dlq;
    ALTER TABLE sync_failures_dlq ENABLE ROW LEVEL SECURITY;
    ALTER TABLE sync_failures_dlq FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_sync_failures_dlq ON sync_failures_dlq
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.commerce_item_mappings') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_commerce_item_mappings ON commerce_item_mappings;
    DROP POLICY IF EXISTS tenant_org_isolation_commerce_item_mappings ON commerce_item_mappings;
    ALTER TABLE commerce_item_mappings ENABLE ROW LEVEL SECURITY;
    ALTER TABLE commerce_item_mappings FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_commerce_item_mappings ON commerce_item_mappings
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;

  IF to_regclass('public.audit_events') IS NOT NULL THEN
    DROP POLICY IF EXISTS tenant_isolation_audit_events ON audit_events;
    DROP POLICY IF EXISTS tenant_org_isolation_audit_events ON audit_events;
    ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
    ALTER TABLE audit_events FORCE ROW LEVEL SECURITY;
    CREATE POLICY tenant_org_isolation_audit_events ON audit_events
      FOR ALL
      USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      )
      WITH CHECK (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
      );
  END IF;
END $$;

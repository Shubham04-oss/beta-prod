package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const ddl = `
ALTER TABLE orders ADD COLUMN IF NOT EXISTS subtotal NUMERIC(19,4) NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount_total NUMERIC(19,4) NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_total NUMERIC(19,4) NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS tax_total NUMERIC(19,4) NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS total NUMERIC(19,4) NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

CREATE TABLE IF NOT EXISTS order_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    tenant_id   UUID NOT NULL,
    order_id    UUID NOT NULL,
    event_type  TEXT NOT NULL,
    actor_id    UUID,
    actor_role  TEXT,
    payload     JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS oms_outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic           TEXT NOT NULL,
    aggregate_id    UUID NOT NULL,
    aggregate_type  TEXT NOT NULL,
    tenant_id       UUID NOT NULL,
    org_id          UUID NOT NULL,
    payload         JSONB NOT NULL,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    retry_count     INT NOT NULL DEFAULT 0
);

CREATE OR REPLACE FUNCTION notify_outbox_event()

RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('outbox_insert', NEW.id::text);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_outbox_insert ON oms_outbox_events;

CREATE TRIGGER trigger_outbox_insert
AFTER INSERT ON oms_outbox_events
FOR EACH ROW
EXECUTE FUNCTION notify_outbox_event();
`

func main() {
	ctx := context.Background()
	connStr := "postgres://dev:dev@shubhams-mac-mini.local:5432/synq_db"

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, ddl)
	if err != nil {
		log.Fatalf("Failed to execute DDL: %v\n", err)
	}

	fmt.Println("Database schema updated successfully!")
}

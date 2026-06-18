# Postgres Row Level Security (RLS)

To support secure enterprise multi-tenancy, Synq enforces strict data isolation using PostgreSQL **Row Level Security (RLS)**. This ensures that one tenant can never read or write data belonging to another tenant, even if there is a bug in the application filtering logic.

## How it Works

Every tenant-scoped table in the database contains a `tenant_id` and an `org_id` column. We enable RLS on these tables and define policies that check these columns against PostgreSQL session configuration variables:

* `app.current_tenant` - The UUID of the active tenant.
* `app.current_org` - The UUID of the active organization.

### SQL Schema Definition

Below is an example of how RLS is configured on the `products` table in our `init.sql` schema:

```sql
-- Enable Row Level Security
ALTER TABLE products ENABLE ROW LEVEL SECURITY;

-- Create Policy for Tenant Isolation
CREATE POLICY tenant_isolation_policy ON products
    FOR ALL
    USING (tenant_id = NULLIF(current_setting('app.current_tenant', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.current_tenant', true), '')::uuid);
```

When a query is executed, PostgreSQL automatically appends these policy conditions to the query execution plan.

## Go Application Integration

In the Go backend (`ops-api`), we run all tenant-specific database operations within a transaction. The first action in the transaction must set the session variables.

### Enforcing RLS in Go

We implement a wrapper helper or run RLS-enforcing queries like this:

```go
tx, err := dbpool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback()

// Set the session configuration variable for the current transaction
_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", tenantID.String())
if err != nil {
    return err
}

// Any queries executed on tx will now be automatically filtered by the database policy!
rows, err := tx.Query(ctx, "SELECT sku, name FROM products")
```

This defense-in-depth security architecture protects against cross-tenant data leaks.

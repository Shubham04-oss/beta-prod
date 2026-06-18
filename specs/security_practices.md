# Security Practices: GCP-Native System

This document outlines the definitive security specification and practices adopted for our GCP-native system. Our architecture is designed with security at its core, leveraging strong typing, secure connection management, and database-level isolation to ensure data integrity and confidentiality across our multi-tenant platform.

## 1. Compile-Time Safety and SQL Injection Prevention with `sqlc`

To interact with our PostgreSQL database, we use **`sqlc`**, a compiler that generates type-safe Go code from SQL. 

### Why `sqlc`?
- **Zero SQL Injection:** `sqlc` relies exclusively on parameterized queries. Because the SQL is written explicitly and parameters are strictly bound by the driver, the risk of SQL injection is fundamentally eliminated. There is no dynamic query building that concatenates strings.
- **Compile-Time Guarantees:** Queries are validated against the actual database schema during the build process. If a column name changes or a type is mismatched, the build fails. This ensures our application code is always perfectly synchronized with the database structure, preventing runtime SQL errors and unexpected behavior.

## 2. Secure Connection Pooling with `pgx/v5`

For our database driver and connection pooling, we rely on **`pgx/v5`**. 

### Key Security Benefits:
- **Native PostgreSQL Features:** `pgx/v5` is built specifically for PostgreSQL, supporting advanced features like LISTEN/NOTIFY and complex data types securely and efficiently.
- **Robust Connection Pooling:** The `pgxpool` package provides highly concurrent and thread-safe connection pooling. This protects our database from connection exhaustion under high load (preventing DoS scenarios).
- **TLS by Default:** Connections to our Cloud SQL instances enforce TLS encryption in transit, guaranteeing that sensitive data cannot be intercepted between our application and the database.

## 3. Multi-Tenant Data Isolation with Postgres Row-Level Security (RLS)

In a multi-tenant environment, enforcing strict data isolation between tenants is critical. We utilize **PostgreSQL Row-Level Security (RLS)** to enforce access policies directly at the database layer. This means that even if a query is malformed or an application bug occurs, the database will strictly prohibit cross-tenant data access.

### The Go Transaction Interceptor

To ensure that every database operation is executed within the correct tenant context, we wrap our database interactions in a Go transaction interceptor. This interceptor automatically injects the tenant ID into the transaction context before any queries are executed.

#### Conceptual Code Example

Below is a conceptual example of how we implement the Go transaction interceptor to enforce RLS.

```go
package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunInTenantContext executes a given function within a database transaction,
// setting the current_tenant_id for Postgres Row-Level Security (RLS).
func RunInTenantContext(ctx context.Context, pool *pgxpool.Pool, tenantID string, fn func(tx pgx.Tx) error) error {
	// Begin a new transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Ensure the transaction is rolled back if it hasn't been committed
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			log.Printf("failed to rollback transaction: %v", rollbackErr)
		}
	}()

	// Inject the tenant ID into the Postgres session state for RLS
	// We use set_config to set a custom configuration parameter 'app.current_tenant_id'
	// The 'true' argument means the setting is local to this transaction.
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context for RLS: %w", err)
	}

	// Execute the business logic function within the configured transaction
	if err := fn(tx); err != nil {
		// The defer block will handle the rollback
		return err
	}

	// Commit the transaction if successful
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
```

### How It Works:
1. **Transaction Initialization:** A new database transaction is started.
2. **Context Injection:** `set_config('app.current_tenant_id', $1, true)` is executed. This securely binds the `tenantID` to the current transaction scope.
3. **RLS Policy Enforcement:** PostgreSQL evaluates RLS policies attached to tables (e.g., `CREATE POLICY tenant_isolation_policy ON users USING (tenant_id = current_setting('app.current_tenant_id')::uuid);`).
4. **Execution & Commit:** The application logic runs safely. If the query attempts to access records belonging to another tenant, PostgreSQL will silently filter them out or deny access based on the policy.

By combining `sqlc`, `pgx/v5`, and strict RLS enforced via our Go interceptor, we achieve a defense-in-depth security posture suitable for a robust, multi-tenant GCP-native application.

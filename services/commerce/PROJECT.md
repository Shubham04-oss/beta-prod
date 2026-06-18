# Project: Commerce Modules

## Architecture
- **PIM Module**: Manages Products and Product Variants.
- **OMS Module**: Manages Customers, Orders, and Order Line Items.
- **Inventory Module**: Manages Locations and Inventory Levels.
- **Unified.to Integration**: Synchronizes data with third-party commerce integrations.
- **Database**: PostgreSQL with `pgvector` for semantic search, configured with Row Level Security per tenant/org.

## Code Layout
- `/cmd/api`: Main API application entry points.
- `/internal/db`: Database connection and utilities.
- `/internal/models`: Go struct representations of `init.sql`.
- `/internal/pim`: Product Information Manager logic.
- `/internal/oms`: Order Management System logic.
- `/internal/inventory`: Inventory Ledgers logic.
- `/internal/unified`: Unified.to API integration logic.
- `/pkg/...`: Shared utilities if any.

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| 1 | Database & Models | Map `init.sql` to Go structs, setup DB connection | none | DONE |
| 2 | PIM Module | CRUD operations for Products and Variants | M1 | FAILED |
| 3 | OMS Module | CRUD operations for Customers and Orders | M1 | DONE |
| 4 | Inventory Module | Logic for Locations, Inventory levels, and reservations | M1 | FAILED |
| 5 | Unified.to Integration | Data synchronization logic via Unified.to API | M2, M3, M4 | FAILED |
| 6 | E2E Testing Pass | Ensure all E2E tests pass (Final Milestone Phase 1) | M1-M5, E2E | FAILED |

## Interface Contracts
### PIM ↔ Inventory
- Creating products/variants initializes inventory records.
- Deleting products cascades or handles inventory gracefully.

### OMS ↔ Inventory
- Creating orders triggers inventory reservation.
- Fulfilling orders deducts inventory.

### Unified.to ↔ PIM/OMS/Inventory
- Sync products, orders, and inventory to/from external systems via standard API payload structures.

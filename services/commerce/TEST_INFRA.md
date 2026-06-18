# E2E Test Infra: Commerce Modules

## Test Philosophy
- Opaque-box, requirement-driven. No dependency on implementation design.
- Methodology: Category-Partition + BVA + Pairwise + Workload Testing.
- Framework: Go's built-in testing framework (`go test`), validating database changes (against `init.sql` schemas) and mock integration with unified.to.

## Feature Inventory
| # | Feature | Source (requirement) | Tier 1 | Tier 2 | Tier 3 |
|---|---------|---------------------|:------:|:------:|:------:|
| 1 | PIM (Product Information Manager) | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ |
| 2 | OMS (Order Management System) | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ |
| 3 | Inventory Ledgers | ORIGINAL_REQUEST §R1 | 5 | 5 | ✓ |
| 4 | Unified.to Integration Sync | ORIGINAL_REQUEST §R2 | 5 | 5 | ✓ |

## Test Architecture
- Test runner: `go test -v ./tests/e2e/...`
- Test suite validates core logic paths and real SQL transactions on the provided schema.
- Uses standard Go testing techniques (e.g., testcontainers or mock db setup, mock HTTP servers for Unified.to).
- Directory layout: `/tests/e2e` for E2E tests, structured by Tiers.

## Real-World Application Scenarios (Tier 4)
| # | Scenario | Features Exercised | Complexity |
|---|----------|--------------------|------------|
| 1 | Full Product Lifecycle | F1, F3, F4 | Medium |
| 2 | End-to-End Order Fulfillment | F1, F2, F3 | High |
| 3 | Bulk Inventory Sync from External | F3, F4 | Medium |
| 4 | Out-of-Stock Order Handling | F2, F3 | Medium |
| 5 | Multi-Tenant Data Isolation Check | F1, F2, F3, F4 | High |

## Coverage Thresholds
- Tier 1: ≥5 per feature (20 total)
- Tier 2: ≥5 per feature (20 total)
- Tier 3: pairwise coverage of major feature interactions (≥6 total)
- Tier 4: ≥5 realistic application scenarios

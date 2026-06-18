# Inventory Domain Architecture & Development Plan

## 1. Executive Summary & User Philosophy

The Inventory domain is the beating heart of commerce operations. As a system, it must be relentlessly reliable; as a product, it must be deeply customizable. Our architecture is designed strictly around the **User's Point of View (POV)**. The system does not dictate rigid business logic; instead, it provides a robust engine that executes the user's bespoke configurations. 

**Key User-Centric Capabilities:**
- **Deep User Customizability:** Every merchant operates differently. The system prioritizes exposing control levers directly to the UI.
- **User-Defined Safety Stock Rules:** Granular controls allowing users to set dynamic buffers, seasonal thresholds, and SKU-level safety nets to prevent overselling.
- **Custom Multi-Location Routing:** A powerful UI that lets users define exactly how orders are routed across warehouses, 3PLs, and retail stores based on geo-proximity, cost, or priority.

## 2. Technology Stack & Tenets

This architecture is built strictly on our existing, non-negotiable tech stack:
- **Backend:** Go (with `chi` router).
- **Database:** PostgreSQL (with `sqlc` and `pgvector`).
- **Orchestration:** Temporal.
- **Frontend:** Next.js (App Router) with Firebase Auth.
- **Integrations:** Unified.to.
- **Infrastructure:** GCP Native (Cloud Run/GKE, Pub/Sub, Eventarc, Secret Manager, Memorystore).

## 3. Data Architecture: Postgres & Event-Sourced Ledgers

Inventory is financial data. We do not simply mutate an `available_quantity` integer. We utilize **Event-Sourced Ledgers** backed by PostgreSQL.

### 3.1 The Ledger Model
All stock movements are immutable entries in an `inventory_ledger` table.
- **Appends Only:** Every allocation, reservation, fulfillment, or manual adjustment is an appended row containing the delta (e.g., `+10`, `-2`), a timestamp, an idempotency key, and a reference ID (Order ID, PO ID).
- **Current State:** The actual on-hand or available inventory is a real-time aggregation of these ledger entries. We can use PostgreSQL Materialized Views or trigger-based rollups to maintain highly performant `current_inventory` snapshots.
- **`sqlc` Integration:** We maintain type safety by writing pure, optimized SQL queries for ledger inserts and aggregations, allowing `sqlc` to generate performant Go structures.

### 3.2 Concurrency & Locking
To prevent overselling during high-throughput flash sales:
- Use **Pessimistic Locking** (`SELECT ... FOR UPDATE SKIP LOCKED`) in Postgres when actively reserving stock for an in-flight order.
- This ensures that concurrent Go routines (or Temporal workers) processing checkouts for the same highly-contested SKU do not step on each other, maintaining strict ledger integrity.

## 4. Orchestration: Temporal as the Engine

Temporal is the eternal orchestrator of the Inventory domain. No complex state machine or distributed process should live in a standard Go HTTP handler.

### 4.1 Order Routing & Fulfillment Sagas
When an order arrives, an `InventoryFulfillmentWorkflow` is spawned:
1. **Rule Evaluation:** The workflow retrieves the user's custom multi-location routing rules (e.g., "Ship from closest warehouse, fallback to primary hub").
2. **Saga Pattern:** It attempts to reserve inventory in the DB (Activity 1). If successful, it proceeds to instruct the warehouse (Activity 2).
3. **Compensation:** If warehouse processing fails or payment is declined, the Temporal Saga automatically executes compensating activities to release the reserved ledger entries back to the pool.

### 4.2 Automated Safety Stock & Reordering Workflows
- **Cron Workflows:** Temporal schedules periodic workflows that evaluate current stock against user-defined Safety Stock Rules.
- **Dynamic Action:** If stock dips below a customized threshold, the workflow can autonomously trigger a `PurchaseOrderWorkflow` or alert the user via external channels.

### 4.3 Third-Party Synchronization
- **Sync Workflows:** Inventory levels across external channels (Shopify, Amazon) are synced via **Unified.to**. 
- Temporal workflows manage these syncs, inherently handling API rate limiting, pagination, and exponential backoff without cluttering the Go application logic.

## 5. Frontend UI/UX: Next.js & Firebase

The frontend must empower the user to leverage this powerful backend effortlessly.

- **Authentication & Security:** Firebase Auth secures access to the Next.js App Router endpoints.
- **Routing Engine UI:** A visual, intuitive interface (potentially node-based or ordered rule lists) allowing merchants to build complex multi-location routing trees. 
- **Safety Stock Dashboard:** Rich data tables allowing bulk editing of safety thresholds, buffer rules, and priority levels.
- **Real-Time Visibility:** Next.js Server Components query the Go API (which in turn queries Postgres and Temporal state) to provide real-time visibility into running syncs. Temporal's querying capabilities allow the UI to show exactly where an inventory sync workflow is at any given moment.

## 6. Integrations & Eventing (GCP)

- **Unified.to:** Acts as our single normalized integration layer for omni-channel inventory. Go backend communicates with Unified.to APIs, wrapped in Temporal activities for resilience.
- **Event-Driven Architecture (Pub/Sub):** When the `inventory_ledger` aggregates a change that breaches a threshold, or when stock becomes zero, the Go backend publishes events to **GCP Pub/Sub**.
- **Eventarc & Memorystore:** Eventarc routes these Pub/Sub messages to other microservices (e.g., Catalog, Search) which update their fast-read layers in **Memorystore (Redis)**. This ensures that the storefront frontend never queries Postgres for stock levels, maintaining extreme high-read performance.


## Frontend & UI Constraints (STRICT)
**CRITICAL RULE:** The overarching frontend design layout is **FIXED**. We will absolutely **NOT** alter the core layout structure.
*   **Strict Design Language:** We will only polish custom components using existing libraries to respect the current design language.
*   **Right-Sidebar Standard:** Every domain must utilize the standard right-sidebar drawer for contextual information.
*   **Visual Polish:** We will utilize approved visual effects (like background blur and glassmorphism on modals/sidebars) to ensure the application feels premium without breaking the underlying grid or routing structure.

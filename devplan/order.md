# Order Management System (OMS) - Development Plan

## 1. Architectural Vision & Design Philosophy

As a Senior Principal Engineer, the goal is to architect an Order Management System (OMS) that is not just technically resilient, but fundamentally **user-empowering**. Merchants and operators need deep customizability without writing code. They need to configure fulfillment workflows, tailor return policies, and manage their operations through an intuitive UI. 

We will build this on our established, rock-solid foundation. We will **not** reinvent the wheel with new databases or orchestrators; instead, we will deeply leverage our existing stack to solve complex OMS challenges.

### Core Technology Stack (Immutable Foundation)
* **Backend:** Go (chi router) for high-performance, strictly-typed APIs.
* **Database:** PostgreSQL (pgvector, sqlc). Relational integrity is paramount for financial and order data. JSONB will provide the flexibility needed for user-defined custom fields.
* **Orchestration:** Temporal. The eternal orchestrator. All state machines, sagas, and long-running processes live here.
* **Frontend:** Next.js (App Router) + Firebase Auth for a highly responsive, SSR/SSG capable admin dashboard.
* **Integrations:** Unified.to for unifying external APIs (Shopify, ERPs, WMS).
* **Cloud:** GCP Native (Cloud Run/GKE, Pub/Sub, Eventarc, Secret Manager).

---

## 2. The User's POV: Deep Customizability

The architecture must serve the user's need to control their business logic. We achieve this by making the backend configuration-driven, orchestrated by Temporal, and managed via Next.js.

### 2.1 Configurable Fulfillment Workflows
Users need to route orders based on rules (e.g., "If order value > $1000, require manual fraud review", "If item is hazardous, route to Warehouse B").
* **UX:** Next.js provides a rule-builder UI.
* **Backend:** Go APIs save these rules as JSON in PostgreSQL.
* **Execution:** Temporal workflows read these rules at runtime. The `OrderFulfillmentWorkflow` branches dynamically based on the merchant's configured rules.

### 2.2 Custom Return Policies (RMA)
Return policies vary wildly (e.g., "30 days for electronics, 14 days for apparel, restocking fee if opened").
* **Implementation:** We will use Postgres to store policy configurations. When an RMA is initiated, a Go service evaluates the order against the policy. If approved, a Temporal `ReturnWorkflow` is triggered to handle shipping label generation, inventory restocking, and refund processing.

### 2.3 Robust UI/UX Settings & Visibility
Users need to know *exactly* where an order is.
* **Implementation:** Temporal allows us to query the state of any running workflow instantly. We will expose Temporal Queries through our Go API to the Next.js frontend, providing real-time, granular order tracking without complex database polling.

---

## 3. Deep Dive: Using the Stack for OMS

### 3.1 Orchestration: Temporal for Sagas & State Machines
Order processing is inherently a Saga. It involves distributed transactions: reserve inventory, charge card, notify warehouse.
* **The `OrderLifecycleWorkflow`:** A single Temporal workflow instance is created per order.
* **Activities:** `ReserveInventoryActivity`, `ChargePaymentActivity`, `TransmitToWMSActivity`.
* **Compensation:** If payment fails after inventory is reserved, Temporal automatically executes the `ReleaseInventoryActivity` compensating transaction.
* **Long-Running Pauses:** If an order is backordered, the Temporal workflow simply `Sleeps` or waits for an `InventoryRestockedSignal`. No cron jobs, no polling.

### 3.2 Database: PostgreSQL + sqlc
* **Schema Design:** We will heavily normalize core transactional data (Orders, LineItems, Payments) for ACID compliance.
* **sqlc:** All Go database access will be generated via sqlc to ensure compile-time type safety against our SQL queries.
* **Custom Fields:** We will utilize Postgres `JSONB` columns on the `Orders` table to store unstructured data from external channels (via Unified.to) or custom merchant attributes without schema migrations.
* **Search:** `pgvector` can be leveraged later for semantic search over order histories or customer support logs.

### 3.3 Backend: Go + chi
* **API Gateway to Temporal:** Go serves as the translation layer. It receives HTTP requests (e.g., `POST /orders`), validates the payload, and starts the Temporal workflow.
* **Webhooks:** Go/chi will handle incoming webhooks from Unified.to (e.g., an order update from Shopify) and translate them into Temporal Signals (e.g., `SendSignal("OrderUpdated", data)`).

### 3.4 Cloud Native: GCP Pub/Sub & Eventarc
* When a Temporal workflow completes a major milestone (e.g., "Order Shipped"), it publishes an event to GCP Pub/Sub.
* Other domains (like Analytics or Notifications) subscribe to these events, ensuring the OMS remains decoupled from downstream side-effects.

---

## 4. Implementation Roadmap

### Phase 1: Data & API Foundation (Weeks 1-2)
* Define Postgres schemas (`orders`, `line_items`, `fulfillment_rules`, `return_policies`).
* Generate Go data access layer using `sqlc`.
* Scaffold Go/chi REST endpoints for CRUD operations on configuration data.

### Phase 2: Temporal Sagas (Weeks 3-5)
* Implement the core `OrderLifecycleWorkflow`.
* Implement Activities for inventory, payment, and external system sync.
* Implement the Compensating Transactions for the Saga pattern.
* Expose Temporal Queries/Signals via Go APIs.

### Phase 3: Integration & Eventing (Weeks 6-7)
* Integrate Unified.to for inbound order ingestion and outbound fulfillment syncing.
* Set up GCP Pub/Sub topics for domain events (`order.created`, `order.shipped`, `order.returned`).

### Phase 4: Frontend Customization (Weeks 8-10)
* Build the Next.js App Router admin dashboard.
* Develop the interactive UI for configuring fulfillment workflows and return policies.
* Implement real-time order tracking using Server-Sent Events (SSE) or polling via Next.js Server Actions calling the Go/Temporal API.


## Frontend & UI Constraints (STRICT)
**CRITICAL RULE:** The overarching frontend design layout is **FIXED**. We will absolutely **NOT** alter the core layout structure.
*   **Strict Design Language:** We will only polish custom components using existing libraries to respect the current design language.
*   **Right-Sidebar Standard:** Every domain must utilize the standard right-sidebar drawer for contextual information.
*   **Visual Polish:** We will utilize approved visual effects (like background blur and glassmorphism on modals/sidebars) to ensure the application feels premium without breaking the underlying grid or routing structure.

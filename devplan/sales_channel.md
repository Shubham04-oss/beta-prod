# Architecture & Development Plan: Sales Channel Domain

## 1. Executive Summary & User POV
The Sales Channel domain empowers users to seamlessly connect, map, and synchronize their products, inventory, and orders across multiple sales channels (e.g., Shopify, Amazon, WooCommerce) via **Unified.to**. From the **User's POV**, this system must feel deeply customizable and robust:
- **Custom Channel Mapping UIs:** Users can visually map their internal product catalog fields to channel-specific schemas, ensuring they are always in control of how their data is presented.
- **User-Configurable Webhooks:** Users can define exactly which events (e.g., "Order Created", "Inventory Depleted") trigger external webhooks to their internal systems or third-party tools.
- **Robust UI/UX Settings:** Users have granular control over sync frequency, conflict resolution rules (e.g., "System of Record vs. Sales Channel override"), and alert thresholds for when syncing encounters errors.

## 2. Core Technology Stack
We are strictly building on top of our existing robust architecture. No core components are being replaced.
- **Frontend:** Next.js (App Router) + Firebase Auth for secure, responsive, and dynamic user interfaces.
- **Backend Core:** Go with the `chi` router for high-performance, lightweight API endpoints.
- **Database:** PostgreSQL (with `pgvector` and `sqlc`). Postgres is the absolute source of truth.
- **Orchestration:** Temporal is our eternal orchestrator for all data synchronization, long-running processes, state machines, and failure recoveries.
- **Integration Layer:** Unified.to is used to abstract and unify third-party ecommerce API connections.
- **Cloud Infrastructure:** GCP Native (Cloud Run/GKE, Pub/Sub, Eventarc, Secret Manager, Memorystore).

## 3. Architecture & Component Design

### 3.1 Omni-Channel Mapping & Sync Engine
1. **Unified.to Integration:**
   - Unified.to acts as the unified API for all connected sales channels.
   - The Go backend manages user OAuth connections via Unified.to and securely stores connection IDs/tokens in Postgres.
2. **Temporal Orchestration (The Sync Heartbeat):**
   - **Workflows:** Temporal workflows orchestrate bidirectional data syncing (e.g., pulling orders, pushing inventory updates). Temporal ensures that if a Unified.to rate limit is hit, or the channel API goes down, the workflow gracefully pauses, waits, and retries without dropping data.
   - **Sagas:** Multi-step processes (e.g., "Publish Product to 5 Channels") are managed as Temporal Sagas. If publishing fails on one channel, Temporal handles compensating transactions or alerts the user via the UI while allowing successful channels to proceed.
   - **State Machines:** Temporal manages the state of each connected channel (e.g., `Syncing`, `Idle`, `Error State`).
3. **PostgreSQL (System of Record):**
   - User configurations, mapping rules, webhook definitions, and local catalog data reside here.
   - `sqlc` is used to generate type-safe Go structs for interacting with tenant mapping configurations.
   - `pgvector` will be leveraged to provide intelligent "auto-mapping" suggestions (e.g., matching a user's custom "Color" attribute to a channel's "Variant Color" field using semantic similarity to enhance the UX).

### 3.2 Backend Implementation (Go + Chi)
- **API Layer:** The `chi` router exposes RESTful endpoints for the Next.js frontend to fetch configurations, trigger manual syncs, and update field mappings.
- **Webhook Receivers:** The Go API exposes endpoints to receive incoming webhooks from Unified.to (e.g., an order arrives from Shopify).
- **Event Dispatching:** Incoming webhooks are validated and pushed to GCP Pub/Sub. Temporal Workers listen to these events (or are triggered via API) to kick off the appropriate processing workflows.

### 3.3 Frontend Implementation (Next.js + Firebase Auth)
- **Deep Customizability:** The Next.js UI features a robust "Mapping Builder". Users map internal schema attributes to Unified.to schema attributes with immediate validation.
- **Real-time Feedback:** Using Server-Sent Events (SSE) or WebSockets integrated with the Go backend, the UI reflects real-time Temporal workflow statuses (e.g., showing a live progress bar as 10,000 products are pushed to a channel).
- **Settings Dashboard:** A comprehensive dashboard where users configure retry policies, sync intervals, webhook endpoints, and notification preferences.

## 4. Execution Plan & Phases

### Phase 1: Foundation & Connections
- **Frontend:** Build the "Connect Channels" UI using Unified.to's embedded authorization widget within Next.js.
- **Backend (Go/Chi):** Create API endpoints to securely handle the Unified.to connection handshake and store connection metadata.
- **Database:** Define the Postgres schema using `sqlc` for storing `Tenant`, `ChannelConnection`, and `SyncConfiguration`.

### Phase 2: Core Data Sync (Temporal)
- **Temporal Workers (Go):** Implement the foundational Temporal workflows:
  - `PullOrdersWorkflow`
  - `PushInventoryWorkflow`
  - `SyncCatalogWorkflow`
- **Error Handling:** Implement robust retry policies, rate-limit handling, and circuit breakers inside Temporal activities, wrapping Unified.to API calls.

### Phase 3: Custom Mapping UI & Engine
- **Frontend:** Develop the visual mapping interface.
- **Database:** Store these mapping rules securely in Postgres (likely leveraging JSONB for flexible schema mapping).
- **AI Assist (pgvector):** Implement a Go service that compares user schemas with Unified.to standard schemas and auto-suggests mappings using `pgvector`.

### Phase 4: User-Configurable Webhooks & Automation
- **Backend:** Build the outbound webhook delivery system. Users register destination URLs and subscribe to specific topics.
- **Orchestration:** When Temporal finishes a sync workflow or detects a new order, it triggers a `DispatchWebhookWorkflow` that handles exponential backoff if the user's receiving endpoint is down.
- **Frontend:** Build the webhook configuration screen and the webhook delivery logs UI so users can debug their integrations.

## 5. Deployment & Infrastructure (GCP)
- Deploy the Go Chi API and Temporal Workers to **Google Cloud Run** or **GKE**.
- Use **Cloud SQL for PostgreSQL** for transactional data persistence and pgvector.
- Store Unified.to API keys and Webhook signing secrets in **Secret Manager**.
- Use **Memorystore (Redis)** for high-speed caching of mapping rules to speed up the sync engine.
- Leverage **Eventarc & Pub/Sub** for scalable, decoupled event routing between the API, webhooks, and Temporal workers.


## Frontend & UI Constraints (STRICT)
**CRITICAL RULE:** The overarching frontend design layout is **FIXED**. We will absolutely **NOT** alter the core layout structure.
*   **Strict Design Language:** We will only polish custom components using existing libraries to respect the current design language.
*   **Right-Sidebar Standard:** Every domain must utilize the standard right-sidebar drawer for contextual information.
*   **Visual Polish:** We will utilize approved visual effects (like background blur and glassmorphism on modals/sidebars) to ensure the application feels premium without breaking the underlying grid or routing structure.
